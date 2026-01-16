package backupComponent

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var maintenanceModeCommand = "docker compose -f %s exec -u www-data %s php occ maintenance:mode --%s"
var nextcloudDBDumpCommand = "docker compose -f %s exec %s pg_dump nextcloud -h localhost -U nextcloud"
var rsyncNextcloudDataDirCommand = "rsync -ax --delete --mkpath %s %s%s"
var rsyncNextcloudDBDumpCommand = "rsync -ax --delete --mkpath %s %s%s/postgres/nextcloud-sqlbkp.bak"

type NextcloudBackup struct {
	BaseComponentAttributes     `yaml:",inline"`
	ComposeSpecificationPath    string `yaml:"composeSpecificationPath"`
	ComposeNextcloudServiceName string `yaml:"composeNextcloudServiceName"`
	ComposeDatabaseServiceName  string `yaml:"composeDatabaseServiceName"`
	TmpBackupDir                string `yaml:"tmpBackupDir"`
	DockerMountDir              string `yaml:"dockerMountDir"`
	DstRelativePath             string `yaml:"dstRelativePath"`
}

func (ncBackup NextcloudBackup) PerformBackup(rsyncTargetHost string) error {
	var err error

	if err = ncBackup.checkPrerequisites(); err != nil {
		slog.Error("Prerequisite check failed", "error", err.Error())
		return err
	}

	if err = ncBackup.setNextcloudMaintenanceMode(true); err != nil {
		slog.Error("Can't freeze nextcloud activity before backup", "error", err.Error())
		return err
	}

	defer ncBackup.cleanupAfterError(err)

	if err = ncBackup.performDataDirBackup(rsyncTargetHost); err != nil {
		return fmt.Errorf("Error performing backup of nc data directory: %w", err)
	}

	dbDumpFilepath, err := ncBackup.createDbDump()
	if err != nil {
		return err
	}

	if err = ncBackup.setNextcloudMaintenanceMode(false); err != nil {
		slog.Error("Error disabling Nextcloud maintenance mode after creating backup", "error", err.Error())
		return err
	}

	err = transferDbDump(dbDumpFilepath, rsyncTargetHost, ncBackup.DstRelativePath)
	if err != nil {
		return err
	}

	if err = removeTmpDbDump(dbDumpFilepath); err != nil {
		return err
	}

	return nil
}

func transferDbDump(dbDumpFilepath string, rsyncTargetHost string, dstRelativePath string) error {
	dumpDstPath := filepath.Join(dstRelativePath, "postgres", "nextcloud-sqlbkp.bak")
	rsyncDBDumpAppliedCommand := fmt.Sprintf(rsyncNextcloudDBDumpCommand, dbDumpFilepath, rsyncTargetHost, dumpDstPath)
	rsyncDBDumpSplitCommand := strings.Split(rsyncDBDumpAppliedCommand, " ")

	slog.Info("Transferring Nextcloud DB dump", "source", dbDumpFilepath, "destination", dumpDstPath, "host", rsyncTargetHost)
	_, err := exec.Command(rsyncDBDumpSplitCommand[0], rsyncDBDumpSplitCommand[1:]...).Output()
	if err != nil {
		return appendErrStderr("Error transferring Nextcloud DB dump", err, "")
	}
	return nil
}

func removeTmpDbDump(dbDumpFilepath string) error {
	fileInfo, err := os.Stat(dbDumpFilepath)
	if err != nil {
		return fmt.Errorf("Error getting file info for Nextcloud DB dump: %w", err)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("Abort deleting tmp db dump; is not a file: %s", dbDumpFilepath)
	}
	err = os.Remove(dbDumpFilepath)
	if err != nil {
		return fmt.Errorf("Error removing Nextcloud DB dump file: %w", err)
	}
	slog.Debug("Removed temporary Nextcloud DB dump file", "file", dbDumpFilepath)
	return nil
}

func (ncBackup NextcloudBackup) checkPrerequisites() error {
	if err := ensureDirExists(ncBackup.DockerMountDir); err != nil {
		return fmt.Errorf("Prerequisite check failed: Docker mount dir: %w", err)
	}

	if existsErr := ensureDirExists(ncBackup.TmpBackupDir); existsErr != nil {
		if mkdirErr := os.MkdirAll(ncBackup.TmpBackupDir, 0700); mkdirErr != nil {
			slog.Error("Error creating tmp backup dir", "error", mkdirErr.Error())
			return mkdirErr
		}
		slog.Info("Created tmp backup dir", "tmpBackupDir", ncBackup.TmpBackupDir)
	} else {
		fileInfo , err := os.Stat(ncBackup.TmpBackupDir)
		if err != nil {
			return fmt.Errorf("Cannot obtain permission info for preexisting tmp backup dir: %w", err)
		}
		if fileInfo.Mode().Perm() != 0700 {
			err := os.Chmod(ncBackup.TmpBackupDir, 0700)
			if err != nil {
				return fmt.Errorf("Cannot set restrictive permissions for preexisting tmp backup dir: %w", err)
			}
		}
	}
	return nil
}

func (ncBackup NextcloudBackup) performDataDirBackup(rsyncTargetHost string) error {
	dataSrcPath := filepath.Join(ncBackup.DockerMountDir, "nextcloud/")
	dataDstPath := filepath.Join(ncBackup.DstRelativePath, "nextcloud/")
	slog.Info("Transferring Nextcloud data directory", "source", dataSrcPath, "destination", dataDstPath, "host", rsyncTargetHost)

	rsyncDataAppliedCommand := fmt.Sprintf(rsyncNextcloudDataDirCommand, dataSrcPath, rsyncTargetHost, dataDstPath)
	rsyncDataSplitCommand := strings.Split(rsyncDataAppliedCommand, " ")

	slog.Debug("Running command", "command", rsyncDataAppliedCommand)

	_, err := exec.Command(rsyncDataSplitCommand[0], rsyncDataSplitCommand[1:]...).Output()
	if err != nil {
		return appendErrStderr("Error transferring Nextcloud data directory", err, "")
	}
	return nil
}

func (ncBackup NextcloudBackup) createDbDump() (string, error) {
	date := time.Now().UTC().Format("2006-01-02")
	dbDumpFilename := fmt.Sprintf("nextcloud-sqlbkp-%s.bak", date)
	dbDumpFilepath := filepath.Join(ncBackup.TmpBackupDir, dbDumpFilename)

	appliedFullCommand := fmt.Sprintf(nextcloudDBDumpCommand, ncBackup.ComposeSpecificationPath, ncBackup.ComposeDatabaseServiceName)
	splitFullCommand := strings.Split(appliedFullCommand, " ")
	command := splitFullCommand[0]
	splitArgs := splitFullCommand[1:]

	cmd := exec.Command(command, splitArgs...)

	var err error
	fileDBDump, err := os.Create(dbDumpFilepath)
	if err != nil {
		return "", fmt.Errorf("Error creating Nextcloud DB dump file in tmp directory: %w", err)
	}
	defer fileDBDump.Close()

	var stderr bytes.Buffer
	

	cmd.Stdout = fileDBDump
	cmd.Stderr = &stderr
	slog.Info("Creating Nextcloud DB dump...", "dbDumpFilepath", dbDumpFilepath)

	slog.Debug("Running command", "command", appliedFullCommand)
	err = cmd.Run()
	if err != nil {
		return "", appendErrStderr("Error creating Nextcloud DB dump", err, stderr.String())
	}

	slog.Debug("Created Nextcloud DB dump", "dbDumpFilepath", dbDumpFilepath)
	return dbDumpFilepath, nil
}

func (ncBackup NextcloudBackup) setNextcloudMaintenanceMode(enable bool) error {
	var flag string

	if enable {
		flag = "on"
	} else {
		flag = "off"
	}

	fullAppliedCommand := fmt.Sprintf(maintenanceModeCommand, ncBackup.ComposeSpecificationPath, ncBackup.ComposeNextcloudServiceName, flag)
	fullSplitCommand := strings.Split(fullAppliedCommand, " ")

	slog.Info("Setting Nextcloud maintenance mode...", "flag", flag)

	slog.Debug("Running command", "command", fullAppliedCommand)

	cmd := exec.Command(fullSplitCommand[0], fullSplitCommand[1:]...)
	_, err := cmd.Output()
	if err != nil {
		return appendErrStderr(fmt.Sprintf("Error setting Nextcloud maintenance mode %s", flag), err, "")
	}

	return nil
}

func (ncBackup NextcloudBackup) cleanupAfterError(err error) {
	if err == nil {
		slog.Debug("Deferred cleanup: no error present, skipping...")
		return
	}

	slog.Error("Aborted backup due to error, deactivating maintenance mode", "error", err)
	if err := ncBackup.setNextcloudMaintenanceMode(false); err != nil {
		slog.Error("Error cleaning up residual maintenance mode", "error", err.Error())
		return
	}
}

func ensureDirExists(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("directory path is empty")
	}

	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", dirPath)
	}
	return nil
}

// Appends the `err` to `message`. If `err` is of type exec.ExitError and contains stderr output,
// the stderr output is also appended to the returned error.
// If stderr is specified, it is used instead of extracting it from err.
func appendErrStderr(message string, err error, stderr string) error {
	fallbackErr := fmt.Errorf("%s: %w", message, err)

	// check if error is of type exitError so we can access the stderr
	exitErr := &exec.ExitError{}
	if !errors.As(err, &exitErr) {
		return fallbackErr
	}
	
	var outputStderr string
	if stderr != "" {
		outputStderr = stderr
	} else {
		if exitErr.Stderr != nil {
			outputStderr = string(exitErr.Stderr)
		} else {
			return fallbackErr
		}
	}

	return fmt.Errorf("%s, %w, stderr:\n%s", message, err, outputStderr)
}

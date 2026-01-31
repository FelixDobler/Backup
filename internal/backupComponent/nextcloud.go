package backupComponent

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	// "time"
)

// Maintenance mode command, takes compose file path, nextcloud service name, and "on" or "off"
var maintenanceModeCommand = "docker compose -f %s exec -u www-data %s php occ maintenance:mode --%s"

// PostgreSQL dump command, takes compose file path and database service name
var nextcloudDBDumpCommand = "docker compose -f %s exec %s pg_dump nextcloud -h localhost -U nextcloud"

// Rsync data dir command, takes source path, rsync target host, and destination relative path
var rsyncNextcloudDataDirCommand = "rsync -ax --delete --mkpath %s %s%s"

// Rsync db command, takes source path, rsync target host, and destination relative path prefix
var rsyncNextcloudDBDumpCommand = "rsync -avx --mkpath %s %s%s"

// Destination path suffix for Nextcloud database dump on rsync target host
var rsyncDbDstPathComponent = "postgres/"

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
		return fmt.Errorf("Prerequisite check failed: %w", err)
	}

	if err = ncBackup.setNextcloudMaintenanceMode(true); err != nil {
		return fmt.Errorf("Can't freeze nextcloud activity before backup: %w", err)
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
		return fmt.Errorf("Error disabling Nextcloud maintenance mode after creating backup: %w", err)
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

func (ncBackup NextcloudBackup) performDataDirBackup(rsyncTargetHost string) error {
	dataSrcPath := filepath.Join(ncBackup.DockerMountDir, "nextcloud/")
	dataDstPath := filepath.Join(ncBackup.DstRelativePath, "nextcloud/")
	slog.Info("Transferring Nextcloud data directory", "source", dataSrcPath, "destination", dataDstPath, "host", rsyncTargetHost)

	rsyncDataAppliedCommand := fmt.Sprintf(rsyncNextcloudDataDirCommand, dataSrcPath, rsyncTargetHost, dataDstPath)
	rsyncDataSplitCommand := strings.Split(rsyncDataAppliedCommand, " ")

	slog.Debug("Running command", "command", rsyncDataAppliedCommand)

	_, err := exec.Command(rsyncDataSplitCommand[0], rsyncDataSplitCommand[1:]...).Output()
	if err != nil {
		return AppendExecErrStderr("Error transferring Nextcloud data directory", err, "")
	}
	return nil
}

func (ncBackup NextcloudBackup) createDbDump() (string, error) {
	// date := time.Now().UTC().Format("2006-01-02")
	// dbDumpFilename := fmt.Sprintf("nextcloud-sqlbkp-%s.bak", date)
	dbDumpFilename := "nextcloud-sqlbkp.bak"
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

	// Specify output file with pg_dump
	cmd.Stdout = fileDBDump
	cmd.Stderr = &stderr
	slog.Info("Creating Nextcloud DB dump...", "dbDumpFilepath", dbDumpFilepath)

	slog.Debug("Running command", "command", appliedFullCommand)
	err = cmd.Run()
	if err != nil {
		return "", AppendExecErrStderr("Error creating Nextcloud DB dump", err, stderr.String())
	}

	slog.Debug("Created Nextcloud DB dump", "dbDumpFilepath", dbDumpFilepath)
	return dbDumpFilepath, nil
}

func transferDbDump(dbDumpFilepath string, rsyncTargetHost string, dstRelativePath string) error {
	fullDstPath := filepath.Join(dstRelativePath, rsyncDbDstPathComponent) + "/"
	rsyncDBDumpAppliedCommand := fmt.Sprintf(rsyncNextcloudDBDumpCommand, dbDumpFilepath, rsyncTargetHost, fullDstPath)
	rsyncDBDumpSplitCommand := strings.Split(rsyncDBDumpAppliedCommand, " ")

	slog.Info("Transferring Nextcloud DB dump", "source", dbDumpFilepath, "destination", fullDstPath, "host", rsyncTargetHost)
	out, err := exec.Command(rsyncDBDumpSplitCommand[0], rsyncDBDumpSplitCommand[1:]...).Output()
	if err != nil {
		return AppendExecErrStderr("Error transferring Nextcloud DB dump", err, "")
	}
	slog.Debug("Rsync DB dump command output", "output", string(out))
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

	_, err := exec.Command(fullSplitCommand[0], fullSplitCommand[1:]...).Output()
	if err != nil {
		return AppendExecErrStderr(fmt.Sprintf("Error setting Nextcloud maintenance mode %s", flag), err, "")
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

func (ncBackup NextcloudBackup) checkPrerequisites() error {
	var requiredTmpDirPerm os.FileMode = 0700

	if err := ensureDirExists(ncBackup.DockerMountDir); err != nil {
		return fmt.Errorf("Docker mount dir doesn't exist: %w", err)
	}

	if existsErr := ensureDirExists(ncBackup.TmpBackupDir); existsErr != nil {
		if mkdirErr := os.MkdirAll(ncBackup.TmpBackupDir, requiredTmpDirPerm); mkdirErr != nil {
			return fmt.Errorf("Cannot create tmp backup dir: %w", mkdirErr)
		}
		slog.Info("Created tmp backup dir", "tmpBackupDir", ncBackup.TmpBackupDir)
	} else {
		entries, err := os.ReadDir(ncBackup.TmpBackupDir)
		if err != nil {
			return fmt.Errorf("Cannot read tmp backup dir to check if empty: %w", err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("Tmp backup dir is not empty: %s", ncBackup.TmpBackupDir)
		}

		err = ensureTmpDirPerms(ncBackup.TmpBackupDir, requiredTmpDirPerm)
		if err != nil {
			return err
		}
	}
	return nil
}
func ensureDirExists(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("Directory path is empty")
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
func ensureTmpDirPerms(tmpBackupDir string, requiredTmpDirPerm os.FileMode) error {
	fileInfo, err := os.Stat(tmpBackupDir)
	if err != nil {
		return fmt.Errorf("Cannot obtain permission info for preexisting tmp backup dir: %w", err)
	}
	if fileInfo.Mode().Perm() != requiredTmpDirPerm {
		slog.Warn("Setting restrictive permissions for preexisting tmp backup dir", "tmpBackupDir", tmpBackupDir, "oldPerm", fileInfo.Mode().Perm())
		err := os.Chmod(tmpBackupDir, requiredTmpDirPerm)
		if err != nil {
			return fmt.Errorf("Cannot set restrictive permissions for preexisting tmp backup dir: %w", err)
		}
	}
	return nil
}

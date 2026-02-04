package backupComponent

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
)

type OSBackup struct {
	BaseComponentAttributes `yaml:",inline"`
	Source                  string   `yaml:"source" validate:"required,file|dir"`
	DstRelativePath         string   `yaml:"dstRelativePath" validate:"required,filepath|dirpath"`
	UseDefaultFiles         bool     `yaml:"useDefaultFiles"`
	UseDefaultExcludes      bool     `yaml:"useDefaultExcludes"`
	FileListPath            string   `yaml:"fileListPath" validate:"omitempty,file"`
	ExcludeListPath         string   `yaml:"excludeListPath" validate:"omitempty,file"`
	AdditionalSources       []string `yaml:"additionalSources" validate:"omitempty,dive,file|dir"`
}

// Include default artifacts in the binary
//go:embed os_artifacts/*
var default_artifacts embed.FS

const dirRsyncDefaults = "/etc/fdbackup/defaults"

func (osBackup OSBackup) PerformBackup(rsyncTargetHost string) error {
	slog.Debug("---Performing OS backup---")
	err := osBackup.bootstrapDefaultArtifacts(dirRsyncDefaults)
	if err != nil {
		return fmt.Errorf("Error bootstrapping default artifacts: %w", err)
	}

	command := "rsync"
	completeArgs := osBackup.buildCommandArgs(rsyncTargetHost)
	slog.Debug("Executing rsync command", "command", command, "args", completeArgs)
	_, err = exec.Command(command, completeArgs...).Output()
	if err != nil {
		slog.Error("Error executing rsync command", "cmd", command, "args", completeArgs)
		return AppendExecErrStderr("Error executing rsync command", err, "")
	}
	// TODO stream output if desired

	return nil
}

func (osBackup OSBackup) bootstrapDefaultArtifacts(artifactsDir string) error {
	slog.Debug("bootstrapping default OS backup artifacts", "artifactsDir", artifactsDir)

	err := dirExists(artifactsDir)
	if err != nil {
		return fmt.Errorf("Default artifacts directory doesn't exist: %w", err)
	}

	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		return fmt.Errorf("failed to read default artifacts dir: %w", err)
	}
	if len(entries) > 0 {
		slog.Debug("Default artifacts directory is not empty, skipping bootstrap")
		for _, entry := range entries {
			slog.Debug("Existing artifact", "name", entry.Name())
		}
		return nil
	}

	defaultFiles, err := default_artifacts.ReadDir("os_artifacts")
	if err != nil {
		return fmt.Errorf("failed to read embedded artifacts: %w", err)
	}
	for _, file := range defaultFiles {
		slog.Debug("bootstrapping default artifact", "name", file.Name())
		data, err := default_artifacts.ReadFile(path.Join("os_artifacts", file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read embedded artifact %s: %w", file.Name(), err)
		}
		err = os.WriteFile(path.Join(artifactsDir, file.Name()), data, 0600)
		slog.Debug("Wrote default artifact", "dstPath", path.Join(artifactsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to write default artifact %s: %w", file.Name(), err)
		}
	}
	slog.Debug("Default OS backup artifacts bootstrapped successfully")
	return nil
}

func (osBackup OSBackup) buildCommandArgs(rsyncTargetHost string) []string {
	var args []string
	if osBackup.UseDefaultFiles {
		args = append(args, "--files-from="+path.Join(dirRsyncDefaults, "default_files.conf"))
	}
	if osBackup.UseDefaultExcludes {
		args = append(args, "--exclude-from="+path.Join(dirRsyncDefaults, "default_excludes.conf"))
	}
	if osBackup.FileListPath != "" {
		args = append(args, "--files-from="+osBackup.FileListPath)
	}
	if osBackup.ExcludeListPath != "" {
		args = append(args, "--exclude-from="+osBackup.ExcludeListPath)
	}

	/*
		Default rsync args for OS backup
		-avx: archive mode, verbose, one file system
		-r: recursive; needed for backing up directories listed in --files-from
		--mkpath: create destination path if it doesn't exist
		--delete: delete files in destination that are not in source
	*/
	args = append(args, "-avx", "-r", "--mkpath", "--delete")
	args = append(args, osBackup.Source)
	for _, additionalSource := range osBackup.AdditionalSources {
		args = append(args, additionalSource)
	}
	args = append(args, rsyncTargetHost + osBackup.DstRelativePath)
	return args
}

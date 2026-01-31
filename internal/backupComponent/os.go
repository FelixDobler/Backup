package backupComponent

import (
	"log/slog"
)

type OSBackup struct {
	BaseComponentAttributes `yaml:",inline"`
	TmpBackupDir            string   `yaml:"tmpBackupDir"`
	DockerMountDir          string   `yaml:"dockerMountDir"`
	Source                  string   `yaml:"source"`
	DstRelativePath         string   `yaml:"dstRelativePath"`
	UseDefaultFiles         bool     `yaml:"useDefaultFiles"`
	UseDefaultExcludes      bool     `yaml:"useDefaultExcludes"`
	FileListPath            string   `yaml:"fileListPath"`
	ExcludeListPath         string   `yaml:"excludeListPath"`
	AdditionalFiles         []string `yaml:"additionalFiles"`
}

func (osBackup OSBackup) PerformBackup(rsyncTargetHost string) error{
	slog.Debug("---Performing OS backup---\n")
	// TODO implement OS backup logic
	return nil
}

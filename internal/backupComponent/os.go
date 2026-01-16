package backupComponent

import "fmt"

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
	fmt.Printf("---Performing OS backup---\n")
	return nil
}

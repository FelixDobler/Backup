package backupComponent

import "fmt"

type RsyncBackup struct {
	BaseComponentAttributes `yaml:",inline"`
	Source                  string `yaml:"source"`
	DstRelativePath         string `yaml:"dstRelativePath"`
}

func (rsyncBackup RsyncBackup) PerformBackup(rsyncTargetHost string) error{
	fmt.Printf("---Performing Rsync backup---\n")
	fmt.Printf("Attributes:\n%+v\n", rsyncBackup)
	return nil
}

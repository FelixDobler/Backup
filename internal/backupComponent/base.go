package backupComponent

type BackupComponent interface {
	PerformBackup(rsyncTargetHost string) error
	GetName() string
	IsEnabled() bool
	GetType() string
	// RunCleanupOnFailure(err error) error
}

type BaseComponentAttributes struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`
}

func (bca BaseComponentAttributes) GetName() string {
	return bca.Name
}

func (bca BaseComponentAttributes) IsEnabled() bool {
	return bca.Enabled
}

func (bca BaseComponentAttributes) GetType() string {
	return bca.Type
}

type BackupType string

const (
	OSBackupType        BackupType = "os"
	NextcloudBackupType BackupType = "nextcloud"
	RsyncBackupType     BackupType = "rsync"
)

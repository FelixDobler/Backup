package config

import backupComponents "fedob/backup/internal/backupComponent"

type ClientConfig struct {
	RootLevelConfig  `yaml:",inline"`
	BackupComponents []backupComponents.BackupComponent `yaml:"components"`
}

type RootLevelConfig struct {
	LogDir          string `yaml:"logDir"`
	Email           string `yaml:"email"`
	RsyncTargetHost string `yaml:"rsyncTargetHost"`
}

// Registry mapping backup types to their constructors
var BackupComponentRegistry = map[backupComponents.BackupType]func() backupComponents.BackupComponent{
	backupComponents.OSBackupType:        func() backupComponents.BackupComponent { return &backupComponents.OSBackup{} },
	backupComponents.NextcloudBackupType: func() backupComponents.BackupComponent { return &backupComponents.NextcloudBackup{} },
	backupComponents.RsyncBackupType:     func() backupComponents.BackupComponent { return &backupComponents.RsyncBackup{} },
}

package config

import backupComponents "fedob/backup/internal/backupComponent"

type ConfigLoader interface {
	LoadConfig() ClientConfig
}

type ClientConfig struct {
	RootLevelConfig  `yaml:",inline"`
	BackupComponents []backupComponents.BackupComponent `yaml:"components" validate:"dive,required"`
}

type RootLevelConfig struct {
	LogDir          string     `yaml:"logDir"`
	// Email           string     `yaml:"email" validate:"required,email"`
	Mail            MailConfig `yaml:"mail"`
	RsyncTargetHost string     `yaml:"rsyncTargetHost" validate:"rsyncTargetHostValidation"`
}

type MailConfig struct {
	Host     string `yaml:"host" validate:"required,hostname"`
	Port     int    `yaml:"port" validate:"required,numeric"`
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	From     string `yaml:"from" validate:"required,email"`
	To       string `yaml:"to" validate:"required,email"`
}

// Registry mapping backup types to their constructors
var BackupComponentRegistry = map[backupComponents.BackupType]func() backupComponents.BackupComponent{
	backupComponents.OSBackupType:        func() backupComponents.BackupComponent { return &backupComponents.OSBackup{} },
	backupComponents.NextcloudBackupType: func() backupComponents.BackupComponent { return &backupComponents.NextcloudBackup{} },
	backupComponents.RsyncBackupType:     func() backupComponents.BackupComponent { return &backupComponents.RsyncBackup{} },
}

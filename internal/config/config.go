package config

// import "fedob/backup/internal/backupComponent"

type ConfigAdapter interface {
	LoadConfig() ClientConfig
	SaveConfig(ClientConfig) error
}

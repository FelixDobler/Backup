package main

import (
	"fedob/backup/internal/config/yaml"
	// "fmt"
	"log/slog"
)

func main() {
	var configConnector = yaml.NewYAMLConfigInterface("cmd/client/config_test.yaml")

	slog.Info("Loading client config", "configConnector", configConnector)
	clientConfig, err := configConnector.LoadConfig()
	if err != nil {
		slog.Error("Failed to load client config", "error", err)
		return
	}

	slog.Debug("Finished loading config")

	for _, component := range clientConfig.BackupComponents {
		if !component.IsEnabled() {
			slog.Info("Skipping disabled component", "componentName", component.GetName())
			continue
		}
		slog.Info("Starting backup for component",
			"component", component.GetName(), "type", component.GetType())
		err := component.PerformBackup(clientConfig.RsyncTargetHost)
		if err != nil {
			slog.Error("Error performing backup for component", "component", component.GetName(), "error", err.Error())
			// TODO: send alert
		} else {
			slog.Info("Successfully completed backup for component", "component", component.GetName(), "type", component.GetType())
		}
	}

	// configConnector.SaveConfig(clientConfig)
}

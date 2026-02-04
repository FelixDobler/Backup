package main

import (
	"fedob/backup/internal/config/yaml"
	"fmt"
	"os"

	"fedob/backup/internal/alert"
	"fedob/backup/internal/config"
	"fedob/backup/internal/logging"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

func main() {
	logging.Init()
	
	var configConnector = yaml.NewYAMLConfigInterface("/config.yaml")

	slog.Info("Loading client config", "configConnector", configConnector)
	clientConfig, err := configConnector.LoadConfig()
	if err != nil {
		slog.Error("Failed to load client config", "error", err)
		return
	}

	slog.Debug("Config", "clientConfig", clientConfig)

	validate := validator.New(validator.WithRequiredStructEnabled())
	config.RegisterCustomValidations(validate)

	if err := validate.Struct(clientConfig); err != nil {
		slog.Error("Config validation error", "error", err)
		panic("stop")
	}
	slog.Debug("YAML config validation succeeded")
	slog.Debug("Finished loading config")

	// FIXME remove injection of MAIL_PASSWORD here
	mailNotifier := clientConfig.Mail
	mailNotifier.Password = os.Getenv("MAIL_PASSWORD")
	notifier := alert.MailAlertNotifier(mailNotifier)

	var alertMessages []alert.AlertMessage
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
			alertMessages = append(alertMessages, alert.AlertMessage{
				Subject: fmt.Sprintf("Error performing backup for component %s", component.GetName()),
				Body:    err.Error(),
			})
		} else {
			slog.Info("Successfully completed backup for component", "component", component.GetName(), "type", component.GetType())
		}
	}
	if len(alertMessages) > 0 && os.Getenv("SUPPRESS_ALERTS") != "true" {
		slog.Info("Sending alert email for backup errors")
		err := notifier.SendAlert(alertMessages)
		if err != nil {
			slog.Error("Failed to send alert email", "error", err)
		}
	}
}

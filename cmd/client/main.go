package main

import (
	"fedob/backup/internal/config/yaml"
	"fmt"
	"os"
	"strings"

	// "fmt"
	"fedob/backup/internal/alert"
	"fedob/backup/internal/logging"
	"log/slog"

	// TODO REMOVE IN PROD
	// "github.com/joho/godotenv"
	"reflect"
)

func main() {
	// err := godotenv.Load(".env")
	// if err != nil {
		// panic("Error loading .env file")
	// }
	logging.Init()
	
	var configConnector = yaml.NewYAMLConfigInterface("cmd/client/config_test.yaml")

	slog.Info("Loading client config", "configConnector", configConnector)
	clientConfig, err := configConnector.LoadConfig()
	if err != nil {
		slog.Error("Failed to load client config", "error", err)
		return
	}

	slog.Debug("Finished loading config")

	mailNotifier := alert.MailAlertNotifier{
		Host: os.Getenv("MAIL_HOST"),
		Port: os.Getenv("MAIL_PORT"),
		Username: os.Getenv("MAIL_USERNAME"),
		Password: os.Getenv("MAIL_PASSWORD"),
		From: os.Getenv("MAIL_FROM"),
		To: clientConfig.Email,
	}
	// ensure the values of mailNotifier aren't empty
	fields := reflect.VisibleFields(reflect.TypeOf(mailNotifier))
	for _, field := range fields {
		value := reflect.ValueOf(mailNotifier).FieldByName(field.Name).String()
		if value == "" {
			slog.Error("Mail notifier field is empty", "field", field.Name)
			return
		}
	}

	// TODO collect individual error messages instead of combining them here
	var errorMsgBuffer strings.Builder
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
			errorMsgBuffer.WriteString(fmt.Sprintf("Error performing backup for component %s: %s\n\n", component.GetName(), err.Error()))
		} else {
			slog.Info("Successfully completed backup for component", "component", component.GetName(), "type", component.GetType())
		}
	}
	if errorMsgBuffer.Len() != 0 {
		slog.Info("Sending alert email for backup errors")
		err :=mailNotifier.SendAlert("Backup Error", errorMsgBuffer.String())
		if err != nil {
			slog.Error("Failed to send alert email", "error", err)
		}
	}
}

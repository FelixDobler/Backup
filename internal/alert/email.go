package alert

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type MailAlertNotifier struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       string
}

var emailBodyTemplate = "From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s"

func (m MailAlertNotifier) SendAlert(messages []AlertMessage) error {
	var combinedSubject strings.Builder
	for i, msg := range messages {
		combinedSubject.WriteString(msg.Subject)
		if i < len(messages)-1 {
			combinedSubject.WriteString(" | ")
		}
	}

	var combinedBody strings.Builder
	for _, msg := range messages {
		combinedBody.WriteString(fmt.Sprintf("---%s---\n", msg.Subject))
		indentedBody := strings.ReplaceAll(msg.Body, "\n", "\n    ")
		combinedBody.WriteString(indentedBody + "\n\n")
	}
	auth := sasl.NewPlainClient("", m.Username, m.Password)
	mailBody := fmt.Sprintf(emailBodyTemplate, m.From, m.To, combinedSubject.String(), combinedBody.String())

	err := smtp.SendMailTLS(fmt.Sprintf("%s:%d", m.Host, m.Port), auth, m.From, []string{m.To}, strings.NewReader(mailBody))
	if err != nil {
		return fmt.Errorf("failed to send alert email: %w", err)
	}
	slog.Info("Alert email sent", "to", m.To, "subject", combinedSubject.String(), "msg", combinedBody.String())
	return nil
}

func (m MailAlertNotifier) SendTestAlert() error {
	testMessage := AlertMessage{
		Subject: "Test Alert",
		Body:    "This is a test alert email.",
	}
	err := m.SendAlert([]AlertMessage{testMessage})
	if err != nil {
		return fmt.Errorf("Failed to send test alert email: %w", err)
	}
	slog.Info("Test alert email sent")
	return nil
}

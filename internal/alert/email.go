package alert

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

// TODO integrate into global config

type MailAlertNotifier struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	To       string
}

var emailBodyTemplate = "From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s"

func (m MailAlertNotifier) SendAlert(subject string, message string) error {
	auth := sasl.NewPlainClient("", m.Username, m.Password)
	mailBody := fmt.Sprintf(emailBodyTemplate, m.From, m.To, subject, message)

	err := smtp.SendMailTLS(m.Host+":"+m.Port, auth, m.From, []string{m.To}, strings.NewReader(mailBody))
	if err != nil {
		return fmt.Errorf("failed to send alert email: %w", err)
	}
	slog.Info("Alert email sent", "to", m.To, "subject", subject, "msg", message)
	return nil
}

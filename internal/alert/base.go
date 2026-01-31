package alert

type AlertNotifier interface {
	SendAlert(subject string, message string) error
	// TODO SendTestAlert() error
}

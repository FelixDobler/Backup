package alert

type AlertMessage struct {
	Subject string
	Body    string
}

type AlertNotifier interface {
	SendAlert(messages []AlertMessage) error
	SendTestAlert() error
}

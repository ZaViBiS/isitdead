package notification

type Notifier interface {
	Send(text string) error
}

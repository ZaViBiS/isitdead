package notification

type TelegramNorifier struct{}

func (t *TelegramNorifier) Send(text string) error {
	return nil
}

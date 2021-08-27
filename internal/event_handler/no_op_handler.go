package event_handler

type NoOpHandler struct {
}

func (eh NoOpHandler) HandleEvent() error {
	return nil
}

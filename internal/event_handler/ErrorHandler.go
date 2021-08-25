package event_handler

type ErrorHandler struct {
	err error
}

func (eh ErrorHandler) HandleEvent() error {
	return eh.err
}

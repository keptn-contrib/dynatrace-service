package event_handler

import "context"

type NoOpHandler struct {
}

// HandleEvent handles an event by doing nothing.
func (eh NoOpHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	return nil
}

package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type NoOpHandler struct {
	event cloudevents.Event
}

func (eh NoOpHandler) HandleEvent() error {
	return nil
}

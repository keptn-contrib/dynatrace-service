package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
)

type LoggingNoOpHandler struct {
	event cloudevents.Event
}

func (eh LoggingNoOpHandler) HandleEvent() error {
	log.WithField("EventType", eh.event.Type()).Info("Ignoring event")

	return nil
}

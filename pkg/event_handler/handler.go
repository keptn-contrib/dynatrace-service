package event_handler

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

type DynatraceEventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event, logger *keptnutils.Logger) (DynatraceEventHandler, error) {
	logger.Debug("Received event: " + event.Type())
	switch event.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return &ConfigureMonitoringEventHandler{Logger: logger, Event: event}, nil
	default:
		return &CDEventHandler{Logger: logger, Event: event}, nil
	}
}

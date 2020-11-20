package event_handler

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type DynatraceEventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event, logger *keptn.Logger) (DynatraceEventHandler, error) {
	logger.Debug("Received event: " + event.Type())
	switch event.Type() {
	case keptn.ConfigureMonitoringEventType:
		return &ConfigureMonitoringEventHandler{Logger: logger, Event: event}, nil
	case keptn.InternalProjectCreateEventType:
		return &CreateProjectEventHandler{Logger: logger, Event: event}, nil
	case keptn.ProblemEventType:
		return &ProblemEventHandler{Logger: logger, Event: event}, nil
	case keptn.ActionTriggeredEventType:
		return &ActionHandler{Logger: logger, Event: event}, nil
	case keptn.ActionStartedEventType:
		return &ActionHandler{Logger: logger, Event: event}, nil
	case keptn.ActionFinishedEventType:
		return &ActionHandler{Logger: logger, Event: event}, nil
	default:
		return &CDEventHandler{Logger: logger, Event: event}, nil
	}
}

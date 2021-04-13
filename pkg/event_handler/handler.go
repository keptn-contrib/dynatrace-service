package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type DynatraceEventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event, logger *keptn.Logger) (DynatraceEventHandler, error) {
	logger.Debug("Received event: " + event.Type())
	dtConfigGetter := &adapter.DynatraceConfigGetter{}
	switch event.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return &ConfigureMonitoringEventHandler{Logger: logger, Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName):
		return &CreateProjectEventHandler{Logger: logger, Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnevents.ProblemEventType:
		return &ProblemEventHandler{Logger: logger, Event: event}, nil
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		return &ActionHandler{Logger: logger, Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName):
		return &ActionHandler{Logger: logger, Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName):
		return &ActionHandler{Logger: logger, Event: event, dtConfigGetter: dtConfigGetter}, nil
	default:
		return &CDEventHandler{Logger: logger, Event: event, dtConfigGetter: dtConfigGetter}, nil
	}
}

package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DynatraceEventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event) (DynatraceEventHandler, error) {
	log.WithField("eventType", event.Type()).Debug("Received event")
	dtConfigGetter := &adapter.DynatraceConfigGetter{}
	switch event.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return &ConfigureMonitoringEventHandler{Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName):
		return &CreateProjectEventHandler{Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnevents.ProblemEventType:
		return &ProblemEventHandler{Event: event}, nil
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		return &ActionHandler{Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName):
		return &ActionHandler{Event: event, dtConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName):
		return &ActionHandler{Event: event, dtConfigGetter: dtConfigGetter}, nil
	default:
		return &CDEventHandler{Event: event, dtConfigGetter: dtConfigGetter}, nil
	}
}

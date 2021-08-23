package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"
	"github.com/keptn-contrib/dynatrace-service/internal/problem"
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
		return monitoring.NewConfigureMonitoringEventHandler(event, dtConfigGetter), nil
	case keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName):
		return monitoring.NewCreateProjectEventHandler(event, dtConfigGetter), nil
	case keptnevents.ProblemEventType:
		return &problem.ProblemEventHandler{Event: event}, nil
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		return &problem.ActionHandler{Event: event, DTConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName):
		return &problem.ActionHandler{Event: event, DTConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName):
		return &problem.ActionHandler{Event: event, DTConfigGetter: dtConfigGetter}, nil
	case keptnv2.GetTriggeredEventType(keptnv2.GetSLITaskName):
		return &GetSLIEventHandler{event: event, dtConfigGetter: dtConfigGetter}, nil
	default:
		return &CDEventHandler{Event: event, dtConfigGetter: dtConfigGetter}, nil
	}
}

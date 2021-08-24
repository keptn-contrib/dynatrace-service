package deployment

import (
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type TestFinishedEventHandler struct {
	event  *TestFinishedAdapter
	client *dynatrace.Client
	config *config.DynatraceConfigFile
}

// NewTestFinishedEventHandler creates a new TestFinishedEventHandler
func NewTestFinishedEventHandler(event *TestFinishedAdapter, client *dynatrace.Client, config *config.DynatraceConfigFile) *TestFinishedEventHandler {
	return &TestFinishedEventHandler{
		event:  event,
		client: client,
		config: config,
	}
}

// Handle handles an action finished event
func (eh *TestFinishedEventHandler) Handle() error {
	// Send Annotation Event
	ae := event.CreateAnnotationEvent(eh.event, eh.config)
	if ae.AnnotationType == "" {
		ae.AnnotationType = "Stop Tests"
	}
	if ae.AnnotationDescription == "" {
		ae.AnnotationDescription = "Stop running tests: against " + eh.event.GetService()
	}

	dynatrace.NewEventsClient(eh.client).SendEvent(ae)

	return nil
}

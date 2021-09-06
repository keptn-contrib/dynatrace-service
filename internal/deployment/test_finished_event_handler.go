package deployment

import (
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type TestFinishedEventHandler struct {
	event       TestFinishedAdapterInterface
	client      dynatrace.ClientInterface
	attachRules *config.DtAttachRules
}

// NewTestFinishedEventHandler creates a new TestFinishedEventHandler
func NewTestFinishedEventHandler(event TestFinishedAdapterInterface, client dynatrace.ClientInterface, attachRules *config.DtAttachRules) *TestFinishedEventHandler {
	return &TestFinishedEventHandler{
		event:       event,
		client:      client,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event
func (eh *TestFinishedEventHandler) HandleEvent() error {
	// Send Annotation Event
	ae := event.CreateAnnotationEvent(eh.event, eh.attachRules)
	if ae.AnnotationType == "" {
		ae.AnnotationType = "Stop Tests"
	}
	if ae.AnnotationDescription == "" {
		ae.AnnotationDescription = "Stop running tests: against " + eh.event.GetService()
	}

	dynatrace.NewEventsClient(eh.client).SendEvent(ae)

	return nil
}

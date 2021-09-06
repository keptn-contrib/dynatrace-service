package deployment

import (
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type TestTriggeredEventHandler struct {
	event       TestTriggeredAdapterInterface
	client      dynatrace.ClientInterface
	attachRules *config.DtAttachRules
}

// NewTestTriggeredEventHandler creates a new TestTriggeredEventHandler
func NewTestTriggeredEventHandler(event TestTriggeredAdapterInterface, client dynatrace.ClientInterface, attachRules *config.DtAttachRules) *TestTriggeredEventHandler {
	return &TestTriggeredEventHandler{
		event:       event,
		client:      client,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event
func (eh *TestTriggeredEventHandler) HandleEvent() error {
	// Send Annotation Event
	ie := event.CreateAnnotationEvent(eh.event, eh.attachRules)
	if ie.AnnotationType == "" {
		ie.AnnotationType = "Start Tests: " + eh.event.GetTestStrategy()
	}
	if ie.AnnotationDescription == "" {
		ie.AnnotationDescription = "Start running tests: " + eh.event.GetTestStrategy() + " against " + eh.event.GetService()
	}

	dynatrace.NewEventsClient(eh.client).SendEvent(ie)

	return nil
}

package deployment

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type TestTriggeredEventHandler struct {
	event       TestTriggeredAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewTestTriggeredEventHandler creates a new TestTriggeredEventHandler
func NewTestTriggeredEventHandler(event TestTriggeredAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *TestTriggeredEventHandler {
	return &TestTriggeredEventHandler{
		event:       event,
		dtClient:    dtClient,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event.
func (eh *TestTriggeredEventHandler) HandleEvent(ctx context.Context) error {

	imageAndTag := eh.eClient.GetImageAndTag(eh.event)

	// Send Annotation Event
	ie := dynatrace.CreateAnnotationEventDTO(eh.event, imageAndTag, eh.attachRules)
	if ie.AnnotationType == "" {
		ie.AnnotationType = "Start Tests: " + eh.event.GetTestStrategy()
	}
	if ie.AnnotationDescription == "" {
		ie.AnnotationDescription = "Start running tests: " + eh.event.GetTestStrategy() + " against " + eh.event.GetService()
	}

	dynatrace.NewEventsClient(eh.dtClient).AddAnnotationEvent(ctx, ie)

	return nil
}

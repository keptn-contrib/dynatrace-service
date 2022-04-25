package action

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
	annotationEvent := dynatrace.AnnotationEvent{
		EventType:             dynatrace.AnnotationEventType,
		Source:                eventSource,
		AnnotationType:        getValueFromLabels(eh.event, "type", "Start Tests: "+eh.event.GetTestStrategy()),
		AnnotationDescription: getValueFromLabels(eh.event, "description", "Start running tests: "+eh.event.GetTestStrategy()+" against "+eh.event.GetService()),
		CustomProperties:      createCustomProperties(eh.event, eh.eClient.GetImageAndTag(eh.event)),
		AttachRules:           *eh.attachRules,
	}

	dynatrace.NewEventsClient(eh.dtClient).AddAnnotationEvent(ctx, annotationEvent)
	return nil
}

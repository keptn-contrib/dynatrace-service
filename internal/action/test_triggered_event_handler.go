package action

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

// TestTriggeredEventHandler handles a test triggered event.
type TestTriggeredEventHandler struct {
	event       TestTriggeredAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewTestTriggeredEventHandler creates a new TestTriggeredEventHandler.
func NewTestTriggeredEventHandler(event TestTriggeredAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *TestTriggeredEventHandler {
	return &TestTriggeredEventHandler{
		event:       event,
		dtClient:    dtClient,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles a test triggered event.
func (eh *TestTriggeredEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	if eh.attachRules == nil {
		eh.attachRules = createDefaultAttachRules(eh.event)
	}

	annotationEvent := dynatrace.AnnotationEvent{
		EventType:             dynatrace.AnnotationEventType,
		Source:                eventSource,
		AnnotationType:        getValueFromLabels(eh.event, "type", "Start Tests: "+eh.event.GetTestStrategy()),
		AnnotationDescription: getValueFromLabels(eh.event, "description", "Start running tests: "+eh.event.GetTestStrategy()+" against "+eh.event.GetService()),
		CustomProperties:      newCustomProperties(eh.event, eh.eClient.GetImageAndTag(workCtx, eh.event), keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event)),
		AttachRules:           *eh.attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddAnnotationEvent(workCtx, annotationEvent)
}

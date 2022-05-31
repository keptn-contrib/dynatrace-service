package action

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type TestFinishedEventHandler struct {
	event       TestFinishedAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewTestFinishedEventHandler creates a new TestFinishedEventHandler
func NewTestFinishedEventHandler(event TestFinishedAdapterInterface, client dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *TestFinishedEventHandler {
	return &TestFinishedEventHandler{
		event:       event,
		dtClient:    client,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event.
func (eh *TestFinishedEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	if eh.attachRules == nil {
		eh.attachRules = createDefaultAttachRules(eh.event)
	}

	annotationEvent := dynatrace.AnnotationEvent{
		EventType:             dynatrace.AnnotationEventType,
		Source:                eventSource,
		AnnotationType:        getValueFromLabels(eh.event, "type", "Stop Tests"),
		AnnotationDescription: getValueFromLabels(eh.event, "description", "Stop running tests: against "+eh.event.GetService()),
		CustomProperties:      createCustomProperties(eh.event, eh.eClient.GetImageAndTag(eh.event), keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event)),
		AttachRules:           *eh.attachRules,
	}

	dynatrace.NewEventsClient(eh.dtClient).AddAnnotationEvent(workCtx, annotationEvent)
	return nil
}

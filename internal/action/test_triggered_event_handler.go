package action

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

// TestTriggeredEventHandler handles a test triggered event.
type TestTriggeredEventHandler struct {
	event            TestTriggeredAdapterInterface
	dtClient         dynatrace.ClientInterface
	eClient          keptn.EventClientInterface
	bridgeURLCreator keptn.BridgeURLCreatorInterface
	attachRules      *dynatrace.AttachRules
}

// NewTestTriggeredEventHandler creates a new TestTriggeredEventHandler.
func NewTestTriggeredEventHandler(event TestTriggeredAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, bridgeURLCreator keptn.BridgeURLCreatorInterface, attachRules *dynatrace.AttachRules) *TestTriggeredEventHandler {
	return &TestTriggeredEventHandler{
		event:            event,
		dtClient:         dtClient,
		eClient:          eClient,
		bridgeURLCreator: bridgeURLCreator,
		attachRules:      attachRules,
	}
}

// HandleEvent handles a test triggered event.
func (eh *TestTriggeredEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	imageAndTag := eh.eClient.GetImageAndTag(workCtx, eh.event)
	attachRules := createAttachRulesForDeploymentTimeFrame(workCtx, eh.dtClient, eh.eClient, eh.event, imageAndTag, eh.attachRules)

	annotationEvent := dynatrace.AnnotationEvent{
		EventType:             dynatrace.AnnotationEventType,
		Source:                eventSource,
		AnnotationType:        getValueFromLabels(eh.event, "type", "Start Tests: "+eh.event.GetTestStrategy()),
		AnnotationDescription: getValueFromLabels(eh.event, "description", "Start running tests: "+eh.event.GetTestStrategy()+" against "+eh.event.GetService()),
		CustomProperties:      newCustomProperties(eh.event, imageAndTag, eh.bridgeURLCreator.TryGetBridgeURLForKeptnContext(workCtx, eh.event)),
		AttachRules:           attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddAnnotationEvent(workCtx, annotationEvent)
}

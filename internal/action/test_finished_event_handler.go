package action

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type TestFinishedEventHandler struct {
	event            TestFinishedAdapterInterface
	dtClient         dynatrace.ClientInterface
	eClient          keptn.EventClientInterface
	bridgeURLCreator keptn.BridgeURLCreatorInterface
	attachRules      *dynatrace.AttachRules
}

// NewTestFinishedEventHandler creates a new TestFinishedEventHandler
func NewTestFinishedEventHandler(event TestFinishedAdapterInterface, client dynatrace.ClientInterface, eClient keptn.EventClientInterface, bridgeURLCreator keptn.BridgeURLCreatorInterface, attachRules *dynatrace.AttachRules) *TestFinishedEventHandler {
	return &TestFinishedEventHandler{
		event:            event,
		dtClient:         client,
		eClient:          eClient,
		bridgeURLCreator: bridgeURLCreator,
		attachRules:      attachRules,
	}
}

// HandleEvent handles an action finished event.
func (eh *TestFinishedEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	imageAndTag := eh.eClient.GetImageAndTag(workCtx, eh.event)
	attachRules := createAttachRulesForDeploymentTimeFrame(workCtx, eh.dtClient, eh.eClient, eh.event, imageAndTag, eh.attachRules)

	annotationEvent := dynatrace.AnnotationEvent{
		EventType:             dynatrace.AnnotationEventType,
		Source:                eventSource,
		AnnotationType:        getValueFromLabels(eh.event, "type", "Stop Tests"),
		AnnotationDescription: getValueFromLabels(eh.event, "description", "Stop running tests: against "+eh.event.GetService()),
		CustomProperties:      newCustomProperties(eh.event, imageAndTag, eh.bridgeURLCreator.TryGetBridgeURLForKeptnContext(workCtx, eh.event)),
		AttachRules:           attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddAnnotationEvent(workCtx, annotationEvent)
}

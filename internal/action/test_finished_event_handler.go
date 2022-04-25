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
func (eh *TestFinishedEventHandler) HandleEvent(ctx context.Context) error {

	imageAndTag := eh.eClient.GetImageAndTag(eh.event)

	ae := dynatrace.CreateAnnotationEventDTO(eh.event, imageAndTag, eh.attachRules)
	if ae.AnnotationType == "" {
		ae.AnnotationType = "Stop Tests"
	}
	if ae.AnnotationDescription == "" {
		ae.AnnotationDescription = "Stop running tests: against " + eh.event.GetService()
	}

	dynatrace.NewEventsClient(eh.dtClient).AddAnnotationEvent(ctx, ae)

	return nil
}

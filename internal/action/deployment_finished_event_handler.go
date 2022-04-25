package action

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type DeploymentFinishedEventHandler struct {
	event       DeploymentFinishedAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewDeploymentFinishedEventHandler creates a new DeploymentFinishedEventHandler
func NewDeploymentFinishedEventHandler(event DeploymentFinishedAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *DeploymentFinishedEventHandler {
	return &DeploymentFinishedEventHandler{
		event:       event,
		dtClient:    dtClient,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event.
func (eh *DeploymentFinishedEventHandler) HandleEvent(ctx context.Context) error {
	imageAndTag := eh.eClient.GetImageAndTag(eh.event)
	customProperties := createCustomProperties(eh.event, imageAndTag)
	de := createDeploymentEventDTO(eh.event, imageAndTag, customProperties, eh.attachRules)
	dynatrace.NewEventsClient(eh.dtClient).AddDeploymentEvent(ctx, de)
	return nil
}
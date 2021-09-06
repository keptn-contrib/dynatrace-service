package deployment

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type DeploymentFinishedEventHandler struct {
	event       DeploymentFinishedAdapterInterface
	client      dynatrace.ClientInterface
	attachRules *dynatrace.DtAttachRules
}

// NewDeploymentFinishedEventHandler creates a new DeploymentFinishedEventHandler
func NewDeploymentFinishedEventHandler(event DeploymentFinishedAdapterInterface, client dynatrace.ClientInterface, attachRules *dynatrace.DtAttachRules) *DeploymentFinishedEventHandler {
	return &DeploymentFinishedEventHandler{
		event:       event,
		client:      client,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event
func (eh *DeploymentFinishedEventHandler) HandleEvent() error {
	// send Deployment Event
	de := dynatrace.CreateDeploymentEventDTO(eh.event, eh.attachRules)

	dynatrace.NewEventsClient(eh.client).AddDeploymentEvent(de)

	return nil
}

package deployment

import (
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type DeploymentFinishedEventHandler struct {
	event       DeploymentFinishedAdapterInterface
	client      dynatrace.ClientInterface
	attachRules *config.DtAttachRules
}

// NewDeploymentFinishedEventHandler creates a new DeploymentFinishedEventHandler
func NewDeploymentFinishedEventHandler(event DeploymentFinishedAdapterInterface, client dynatrace.ClientInterface, attachRules *config.DtAttachRules) *DeploymentFinishedEventHandler {
	return &DeploymentFinishedEventHandler{
		event:       event,
		client:      client,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event
func (eh *DeploymentFinishedEventHandler) HandleEvent() error {
	// send Deployment Event
	de := event.CreateDeploymentEvent(eh.event, eh.attachRules)

	dynatrace.NewEventsClient(eh.client).SendEvent(de)

	return nil
}

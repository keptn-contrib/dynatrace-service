package deployment

import (
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type DeploymentFinishedEventHandler struct {
	event  *DeploymentFinishedAdapter
	client *dynatrace.Client
	config *config.DynatraceConfigFile
}

// NewDeploymentFinishedEventHandler creates a new DeploymentFinishedEventHandler
func NewDeploymentFinishedEventHandler(event *DeploymentFinishedAdapter, client *dynatrace.Client, config *config.DynatraceConfigFile) *DeploymentFinishedEventHandler {
	return &DeploymentFinishedEventHandler{
		event:  event,
		client: client,
		config: config,
	}
}

// HandleEvent handles an action finished event
func (eh *DeploymentFinishedEventHandler) HandleEvent() error {
	// send Deployment Event
	de := event.CreateDeploymentEvent(eh.event, eh.config)

	dynatrace.NewEventsClient(eh.client).SendEvent(de)

	return nil
}

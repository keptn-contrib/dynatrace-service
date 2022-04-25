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

	deploymentEvent := dynatrace.DeploymentEvent{
		EventType:         dynatrace.DeploymentEventType,
		Source:            eventSource,
		DeploymentName:    getValueFromLabels(eh.event, "deploymentName", "Deploy "+eh.event.GetService()+" "+imageAndTag.Tag()+" with strategy "+eh.event.GetDeploymentStrategy()),
		DeploymentProject: getValueFromLabels(eh.event, "deploymentProject", eh.event.GetProject()),
		DeploymentVersion: getValueFromLabels(eh.event, "deploymentVersion", imageAndTag.Tag()),
		CiBackLink:        getValueFromLabels(eh.event, "ciBackLink", ""),
		RemediationAction: getValueFromLabels(eh.event, "remediationAction", ""),
		CustomProperties:  createCustomProperties(eh.event, imageAndTag),
		AttachRules:       *eh.attachRules,
	}

	dynatrace.NewEventsClient(eh.dtClient).AddDeploymentEvent(ctx, deploymentEvent)
	return nil
}

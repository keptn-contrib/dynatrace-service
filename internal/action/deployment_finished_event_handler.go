package action

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

// DeploymentFinishedEventHandler handles a deployment finished event.
type DeploymentFinishedEventHandler struct {
	event       DeploymentFinishedAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewDeploymentFinishedEventHandler creates a new DeploymentFinishedEventHandler.
func NewDeploymentFinishedEventHandler(event DeploymentFinishedAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *DeploymentFinishedEventHandler {
	return &DeploymentFinishedEventHandler{
		event:       event,
		dtClient:    dtClient,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles a deployment finished event.
func (eh *DeploymentFinishedEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	imageAndTag := eh.eClient.GetImageAndTag(workCtx, eh.event)
	attachRules, err := createAttachRules(workCtx, eh.dtClient, eh.eClient, eh.event, imageAndTag, eh.attachRules)
	if err != nil {
		return fmt.Errorf("could not setup correct attach rules: %w", err)
	}

	deploymentEvent := dynatrace.DeploymentEvent{
		EventType:         dynatrace.DeploymentEventType,
		Source:            eventSource,
		DeploymentName:    getValueFromLabels(eh.event, "deploymentName", "Deploy "+eh.event.GetService()+" "+imageAndTag.Tag()+" with strategy "+eh.event.GetDeploymentStrategy()),
		DeploymentProject: getValueFromLabels(eh.event, "deploymentProject", eh.event.GetProject()),
		DeploymentVersion: getValueFromLabels(eh.event, "deploymentVersion", imageAndTag.Tag()),
		CiBackLink:        getValueFromLabels(eh.event, "ciBackLink", ""),
		RemediationAction: getValueFromLabels(eh.event, "remediationAction", ""),
		CustomProperties:  newCustomProperties(eh.event, imageAndTag, keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event)),
		AttachRules:       attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddDeploymentEvent(workCtx, deploymentEvent)
}

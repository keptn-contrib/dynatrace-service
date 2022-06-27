package action

import (
	"context"

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

	if eh.attachRules == nil {
		eh.attachRules = createDefaultAttachRules(eh.event)
	}

	deploymentEvent := dynatrace.DeploymentEvent{
		EventType:         dynatrace.DeploymentEventType,
		Source:            eventSource,
		DeploymentName:    getValueFromLabels(eh.event, "deploymentName", "Deploy "+eh.event.GetService()+" "+imageAndTag.Tag()+" with strategy "+eh.event.GetDeploymentStrategy()),
		DeploymentProject: getValueFromLabels(eh.event, "deploymentProject", eh.event.GetProject()),
		DeploymentVersion: getValueFromLabels(eh.event, "deploymentVersion", imageAndTag.Tag()),
		CiBackLink:        getValueFromLabels(eh.event, "ciBackLink", ""),
		RemediationAction: getValueFromLabels(eh.event, "remediationAction", ""),
		CustomProperties:  NewCustomProperties(eh.event, imageAndTag, keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event)),
		AttachRules:       *eh.attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddDeploymentEvent(workCtx, deploymentEvent)
}

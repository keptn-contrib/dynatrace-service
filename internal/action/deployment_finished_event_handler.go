package action

import (
	"context"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
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
	attachRules := eh.createAttachRules(workCtx, imageAndTag)

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

func (eh *DeploymentFinishedEventHandler) createAttachRules(ctx context.Context, imageAndTag common.ImageAndTag) dynatrace.AttachRules {
	eventTime := eh.event.GetTime()
	if eventTime == (time.Time{}) {
		// TODO 2022-07-05: there is a bug in .ToCloudEvent() method - no time is set there. This should be fixed with Keptn 0.18.0 release
		eventTime = time.Now().UTC()
	}

	deploymentStartedTime, err := eh.eClient.GetEventTimeStampForType(ctx, eh.event, keptnv2.GetStartedEventType(keptnv2.DeploymentTaskName))
	if err != nil {
		log.WithError(err).Warn("Could not find the corresponding deployment.started event")

		// set the start time to 3 secs before event time - at least we can try to find sth.
		startTime := eventTime.Add(-3 * time.Second)
		deploymentStartedTime = &startTime
	}

	// ignoring the error here, because it should not be possible to create an invalid timeframe here
	timeframe, _ := common.NewTimeframe(*deploymentStartedTime, eventTime)
	return createOrUpdateAttachRules(ctx, eh.dtClient, eh.attachRules, imageAndTag, eh.event, timeframe)
}

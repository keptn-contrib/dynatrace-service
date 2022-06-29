package action

import (
	"context"
	"fmt"

	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type ReleaseTriggeredEventHandler struct {
	event       ReleaseTriggeredAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewReleaseTriggeredEventHandler creates a new ReleaseTriggeredEventHandler
func NewReleaseTriggeredEventHandler(event ReleaseTriggeredAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *ReleaseTriggeredEventHandler {
	return &ReleaseTriggeredEventHandler{
		event:       event,
		dtClient:    dtClient,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event.
func (eh *ReleaseTriggeredEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	strategy, err := keptnevents.GetDeploymentStrategy(eh.event.GetDeploymentStrategy())
	if err != nil {
		log.WithError(err).Error("Could not determine deployment strategy")
		return err
	}

	if eh.attachRules == nil {
		eh.attachRules = createDefaultAttachRules(eh.event)
	}

	infoEvent := dynatrace.InfoEvent{
		EventType:        dynatrace.InfoEventType,
		Source:           eventSource,
		Title:            eh.getTitle(strategy, eh.event.GetLabels()["title"]),
		Description:      eh.getTitle(strategy, eh.event.GetLabels()["description"]),
		CustomProperties: createCustomProperties(eh.event, eh.eClient.GetImageAndTag(workCtx, eh.event), keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event)),
		AttachRules:      *eh.attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(workCtx, infoEvent)
}

func (eh *ReleaseTriggeredEventHandler) getTitle(strategy keptnevents.DeploymentStrategy, defaultValue string) string {
	if strategy == keptnevents.Direct && eh.event.GetResult() == keptnv2.ResultPass || eh.event.GetResult() == keptnv2.ResultWarning {
		return fmt.Sprintf("PROMOTING from %s to next stage", eh.event.GetStage())
	}

	if eh.event.GetResult() == keptnv2.ResultFailed {
		if strategy == keptnevents.Duplicate {
			return "Rollback Artifact (Switch Blue/Green) in " + eh.event.GetStage()
		}

		return fmt.Sprintf("NOT PROMOTING from %s to next stage", eh.event.GetStage())
	}

	if defaultValue == "" {
		return fmt.Sprintf("Release triggered for %s, %s and %s", eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
	}

	return defaultValue
}

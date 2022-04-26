package action

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
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
func (eh *ReleaseTriggeredEventHandler) HandleEvent(ctx context.Context) error {
	strategy, err := keptnevents.GetDeploymentStrategy(eh.event.GetDeploymentStrategy())
	if err != nil {
		log.WithError(err).Error("Could not determine deployment strategy")
		return err
	}

	infoEvent := dynatrace.InfoEvent{
		EventType:        dynatrace.InfoEventType,
		Source:           eventSource,
		Title:            eh.getTitle(strategy, eh.event.GetLabels()["title"]),
		Description:      eh.getTitle(strategy, eh.event.GetLabels()["description"]),
		CustomProperties: createCustomProperties(eh.event, eh.eClient.GetImageAndTag(eh.event), keptn.TryGetBridgeURLForKeptnContext(eh.event)),
		AttachRules:      *eh.attachRules,
	}

	dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(ctx, infoEvent)
	return nil
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

	return defaultValue
}

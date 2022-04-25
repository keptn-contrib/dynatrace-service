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

	imageAndTag := eh.eClient.GetImageAndTag(eh.event)

	ie := createInfoEventDTO(eh.event, imageAndTag, eh.attachRules)
	if strategy == keptnevents.Direct && eh.event.GetResult() == keptnv2.ResultPass || eh.event.GetResult() == keptnv2.ResultWarning {
		title := fmt.Sprintf("PROMOTING from %s to next stage", eh.event.GetStage())
		ie.Title = title
		ie.Description = title
	} else if eh.event.GetResult() == keptnv2.ResultFailed {
		if strategy == keptnevents.Duplicate {
			title := "Rollback Artifact (Switch Blue/Green) in " + eh.event.GetStage()
			ie.Title = title
			ie.Description = title
		} else {
			title := fmt.Sprintf("NOT PROMOTING from %s to next stage", eh.event.GetStage())
			ie.Title = title
			ie.Description = title
		}
	}

	dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(ctx, ie)

	return nil
}

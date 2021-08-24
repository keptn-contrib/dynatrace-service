package deployment

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type ReleaseTriggeredEventHandler struct {
	event  *ReleaseTriggeredAdapter
	client *dynatrace.Client
	config *config.DynatraceConfigFile
}

// NewReleaseTriggeredEventHandler creates a new ReleaseTriggeredEventHandler
func NewReleaseTriggeredEventHandler(event *ReleaseTriggeredAdapter, client *dynatrace.Client, config *config.DynatraceConfigFile) *ReleaseTriggeredEventHandler {
	return &ReleaseTriggeredEventHandler{
		event:  event,
		client: client,
		config: config,
	}
}

// Handle handles an action finished event
func (eh *ReleaseTriggeredEventHandler) Handle() error {
	strategy, err := keptnevents.GetDeploymentStrategy(eh.event.GetDeploymentStrategy())
	if err != nil {
		log.WithError(err).Error("Could not determine deployment strategy")
		return err
	}

	ie := event.CreateInfoEvent(eh.event, eh.config)
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

	dynatrace.NewEventsClient(eh.client).SendEvent(ie)

	return nil
}

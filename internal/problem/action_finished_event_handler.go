package problem

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type ActionFinishedEventHandler struct {
	event       ActionFinishedAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewActionFinishedEventHandler creates a new ActionFinishedEventHandler
func NewActionFinishedEventHandler(event ActionFinishedAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *ActionFinishedEventHandler {
	return &ActionFinishedEventHandler{
		event:       event,
		dtClient:    dtClient,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event.
func (eh *ActionFinishedEventHandler) HandleEvent(ctx context.Context) error {
	// lets find our dynatrace problem details for this remediation workflow
	pid, err := eh.eClient.FindProblemID(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	// Comment text we want to push over
	comment := fmt.Sprintf("[Keptn finished execution](%s) of action by: %s\nResult: %s\nStatus: %s",
		eh.event.GetLabels()[common.BridgeLabel],
		eh.event.GetSource(),
		eh.event.GetResult(),
		eh.event.GetStatus())

	imageAndTag := eh.eClient.GetImageAndTag(eh.event)

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// Additionally to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	if eh.event.GetStatus() == keptnv2.StatusSucceeded {

		dtConfigEvent := dynatrace.CreateConfigurationEventDTO(eh.event, imageAndTag, eh.attachRules)
		dtConfigEvent.Description = "Keptn Remediation Action Finished"
		dtConfigEvent.Configuration = "successful"

		dynatrace.NewEventsClient(eh.dtClient).AddConfigurationEvent(ctx, dtConfigEvent)
	} else {
		dtInfoEvent := dynatrace.CreateInfoEventDTO(eh.event, imageAndTag, eh.attachRules)
		dtInfoEvent.Title = "Keptn Remediation Action Finished"
		dtInfoEvent.Description = "error during execution"

		dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(ctx, dtInfoEvent)
	}

	dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(ctx, pid, comment)

	return nil
}

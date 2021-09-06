package problem

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type ActionFinishedEventHandler struct {
	event       ActionFinishedAdapterInterface
	client      dynatrace.ClientInterface
	attachRules *dynatrace.DtAttachRules
}

// NewActionFinishedEventHandler creates a new ActionFinishedEventHandler
func NewActionFinishedEventHandler(event ActionFinishedAdapterInterface, client dynatrace.ClientInterface, attachRules *dynatrace.DtAttachRules) *ActionFinishedEventHandler {
	return &ActionFinishedEventHandler{
		event:       event,
		client:      client,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event
func (eh *ActionFinishedEventHandler) HandleEvent() error {
	// lets find our dynatrace problem details for this remediation workflow
	pid, err := common.FindProblemIDForEvent(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	// Comment text we want to push over
	comment := fmt.Sprintf("[Keptn finished execution](%s) of action by: %s\nResult: %s\nStatus: %s",
		eh.event.GetLabels()[common.KEPTNSBRIDGE_LABEL],
		eh.event.GetSource(),
		eh.event.GetResult(),
		eh.event.GetStatus())

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// Additionally to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	if eh.event.GetStatus() == keptnv2.StatusSucceeded {
		dtConfigEvent := dynatrace.CreateConfigurationEventDTO(eh.event, eh.attachRules)
		dtConfigEvent.Description = "Keptn Remediation Action Finished"
		dtConfigEvent.Configuration = "successful"

		dynatrace.NewEventsClient(eh.client).AddConfigurationEvent(dtConfigEvent)
	} else {
		dtInfoEvent := dynatrace.CreateInfoEventDTO(eh.event, eh.attachRules)
		dtInfoEvent.Title = "Keptn Remediation Action Finished"
		dtInfoEvent.Description = "error during execution"

		dynatrace.NewEventsClient(eh.client).AddInfoEvent(dtInfoEvent)
	}

	dynatrace.NewProblemsClient(eh.client).AddProblemComment(pid, comment)

	return nil
}

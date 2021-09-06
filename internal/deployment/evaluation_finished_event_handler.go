package deployment

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type EvaluationFinishedEventHandler struct {
	event       EvaluationFinishedAdapterInterface
	client      dynatrace.ClientInterface
	attachRules *config.DtAttachRules
}

// NewEvaluationFinishedEventHandler creates a new EvaluationFinishedEventHandler
func NewEvaluationFinishedEventHandler(event EvaluationFinishedAdapterInterface, client dynatrace.ClientInterface, attachRules *config.DtAttachRules) *EvaluationFinishedEventHandler {
	return &EvaluationFinishedEventHandler{
		event:       event,
		client:      client,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event
func (eh *EvaluationFinishedEventHandler) HandleEvent() error {
	// Send Info Event
	ie := event.CreateInfoEvent(eh.event, eh.attachRules)
	qualityGateDescription := fmt.Sprintf("Quality Gate Result in stage %s: %s (%.2f/100)", eh.event.GetStage(), eh.event.GetResult(), eh.event.GetEvaluationScore())
	ie.Title = fmt.Sprintf("Evaluation result: %s", eh.event.GetResult())

	if eh.event.IsPartOfRemediation() {
		if eh.event.GetResult() == keptnv2.ResultPass || eh.event.GetResult() == keptnv2.ResultWarning {
			ie.Title = "Remediation action successful"
		} else {
			ie.Title = "Remediation action not successful"
		}
		// If evaluation was done in context of a problem remediation workflow then post comments to the Dynatrace Problem
		pid, err := common.FindProblemIDForEvent(eh.event)
		if err == nil && pid != "" {
			// Comment we push over
			comment := fmt.Sprintf("[Keptn remediation evaluation](%s) resulted in %s (%.2f/100)", eh.event.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.event.GetResult(), eh.event.GetEvaluationScore())

			// this is posting the Event on the problem as a comment
			dynatrace.NewProblemsClient(eh.client).AddProblemComment(pid, comment)
		}
	}
	ie.Description = qualityGateDescription

	dynatrace.NewEventsClient(eh.client).SendEvent(ie)

	return nil
}

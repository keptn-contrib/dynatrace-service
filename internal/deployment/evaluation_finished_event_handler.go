package deployment

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type EvaluationFinishedEventHandler struct {
	event       EvaluationFinishedAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewEvaluationFinishedEventHandler creates a new EvaluationFinishedEventHandler
func NewEvaluationFinishedEventHandler(event EvaluationFinishedAdapterInterface, client dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *EvaluationFinishedEventHandler {
	return &EvaluationFinishedEventHandler{
		event:       event,
		dtClient:    client,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action finished event
func (eh *EvaluationFinishedEventHandler) HandleEvent() error {

	imageAndTag := eh.eClient.GetImageAndTag(eh.event)

	ie := dynatrace.CreateInfoEventDTO(eh.event, imageAndTag, eh.attachRules)
	qualityGateDescription := fmt.Sprintf("Quality Gate Result in stage %s: %s (%.2f/100)", eh.event.GetStage(), eh.event.GetResult(), eh.event.GetEvaluationScore())
	ie.Title = fmt.Sprintf("Evaluation result: %s", eh.event.GetResult())

	isPartOfRemediation, err := eh.eClient.IsPartOfRemediation(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not check for remediation status of event")
	}

	if isPartOfRemediation {
		if eh.event.GetResult() == keptnv2.ResultPass || eh.event.GetResult() == keptnv2.ResultWarning {
			ie.Title = "Remediation action successful"
		} else {
			ie.Title = "Remediation action not successful"
		}
		// If evaluation was done in context of a problem remediation workflow then post comments to the Dynatrace Problem
		pid, err := eh.eClient.FindProblemID(eh.event)
		if err == nil && pid != "" {
			// Comment we push over
			comment := fmt.Sprintf("[Keptn remediation evaluation](%s) resulted in %s (%.2f/100)", eh.event.GetLabels()[common.BridgeLabel], eh.event.GetResult(), eh.event.GetEvaluationScore())

			// this is posting the Event on the problem as a comment
			dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(pid, comment)
		}
	}
	ie.Description = qualityGateDescription

	dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(ie)

	return nil
}

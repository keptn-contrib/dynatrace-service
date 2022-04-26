package action

import (
	"context"
	"fmt"

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

// HandleEvent handles an action finished event.
func (eh *EvaluationFinishedEventHandler) HandleEvent(ctx context.Context) error {

	isPartOfRemediation, err := eh.eClient.IsPartOfRemediation(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not check for remediation status of event")
	}

	bridgeURL := keptn.TryGetBridgeURLForKeptnContext(eh.event)

	if isPartOfRemediation {
		pid, err := eh.eClient.FindProblemID(eh.event)
		if err == nil && pid != "" {
			comment := fmt.Sprintf("[Keptn remediation evaluation](%s) resulted in %s (%.2f/100)", bridgeURL, eh.event.GetResult(), eh.event.GetEvaluationScore())
			dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(ctx, pid, comment)
		}
	}

	infoEvent := dynatrace.InfoEvent{
		EventType:        dynatrace.InfoEventType,
		Source:           eventSource,
		Title:            eh.getTitle(isPartOfRemediation),
		Description:      fmt.Sprintf("Quality Gate Result in stage %s: %s (%.2f/100)", eh.event.GetStage(), eh.event.GetResult(), eh.event.GetEvaluationScore()),
		CustomProperties: createCustomProperties(eh.event, eh.eClient.GetImageAndTag(eh.event), bridgeURL),
		AttachRules:      *eh.attachRules,
	}

	dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(ctx, infoEvent)

	return nil
}

func (eh *EvaluationFinishedEventHandler) getTitle(isPartOfRemediation bool) string {
	if !isPartOfRemediation {
		return fmt.Sprintf("Evaluation result: %s", eh.event.GetResult())
	}

	if eh.event.GetResult() == keptnv2.ResultPass || eh.event.GetResult() == keptnv2.ResultWarning {
		return "Remediation action successful"
	}

	return "Remediation action not successful"
}

package action

import (
	"context"
	"fmt"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

const evaluationURLKey = "evaluationHeatmapURL"

// EvaluationFinishedEventHandler handles an evaluation finished event.
type EvaluationFinishedEventHandler struct {
	event       EvaluationFinishedAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewEvaluationFinishedEventHandler creates a new EvaluationFinishedEventHandler.
func NewEvaluationFinishedEventHandler(event EvaluationFinishedAdapterInterface, client dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *EvaluationFinishedEventHandler {
	return &EvaluationFinishedEventHandler{
		event:       event,
		dtClient:    client,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an evaluation finished event.
func (eh *EvaluationFinishedEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {

	isPartOfRemediation, err := eh.eClient.IsPartOfRemediation(workCtx, eh.event)
	if err != nil {
		log.WithError(err).Error("Could not check for remediation status of event")
	}

	bridgeURL := keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event)

	if isPartOfRemediation {
		pid, err := eh.eClient.FindProblemID(workCtx, eh.event)
		if err == nil && pid != "" {
			comment := fmt.Sprintf("[Keptn remediation evaluation](%s) resulted in %s (%.2f/100)", bridgeURL, eh.event.GetResult(), eh.event.GetEvaluationScore())
			dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(workCtx, pid, comment)
		}
	}

	imageAndTag := eh.eClient.GetImageAndTag(workCtx, eh.event)
	attachRules := eh.createAttachRules(workCtx, imageAndTag)

	customProperties := newCustomProperties(eh.event, imageAndTag, bridgeURL)
	customProperties.addIfNonEmpty(evaluationURLKey, keptn.TryGetBridgeURLForEvaluation(workCtx, eh.event))

	infoEvent := dynatrace.InfoEvent{
		EventType:        dynatrace.InfoEventType,
		Source:           eventSource,
		Title:            eh.getTitle(isPartOfRemediation),
		Description:      fmt.Sprintf("Quality Gate Result in stage %s: %s (%.2f/100)", eh.event.GetStage(), eh.event.GetResult(), eh.event.GetEvaluationScore()),
		CustomProperties: customProperties,
		AttachRules:      attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(workCtx, infoEvent)
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

func (eh *EvaluationFinishedEventHandler) createAttachRules(ctx context.Context, imageAndTag common.ImageAndTag) dynatrace.AttachRules {
	timeframe, err := common.NewTimeframeParser(eh.event.GetStartTime(), eh.event.GetEndTime()).Parse()
	if err != nil {
		log.WithFields(log.Fields{
			"start": eh.event.GetStartTime(),
			"end":   eh.event.GetEndTime(),
		}).Error("Could not parse evaluation finished timeframe")
	}

	return createOrUpdateAttachRules(ctx, eh.dtClient, eh.attachRules, imageAndTag, eh.event, timeframe)
}

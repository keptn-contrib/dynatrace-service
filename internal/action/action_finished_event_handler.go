package action

import (
	"context"
	"fmt"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type ActionFinishedEventHandler struct {
	event            ActionFinishedAdapterInterface
	dtClient         dynatrace.ClientInterface
	eClient          keptn.EventClientInterface
	bridgeURLCreator keptn.BridgeURLCreatorInterface
	attachRules      *dynatrace.AttachRules
}

// NewActionFinishedEventHandler creates a new ActionFinishedEventHandler
func NewActionFinishedEventHandler(event ActionFinishedAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, bridgeURLCreator keptn.BridgeURLCreatorInterface, attachRules *dynatrace.AttachRules) *ActionFinishedEventHandler {
	return &ActionFinishedEventHandler{
		event:            event,
		dtClient:         dtClient,
		eClient:          eClient,
		bridgeURLCreator: bridgeURLCreator,
		attachRules:      attachRules,
	}
}

// HandleEvent handles an action finished event.
func (eh *ActionFinishedEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	// lets find our dynatrace problem details for this remediation workflow
	pid, err := eh.eClient.FindProblemID(workCtx, eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	bridgeURL := eh.bridgeURLCreator.TryGetBridgeURLForKeptnContext(workCtx, eh.event)

	comment := fmt.Sprintf("[Keptn finished execution](%s) of action by: %s\nResult: %s\nStatus: %s",
		bridgeURL,
		eh.event.GetSource(),
		eh.event.GetResult(),
		eh.event.GetStatus())
	dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(workCtx, pid, comment)

	if eh.attachRules == nil {
		eh.attachRules = createDefaultAttachRules(eh.event)
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// Additionally to the problem comment, send Info or Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	customProperties := newCustomProperties(eh.event, eh.eClient.GetImageAndTag(workCtx, eh.event), bridgeURL)
	if eh.event.GetStatus() == keptnv2.StatusSucceeded {
		configurationEvent := dynatrace.ConfigurationEvent{
			EventType:        dynatrace.ConfigurationEventType,
			Description:      "Keptn Remediation Action Finished",
			Source:           eventSource,
			Configuration:    "successful",
			CustomProperties: customProperties,
			AttachRules:      *eh.attachRules,
		}

		return dynatrace.NewEventsClient(eh.dtClient).AddConfigurationEvent(workCtx, configurationEvent)
	}

	infoEvent := dynatrace.InfoEvent{
		EventType:        dynatrace.InfoEventType,
		Source:           eventSource,
		Title:            "Keptn Remediation Action Finished",
		Description:      "error during execution",
		CustomProperties: customProperties,
		AttachRules:      *eh.attachRules,
	}

	return dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(workCtx, infoEvent)
}

package action

import (
	"context"
	"fmt"

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

	bridgeURL := keptn.TryGetBridgeURLForKeptnContext(eh.event)

	comment := fmt.Sprintf("[Keptn finished execution](%s) of action by: %s\nResult: %s\nStatus: %s",
		bridgeURL,
		eh.event.GetSource(),
		eh.event.GetResult(),
		eh.event.GetStatus())
	dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(ctx, pid, comment)

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// Additionally to the problem comment, send Info or Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	customProperties := createCustomProperties(eh.event, eh.eClient.GetImageAndTag(eh.event), bridgeURL)
	if eh.event.GetStatus() == keptnv2.StatusSucceeded {
		configurationEvent := dynatrace.ConfigurationEvent{
			EventType:        dynatrace.ConfigurationEventType,
			Description:      "Keptn Remediation Action Finished",
			Source:           eventSource,
			Configuration:    "successful",
			CustomProperties: customProperties,
			AttachRules:      *eh.attachRules,
		}

		dynatrace.NewEventsClient(eh.dtClient).AddConfigurationEvent(ctx, configurationEvent)
	} else {
		infoEvent := dynatrace.InfoEvent{
			EventType:        dynatrace.InfoEventType,
			Source:           eventSource,
			Title:            "Keptn Remediation Action Finished",
			Description:      "error during execution",
			CustomProperties: customProperties,
			AttachRules:      *eh.attachRules,
		}

		dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(ctx, infoEvent)
	}

	return nil
}

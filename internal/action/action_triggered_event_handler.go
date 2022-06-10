package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	log "github.com/sirupsen/logrus"
)

type ActionTriggeredEventHandler struct {
	event       ActionTriggeredAdapterInterface
	dtClient    dynatrace.ClientInterface
	eClient     keptn.EventClientInterface
	attachRules *dynatrace.AttachRules
}

// NewActionTriggeredEventHandler creates a new ActionTriggeredEventHandler
func NewActionTriggeredEventHandler(event ActionTriggeredAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules) *ActionTriggeredEventHandler {
	return &ActionTriggeredEventHandler{
		event:       event,
		dtClient:    dtClient,
		eClient:     eClient,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action triggered event.
func (eh *ActionTriggeredEventHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	pid, err := eh.eClient.FindProblemID(workCtx, eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	if pid == "" {
		log.Error("Cannot send DT problem comment: No problem ID is included in the event.")
		return errors.New("cannot send DT problem comment: no problem ID is included in the event")
	}

	bridgeURL := keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event)

	comment := fmt.Sprintf("[Keptn triggered action](%s) %s", bridgeURL, eh.event.GetAction())
	if eh.event.GetActionDescription() != "" {
		comment = comment + ": " + eh.event.GetActionDescription()
	}

	dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(workCtx, pid, comment)

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// In addition to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	infoEvent := dynatrace.InfoEvent{
		EventType:        dynatrace.InfoEventType,
		Source:           eventSource,
		Title:            "Keptn Remediation Action Triggered",
		Description:      eh.event.GetAction(),
		CustomProperties: createCustomProperties(eh.event, eh.eClient.GetImageAndTag(workCtx, eh.event), bridgeURL),
		AttachRules:      *eh.attachRules,
	}

	dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(workCtx, infoEvent)

	return nil
}

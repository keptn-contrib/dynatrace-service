package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
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
func (eh *ActionTriggeredEventHandler) HandleEvent(ctx context.Context) error {
	pid, err := eh.eClient.FindProblemID(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	if pid == "" {
		log.Error("Cannot send DT problem comment: No problem ID is included in the event.")
		return errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	comment := "Keptn triggered action " + eh.event.GetAction()
	if eh.event.GetActionDescription() != "" {
		comment = comment + ": " + eh.event.GetActionDescription()
	}

	imageAndTag := eh.eClient.GetImageAndTag(eh.event)

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// In addition to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	dtInfoEvent := dynatrace.CreateInfoEventDTO(eh.event, imageAndTag, eh.attachRules)
	dtInfoEvent.Title = "Keptn Remediation Action Triggered"
	dtInfoEvent.Description = eh.event.GetAction()

	dynatrace.NewEventsClient(eh.dtClient).AddInfoEvent(ctx, dtInfoEvent)

	// this is posting the Event on the problem as a comment
	comment = fmt.Sprintf("[Keptn triggered action](%s) %s", eh.event.GetLabels()[common.BridgeLabel], eh.event.GetAction())
	if eh.event.GetActionDescription() != "" {
		comment = comment + ": " + eh.event.GetActionDescription()
	}

	dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(ctx, pid, comment)

	return nil
}

package problem

import (
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	log "github.com/sirupsen/logrus"
)

type ActionTriggeredEventHandler struct {
	event       ActionTriggeredAdapterInterface
	client      dynatrace.ClientInterface
	attachRules *config.DtAttachRules
}

// NewActionTriggeredEventHandler creates a new ActionTriggeredEventHandler
func NewActionTriggeredEventHandler(event ActionTriggeredAdapterInterface, client dynatrace.ClientInterface, attachRules *config.DtAttachRules) *ActionTriggeredEventHandler {
	return &ActionTriggeredEventHandler{
		event:       event,
		client:      client,
		attachRules: attachRules,
	}
}

// HandleEvent handles an action triggered event
func (eh *ActionTriggeredEventHandler) HandleEvent() error {
	pid, err := common.FindProblemIDForEvent(eh.event)
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

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// In addition to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	dtInfoEvent := event.CreateInfoEvent(eh.event, eh.attachRules)
	dtInfoEvent.Title = "Keptn Remediation Action Triggered"
	dtInfoEvent.Description = eh.event.GetAction()

	dynatrace.NewEventsClient(eh.client).SendEvent(dtInfoEvent)

	// this is posting the Event on the problem as a comment
	comment = fmt.Sprintf("[Keptn triggered action](%s) %s", eh.event.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.event.GetAction())
	if eh.event.GetActionDescription() != "" {
		comment = comment + ": " + eh.event.GetActionDescription()
	}

	dynatrace.NewProblemsClient(eh.client).AddProblemComment(pid, comment)

	return nil
}

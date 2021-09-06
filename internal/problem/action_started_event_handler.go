package problem

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	log "github.com/sirupsen/logrus"
)

type ActionStartedEventHandler struct {
	event  ActionStartedAdapterInterface
	client *dynatrace.Client
}

// NewActionStartedEventHandler creates a new ActionStartedEventHandler
func NewActionStartedEventHandler(event ActionStartedAdapterInterface, client *dynatrace.Client) *ActionStartedEventHandler {
	return &ActionStartedEventHandler{
		event:  event,
		client: client,
	}
}

// HandleEvent handles an action started event
func (eh *ActionStartedEventHandler) HandleEvent() error {
	pid, err := common.FindProblemIDForEvent(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	// Comment we push over
	comment := fmt.Sprintf("[Keptn remediation action](%s) started execution by: %s", eh.event.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.event.GetSource())

	dynatrace.NewProblemsClient(eh.client).AddProblemComment(pid, comment)

	return nil
}

package problem

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	log "github.com/sirupsen/logrus"
)

type ActionStartedEventHandler struct {
	event       *ActionStartedAdapter
	client      *dynatrace.Client
	eventSource string
}

// NewActionStartedEventHandler creates a new ActionStartedEventHandler
func NewActionStartedEventHandler(event *ActionStartedAdapter, client *dynatrace.Client, eventSource string) *ActionStartedEventHandler {
	return &ActionStartedEventHandler{
		event:       event,
		client:      client,
		eventSource: eventSource,
	}
}

// Handle handles an action started event
func (eh *ActionStartedEventHandler) Handle() error {
	pid, err := common.FindProblemIDForEvent(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	// Comment we push over
	comment := fmt.Sprintf("[Keptn remediation action](%s) started execution by: %s", eh.event.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.eventSource)

	dynatrace.NewProblemsClient(eh.client).AddProblemComment(pid, comment)

	return nil
}

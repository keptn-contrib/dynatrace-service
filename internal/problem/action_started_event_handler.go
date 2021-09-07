package problem

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	log "github.com/sirupsen/logrus"
)

type ActionStartedEventHandler struct {
	event    ActionStartedAdapterInterface
	dtClient dynatrace.ClientInterface
	eClient  keptn.EventClientInterface
}

// NewActionStartedEventHandler creates a new ActionStartedEventHandler
func NewActionStartedEventHandler(event ActionStartedAdapterInterface, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface) *ActionStartedEventHandler {
	return &ActionStartedEventHandler{
		event:    event,
		dtClient: dtClient,
		eClient:  eClient,
	}
}

// HandleEvent handles an action started event
func (eh *ActionStartedEventHandler) HandleEvent() error {
	pid, err := eh.eClient.FindProblemID(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	// Comment we push over
	comment := fmt.Sprintf("[Keptn remediation action](%s) started execution by: %s", eh.event.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.event.GetSource())

	dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(pid, comment)

	return nil
}

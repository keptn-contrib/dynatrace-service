package action

import (
	"context"
	"fmt"

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

// HandleEvent handles an action started event.
func (eh *ActionStartedEventHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	pid, err := eh.eClient.FindProblemID(eh.event)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	comment := fmt.Sprintf("[Keptn remediation action](%s) started execution by: %s", keptn.TryGetBridgeURLForKeptnContext(workCtx, eh.event), eh.event.GetSource())
	dynatrace.NewProblemsClient(eh.dtClient).AddProblemComment(workCtx, pid, comment)

	return nil
}

package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type ProjectCreateFinishedEventHandler struct {
	event    ProjectCreateFinishedAdapterInterface
	dtClient dynatrace.ClientInterface
	kClient  *keptnv2.Keptn
}

// NewProjectCreateFinishedEventHandler creates a new ProjectCreateFinishedEventHandler
func NewProjectCreateFinishedEventHandler(event ProjectCreateFinishedAdapterInterface, dtClient dynatrace.ClientInterface, kClient *keptnv2.Keptn) ProjectCreateFinishedEventHandler {
	return ProjectCreateFinishedEventHandler{
		event:    event,
		dtClient: dtClient,
		kClient:  kClient,
	}
}

func (eh ProjectCreateFinishedEventHandler) HandleEvent() error {
	shipyard, err := eh.event.GetShipyard()
	if err != nil {
		log.WithError(err).Error("Could not load Keptn shipyard file")
	}

	cfg := NewConfiguration(eh.dtClient, eh.kClient)

	_, err = cfg.ConfigureMonitoring(eh.event.GetProject(), shipyard)
	if err != nil {
		return err
	}

	log.Info("Dynatrace Monitoring setup done")
	return nil
}

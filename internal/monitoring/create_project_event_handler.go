package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type CreateProjectEventHandler struct {
	event    ProjectCreateAdapterInterface
	dtClient *dynatrace.Client
	kClient  *keptnv2.Keptn
}

// NewCreateProjectEventHandler creates a new CreateProjectEventHandler
func NewCreateProjectEventHandler(event ProjectCreateAdapterInterface, dtClient *dynatrace.Client, kClient *keptnv2.Keptn) CreateProjectEventHandler {
	return CreateProjectEventHandler{
		event:    event,
		dtClient: dtClient,
		kClient:  kClient,
	}
}

func (eh CreateProjectEventHandler) HandleEvent() error {
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

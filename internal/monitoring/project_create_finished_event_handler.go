package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	log "github.com/sirupsen/logrus"
)

type ProjectCreateFinishedEventHandler struct {
	event          ProjectCreateFinishedAdapterInterface
	dtClient       dynatrace.ClientInterface
	kClient        keptn.ClientInterface
	resourceClient keptn.ResourceClientInterface
	serviceClient  keptn.ServiceClientInterface
}

// NewProjectCreateFinishedEventHandler creates a new ProjectCreateFinishedEventHandler
func NewProjectCreateFinishedEventHandler(event ProjectCreateFinishedAdapterInterface, dtClient dynatrace.ClientInterface, kClient keptn.ClientInterface, resourceClient keptn.ResourceClientInterface, serviceClient keptn.ServiceClientInterface) ProjectCreateFinishedEventHandler {
	return ProjectCreateFinishedEventHandler{
		event:          event,
		dtClient:       dtClient,
		kClient:        kClient,
		resourceClient: resourceClient,
		serviceClient:  serviceClient,
	}
}

func (eh ProjectCreateFinishedEventHandler) HandleEvent() error {
	shipyard, err := eh.event.GetShipyard()
	if err != nil {
		log.WithError(err).Error("Could not load Keptn shipyard file")
	}

	cfg := NewConfiguration(eh.dtClient, eh.kClient, eh.resourceClient, eh.serviceClient)

	_, err = cfg.ConfigureMonitoring(eh.event.GetProject(), shipyard)
	if err != nil {
		return err
	}

	log.Info("Dynatrace Monitoring setup done")
	return nil
}

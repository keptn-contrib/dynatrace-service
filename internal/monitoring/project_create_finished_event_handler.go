package monitoring

import (
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type ProjectCreateFinishedEventHandler struct {
	event         ProjectCreateFinishedAdapterInterface
	dtClient      dynatrace.ClientInterface
	kClient       keptn.ClientInterface
	sloReader     keptn.SLOResourceReaderInterface
	serviceClient keptn.ServiceClientInterface
}

// NewProjectCreateFinishedEventHandler creates a new ProjectCreateFinishedEventHandler
func NewProjectCreateFinishedEventHandler(event ProjectCreateFinishedAdapterInterface, dtClient dynatrace.ClientInterface, kClient keptn.ClientInterface, sloReader keptn.SLOResourceReaderInterface, serviceClient keptn.ServiceClientInterface) ProjectCreateFinishedEventHandler {
	return ProjectCreateFinishedEventHandler{
		event:         event,
		dtClient:      dtClient,
		kClient:       kClient,
		sloReader:     sloReader,
		serviceClient: serviceClient,
	}
}

func (eh ProjectCreateFinishedEventHandler) HandleEvent() error {
	shipyard, err := eh.event.GetShipyard()
	if err != nil {
		log.WithError(err).Error("Could not load Keptn shipyard file")
	}

	cfg := NewConfiguration(eh.dtClient, eh.kClient, eh.sloReader, eh.serviceClient)

	_, err = cfg.ConfigureMonitoring(eh.event.GetProject(), shipyard)
	if err != nil {
		return err
	}

	log.Info("Dynatrace Monitoring setup done")
	return nil
}

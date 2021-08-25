package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type CreateProjectEventHandler struct {
	event         *ProjectCreateAdapter
	client        *dynatrace.Client
	incomingEvent cloudevents.Event
}

// NewCreateProjectEventHandler creates a new CreateProjectEventHandler
func NewCreateProjectEventHandler(event *ProjectCreateAdapter, client *dynatrace.Client, incomingEvent cloudevents.Event) CreateProjectEventHandler {
	return CreateProjectEventHandler{
		event:         event,
		client:        client,
		incomingEvent: incomingEvent,
	}
}

func (eh CreateProjectEventHandler) HandleEvent() error {
	shipyard, err := eh.event.GetShipyard()
	if err != nil {
		log.WithError(err).Error("Could not load Keptn shipyard file")
	}

	keptnHandler, err := keptnv2.NewKeptn(&eh.incomingEvent, keptn.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not create Keptn handler")
	}

	cfg := NewConfiguration(eh.client, keptnHandler)

	_, err = cfg.ConfigureMonitoring(eh.event.GetProject(), shipyard)
	if err != nil {
		return err
	}

	log.Info("Dynatrace Monitoring setup done")
	return nil
}

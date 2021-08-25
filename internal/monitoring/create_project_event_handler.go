package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type CreateProjectEventHandler struct {
	event          cloudevents.Event
	dtConfigGetter config.DynatraceConfigGetterInterface
}

func NewCreateProjectEventHandler(event cloudevents.Event, configGetter config.DynatraceConfigGetterInterface) CreateProjectEventHandler {
	return CreateProjectEventHandler{
		event:          event,
		dtConfigGetter: configGetter,
	}
}

func (eh CreateProjectEventHandler) HandleEvent() error {
	keptnEvent, err := NewProjectCreateAdapterFromEvent(eh.event)
	if err != nil {
		return err
	}

	shipyard, err := keptnEvent.GetShipyard()
	if err != nil {
		log.WithError(err).Error("Could not load Keptn shipyard file")
	}

	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("failed to load Dynatrace config")
		return err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace credentials")
		return err
	}
	keptnHandler, err := keptnv2.NewKeptn(&eh.event, keptn.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not create Keptn handler")
	}

	cfg := NewConfiguration(dynatrace.NewClient(creds), keptnHandler)

	_, err = cfg.ConfigureMonitoring(keptnEvent.GetProject(), shipyard)
	if err != nil {
		return err
	}

	log.Info("Dynatrace Monitoring setup done")
	return nil
}

package event_handler

import (
	"encoding/base64"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"

	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	"gopkg.in/yaml.v2"
)

type CreateProjectEventHandler struct {
	Event          cloudevents.Event
	dtConfigGetter adapter.DynatraceConfigGetterInterface
}

func (eh CreateProjectEventHandler) HandleEvent() error {

	e := &keptnv2.ProjectCreateFinishedEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		log.WithError(err).Error("Could not parse event payload")
		return err
	}

	shipyard := &keptnv2.Shipyard{}
	decodedShipyard, err := base64.StdEncoding.DecodeString(e.CreatedProject.Shipyard)
	if err != nil {
		log.WithError(err).Error("Could not decode shipyard")
	}
	err = yaml.Unmarshal(decodedShipyard, shipyard)
	if err != nil {
		log.WithError(err).Error("Could not parse shipyard")
	}

	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not create Keptn handler")
	}

	keptnEvent := adapter.NewProjectCreateAdapter(*e, keptnHandler.KeptnContext, eh.Event.Source())

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
	config := monitoring.NewConfiguration(lib.NewDynatraceHelper(keptnHandler, creds))

	_, err = config.ConfigureMonitoring(e.Project, shipyard)
	if err != nil {
		return err
	}

	log.Info("Dynatrace Monitoring setup done")
	return nil
}

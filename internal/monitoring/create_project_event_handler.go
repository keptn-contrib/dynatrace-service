package monitoring

import (
	"encoding/base64"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"gopkg.in/yaml.v2"
)

type CreateProjectEventHandler struct {
	event          cloudevents.Event
	dtConfigGetter adapter.DynatraceConfigGetterInterface
}

func NewCreateProjectEventHandler(event cloudevents.Event, configGetter adapter.DynatraceConfigGetterInterface) CreateProjectEventHandler {
	return CreateProjectEventHandler{
		event:          event,
		dtConfigGetter: configGetter,
	}
}

func (eh CreateProjectEventHandler) HandleEvent() error {

	e := &keptnv2.ProjectCreateFinishedEventData{}
	err := eh.event.DataAs(e)
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

	keptnHandler, err := keptnv2.NewKeptn(&eh.event, keptn.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not create Keptn handler")
	}

	keptnEvent := adapter.NewProjectCreateAdapter(*e, keptnHandler.KeptnContext, eh.event.Source())

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
	config := NewConfiguration(dynatrace.NewClient(creds), keptnHandler)

	_, err = config.ConfigureMonitoring(e.Project, shipyard)
	if err != nil {
		return err
	}

	log.Info("Dynatrace Monitoring setup done")
	return nil
}

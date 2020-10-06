package event_handler

import (
	"encoding/base64"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"gopkg.in/yaml.v2"
)

type CreateProjectEventHandler struct {
	Logger keptn.LoggerInterface
	Event  cloudevents.Event
}

func (eh CreateProjectEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	e := &keptn.ProjectCreateEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		eh.Logger.Error("Could not parse event payload: " + err.Error())
		return err
	}

	var shipyard *keptn.Shipyard
	decodedShipyard, err := base64.StdEncoding.DecodeString(e.Shipyard)
	if err != nil {
		eh.Logger.Error("Could not decode shipyard: " + err.Error())
	}
	err = yaml.Unmarshal(decodedShipyard, shipyard)
	if err != nil {
		eh.Logger.Error("Could not parse shipyard: " + err.Error())
	}

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}

	keptnEvent := adapter.NewProjectCreateAdapter(*e, keptnHandler.KeptnContext, eh.Event.Source())

	dynatraceConfig, err := config.GetDynatraceConfig(keptnEvent, eh.Logger)
	if err != nil {
		eh.Logger.Error("failed to load Dynatrace config: " + err.Error())
		return err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		eh.Logger.Error("failed to load Dynatrace credentials: " + err.Error())
		return err
	}
	dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

	err = dtHelper.ConfigureMonitoring(e.Project, shipyard)
	if err != nil {
		return err
	}

	eh.Logger.Info("Dynatrace Monitoring setup done")
	return nil
}

package event_handler

import (
	"encoding/base64"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	"gopkg.in/yaml.v2"
)

type CreateProjectEventHandler struct {
	Logger         keptn.LoggerInterface
	Event          cloudevents.Event
	dtConfigGetter adapter.DynatraceConfigGetterInterface
}

func (eh CreateProjectEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	e := &keptnv2.ProjectCreateFinishedEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		eh.Logger.Error("Could not parse event payload: " + err.Error())
		return err
	}

	shipyard := &keptnv2.Shipyard{}
	decodedShipyard, err := base64.StdEncoding.DecodeString(e.CreatedProject.Shipyard)
	if err != nil {
		eh.Logger.Error("Could not decode shipyard: " + err.Error())
	}
	err = yaml.Unmarshal(decodedShipyard, shipyard)
	if err != nil {
		eh.Logger.Error("Could not parse shipyard: " + err.Error())
	}

	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}

	keptnEvent := adapter.NewProjectCreateAdapter(*e, keptnHandler.KeptnContext, eh.Event.Source())

	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent, eh.Logger)
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

	_, err = dtHelper.ConfigureMonitoring(e.Project, shipyard)
	if err != nil {
		return err
	}

	eh.Logger.Info("Dynatrace Monitoring setup done")
	return nil
}

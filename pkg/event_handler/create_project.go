package event_handler

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/ghodss/yaml"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnmodels "github.com/keptn/go-utils/pkg/models"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

type CreateProjectEventHandler struct {
	Logger   keptnutils.LoggerInterface
	Event    cloudevents.Event
	DTHelper *lib.DynatraceHelper
}

func (eh CreateProjectEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	loggingDone := make(chan bool)
	connData := &keptnutils.ConnectionData{}

	stdLogger := keptnutils.NewLogger(shkeptncontext, eh.Event.Context.GetID(), "dynatrace-service")

	if err := eh.Event.DataAs(connData); err == nil &&
		connData.EventContext.KeptnContext != nil && connData.EventContext.Token != nil {

		ws, _, err := keptnutils.OpenWS(*connData, url.URL{
			Scheme: "http",
			Host:   "api.keptn:8080",
		})
		defer ws.Close()
		if err != nil {
			eh.Logger.Error(fmt.Sprintf("Opening websocket connection failed. %s", err.Error()))
			return nil
		}
		combinedLogger := keptnutils.NewCombinedLogger(stdLogger, ws, shkeptncontext)
		eh.Logger = combinedLogger
		go closeLogger(loggingDone, combinedLogger, ws)
	}

	e := &keptnevents.ProjectCreateEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		eh.Logger.Error("Could not parse event payload: " + err.Error())
		return err
	}

	shipyard := &keptnmodels.Shipyard{}

	decodedShipyard, err := base64.StdEncoding.DecodeString(e.Shipyard)
	if err != nil {
		eh.Logger.Error("Could not decode shipyard: " + err.Error())
	}
	err = yaml.Unmarshal(decodedShipyard, shipyard)
	if err != nil {
		eh.Logger.Error("Could not parse shipyard: " + err.Error())
	}

	clientSet, err := keptnutils.GetClientset(true)
	if err != nil {
		eh.Logger.Error("could not create k8s client")
	}

	dtHelper, err := lib.NewDynatraceHelper()
	if err != nil {
		eh.Logger.Error("Could not create Dynatrace Helper: " + err.Error())
	}
	dtHelper.KubeApi = clientSet
	dtHelper.Logger = eh.Logger
	eh.DTHelper = dtHelper

	err = eh.DTHelper.CreateCalculatedMetrics(e.Project)
	if err != nil {
		eh.Logger.Error("Could not create calculated metrics: " + err.Error())
	}

	err = eh.DTHelper.CreateTestStepCalculatedMetrics(e.Project)
	if err != nil {
		eh.Logger.Error("Could not create calculated metrics: " + err.Error())
	}

	err = eh.DTHelper.CreateDashboard(e.Project, *shipyard, nil)

	return nil
}

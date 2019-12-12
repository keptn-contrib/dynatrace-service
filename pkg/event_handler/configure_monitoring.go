package event_handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/keptn/go-utils/pkg/models"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/gorilla/websocket"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptnmodels "github.com/keptn/go-utils/pkg/configuration-service/models"
	"github.com/keptn/go-utils/pkg/configuration-service/utils"
	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

type ConfigureMonitoringEventHandler struct {
	Logger   keptnutils.LoggerInterface
	Event    cloudevents.Event
	DTHelper *lib.DynatraceHelper
}

func (eh ConfigureMonitoringEventHandler) HandleEvent() error {
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

	e := &keptnevents.ConfigureMonitoringEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		eh.Logger.Error("Could not parse event payload: " + err.Error())
		return err
	}
	if e.Type != "dynatrace" {
		return nil
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

	err = eh.DTHelper.EnsureDTIsInstalled()

	if err != nil {
		eh.Logger.Error("could not install Dynatrace: " + err.Error())
		return err
	}

	err = eh.DTHelper.EnsureDTTaggingRulesAreSetUp()

	if e.Project != "" {
		resourceHandler := utils.NewResourceHandler("configuration-service:8080")
		keptnHandler := keptnutils.NewKeptnHandler(resourceHandler)

		shipyard, err := keptnHandler.GetShipyard(e.Project)
		if err != nil {
			eh.Logger.Error("Could not retrieve shipyard for project " + e.Project + ": " + err.Error())
			return err
		}

		services, err := getServicesInProject(e.Project, *shipyard, "")

		if err != nil {
			eh.Logger.Error("Could not retrieve services of project " + e.Project + ": " + err.Error())
			return err
		}

		err = eh.DTHelper.CreateDashboard(e.Project, *shipyard, services)
		if err != nil {
			eh.Logger.Error("Could not create Dynatrace dashboard for project " + e.Project + ": " + err.Error())
			return err
		}
	}

	loggingDone <- true
	return nil
}

func getServicesInProject(project string, shipyard models.Shipyard, addService string) ([]string, error) {
	services := []string{}
	if addService != "" {
		services = []string{addService}
	}
	req, err := http.NewRequest("GET", "http://configuration-service:8080/v1/project/"+project+"/stage/"+shipyard.Stages[0].Name+"/service?pageSize=50", nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	servicesResponse := &keptnmodels.Services{}

	err = json.Unmarshal(body, servicesResponse)
	if err != nil {
		return nil, err
	}

	for _, service := range servicesResponse.Services {
		if service.ServiceName != addService {
			services = append(services, service.ServiceName)
		}
	}
	return services, nil

}

func closeLogger(loggingDone chan bool, combinedLogger *keptnutils.CombinedLogger, ws *websocket.Conn) {
	<-loggingDone
	combinedLogger.Terminate()
	ws.Close()
}

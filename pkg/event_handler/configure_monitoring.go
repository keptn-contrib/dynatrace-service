package event_handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/gorilla/websocket"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptnmodels "github.com/keptn/go-utils/pkg/api/models"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ConfigureMonitoringEventHandler struct {
	Logger           keptn.LoggerInterface
	Event            cloudevents.Event
	DTHelper         *lib.DynatraceHelper
	IsCombinedLogger bool
	WebSocket        *websocket.Conn
}

func (eh ConfigureMonitoringEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	if eh.Event.Type() == keptn.ConfigureMonitoringEventType {
		eventData := &keptn.ConfigureMonitoringEventData{}
		if err := eh.Event.DataAs(eventData); err != nil {
			return err
		}
		if eventData.Type != "dynatrace" {
			return nil
		}
	}
	// open WebSocket, if connection data is available
	connData := keptn.ConnectionData{}
	if err := eh.Event.DataAs(&connData); err != nil ||
		connData.EventContext.KeptnContext == nil || connData.EventContext.Token == nil ||
		*connData.EventContext.KeptnContext == "" || *connData.EventContext.Token == "" {
		eh.Logger.Debug("No WebSocket connection data available")
	} else {
		eh.openWebSocketLogger(connData, shkeptncontext)
	}
	eh.configureMonitoring()
	eh.closeWebSocketConnection()
	return nil
}

func (eh *ConfigureMonitoringEventHandler) openWebSocketLogger(connData keptn.ConnectionData, shkeptncontext string) {
	wsURL, err := getServiceEndpoint("API_WEBSOCKET_URL")
	if err != nil {
		eh.Logger.Error(err.Error())
		return
	}
	ws, _, err := keptn.OpenWS(connData, wsURL)
	if err != nil {
		eh.Logger.Error("Opening WebSocket connection failed:" + err.Error())
		return
	}
	stdLogger := keptn.NewLogger(shkeptncontext, eh.Event.Context.GetID(), "dynatrace-service")
	combinedLogger := keptn.NewCombinedLogger(stdLogger, ws, shkeptncontext)
	eh.Logger = combinedLogger
	eh.WebSocket = ws
	eh.IsCombinedLogger = true
}

func (eh ConfigureMonitoringEventHandler) configureMonitoring() error {
	eh.Logger.Info("Configuring Dynatrace monitoring")
	e := &keptn.ConfigureMonitoringEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		eh.Logger.Error("could not parse event payload: " + err.Error())
		return err
	}
	if e.Type != "dynatrace" {
		return nil
	}

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}

	clientSet, err := common.GetKubernetesClient()
	if err != nil {
		eh.Logger.Error("could not create k8s client")
	}

	dtHelper, err := lib.NewDynatraceHelper(keptnHandler)
	if err != nil {
		eh.Logger.Error("could not create Dynatrace Helper: " + err.Error())
	}
	dtHelper.KubeApi = clientSet
	dtHelper.Logger = eh.Logger
	eh.DTHelper = dtHelper

	err = eh.DTHelper.EnsureDTTaggingRulesAreSetUp()
	if err != nil {
		eh.Logger.Error("Could not set up tagging rules: " + err.Error())
	}

	err = eh.DTHelper.EnsureProblemNotificationsAreSetUp()
	if err != nil {
		eh.Logger.Error("Could not set up problem notification: " + err.Error())
	}

	if e.Project != "" {

		shipyard, err := keptnHandler.GetShipyard()
		if err != nil {
			eh.Logger.Error("Could not retrieve shipyard for project " + e.Project + ": " + err.Error())
			return err
		}

		err = eh.DTHelper.CreateManagementZones(e.Project, *shipyard)
		if err != nil {
			eh.Logger.Error("Could not create Management Zones for project " + e.Project + ": " + err.Error())
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
			// do not return because there are no dependencies to the dashboard
		}

		// try to create metric events - if one fails, don't fail the whole setup
		for _, stage := range shipyard.Stages {
			if stage.RemediationStrategy == "automated" {
				for _, service := range services {
					_ = eh.DTHelper.CreateMetricEvents(e.Project, stage.Name, service)
				}
			}
		}
	}
	eh.Logger.Info("Dynatrace Monitoring setup done")
	return nil
}

func (eh *ConfigureMonitoringEventHandler) closeWebSocketConnection() {
	if eh.IsCombinedLogger {
		eh.Logger.(*keptn.CombinedLogger).Terminate()
		eh.WebSocket.Close()
	}
}

func getServicesInProject(project string, shipyard keptn.Shipyard, addService string) ([]string, error) {
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

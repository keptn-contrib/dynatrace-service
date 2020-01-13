package event_handler

import (
	"encoding/json"
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
	Logger           keptnutils.LoggerInterface
	Event            cloudevents.Event
	DTHelper         *lib.DynatraceHelper
	IsCombinedLogger bool
	WebSocket        *websocket.Conn
}

func (eh ConfigureMonitoringEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	if eh.Event.Type() == keptnevents.ConfigureMonitoringEventType {
		eventData := &keptnevents.ConfigureMonitoringEventData{}
		if err := eh.Event.DataAs(eventData); err != nil {
			return err
		}
		if eventData.Type != "dynatrace" {
			return nil
		}
	}
	// open WebSocket, if connection data is available
	connData := keptnutils.ConnectionData{}
	if err := eh.Event.DataAs(&connData); err != nil ||
		connData.EventContext.KeptnContext == nil || connData.EventContext.Token == nil ||
		*connData.EventContext.KeptnContext == "" || *connData.EventContext.Token == "" {
		eh.Logger.Debug("No WebSocket connection data available")
	} else {
		apiServiceURL, err := url.Parse("ws://api.keptn.svc.cluster.local")
		if err != nil {
			eh.Logger.Error(err.Error())
			return nil
		}
		ws, _, err := keptnutils.OpenWS(connData, *apiServiceURL)
		if err != nil {
			eh.Logger.Error("Opening WebSocket connection failed:" + err.Error())
			return nil
		}
		stdLogger := keptnutils.NewLogger(shkeptncontext, eh.Event.Context.GetID(), "dynatrace-service")
		combinedLogger := keptnutils.NewCombinedLogger(stdLogger, ws, shkeptncontext)
		eh.Logger = combinedLogger
		eh.WebSocket = ws
		eh.IsCombinedLogger = true
	}
	eh.configureMonitoring()
	eh.closeWebSocketConnection()
	return nil
}

func (eh ConfigureMonitoringEventHandler) configureMonitoring() error {
	eh.Logger.Info("Configuring Dynatrace monitoring")
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
	if err != nil {
		eh.Logger.Error("Could not set up tagging rules: " + err.Error())
	}

	err = eh.DTHelper.EnsureProblemNotificationsAreSetUp()
	if err != nil {
		eh.Logger.Error("Could not set up problem notification: " + err.Error())
	}

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

		err = eh.DTHelper.CreateManagementZones(e.Project, *shipyard)
		if err != nil {
			eh.Logger.Error("Could not create Management Zones for project " + e.Project + ": " + err.Error())
			return err
		}
	}
	eh.Logger.Info("Dynatrace Monitoring setup done")
	return nil
}

func (eh *ConfigureMonitoringEventHandler) closeWebSocketConnection() {
	if eh.IsCombinedLogger {
		eh.Logger.(*keptnutils.CombinedLogger).Terminate()
		eh.WebSocket.Close()
	}
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

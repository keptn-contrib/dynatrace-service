package event_handler

import (
	"fmt"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gorilla/websocket"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ConfigureMonitoringEventHandler struct {
	Logger           keptncommon.LoggerInterface
	Event            cloudevents.Event
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
	connData := keptncommon.ConnectionData{}
	if err := eh.Event.DataAs(&connData); err != nil ||
		connData.EventContext.KeptnContext == nil || connData.EventContext.Token == nil ||
		*connData.EventContext.KeptnContext == "" || *connData.EventContext.Token == "" {
		eh.Logger.Debug("No WebSocket connection data available")
	} else {
		eh.openWebSocketLogger(connData, shkeptncontext)
	}
	err := eh.configureMonitoring()
	if err != nil {
		eh.Logger.Error(err.Error())
	}
	eh.closeWebSocketConnection()
	return nil
}

func (eh *ConfigureMonitoringEventHandler) openWebSocketLogger(connData keptncommon.ConnectionData, shkeptncontext string) {
	wsURL, err := getServiceEndpoint("API_WEBSOCKET_URL")
	if err != nil {
		eh.Logger.Error(err.Error())
		return
	}
	ws, _, err := keptncommon.OpenWS(connData, wsURL)
	if err != nil {
		eh.Logger.Error("Opening WebSocket connection failed:" + err.Error())
		return
	}
	stdLogger := keptncommon.NewLogger(shkeptncontext, eh.Event.Context.GetID(), "dynatrace-service")
	combinedLogger := keptncommon.NewCombinedLogger(stdLogger, ws, shkeptncontext)
	eh.Logger = combinedLogger
	eh.WebSocket = ws
	eh.IsCombinedLogger = true
}

func (eh ConfigureMonitoringEventHandler) configureMonitoring() error {
	eh.Logger.Info("Configuring Dynatrace monitoring")
	e := &keptn.ConfigureMonitoringEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		return fmt.Errorf("could not parse event payload: %v", err)
	}
	if e.Type != "dynatrace" {
		return nil
	}

	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptncommon.KeptnOpts{})
	if err != nil {
		return fmt.Errorf("could not create Keptn handler: %v", err)
	}

	var shipyard *keptnv2.Shipyard
	if e.Project != "" {
		shipyard, err = keptnHandler.GetShipyard()
		if err != nil {
			return fmt.Errorf("failed to retrieve shipyard for project %s: %v", e.Project, err)
		}
	}

	keptnEvent := adapter.NewConfigureMonitoringAdapter(*e, keptnHandler.KeptnContext, eh.Event.Source())

	dynatraceConfig, err := config.GetDynatraceConfig(keptnEvent, eh.Logger)
	if err != nil {
		return fmt.Errorf("failed to load Dynatrace config: %v", err)
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		return fmt.Errorf("failed to load Dynatrace credentials: %v", err)
	}
	dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

	err = dtHelper.ConfigureMonitoring(e.Project, shipyard)
	if err != nil {
		return err
	}

	eh.Logger.Info("Dynatrace Monitoring setup done")
	return nil
}

func (eh *ConfigureMonitoringEventHandler) closeWebSocketConnection() {
	if eh.IsCombinedLogger {
		eh.Logger.(*keptncommon.CombinedLogger).Terminate("")
		eh.WebSocket.Close()
	}
}

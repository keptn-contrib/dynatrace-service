package event_handler

import (
	"fmt"
	"net/url"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/gorilla/websocket"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
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

	eh.DTHelper = &lib.DynatraceHelper{
		KubeApi: clientSet,
		DynatraceCreds: &lib.DTCredentials{
			Tenant:    "",
			ApiToken:  "",
			PaaSToken: "",
		},
		Logger: eh.Logger,
	}

	err = eh.DTHelper.EnsureDTIsInstalled()

	if err != nil {
		eh.Logger.Error("could not install Dynatrace: " + err.Error())
	}

	err = eh.DTHelper.EnsureDTTaggingRulesAreSetUp()
	loggingDone <- true
	return nil
}

func closeLogger(loggingDone chan bool, combinedLogger *keptnutils.CombinedLogger, ws *websocket.Conn) {
	<-loggingDone
	combinedLogger.Terminate()
	ws.Close()
}

package event_handler

import (
	"errors"
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
	KeptnHandler     *keptnv2.Keptn
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
	err := eh.configureMonitoring()
	if err != nil {
		eh.Logger.Error(err.Error())
	}
	return nil
}

func (eh *ConfigureMonitoringEventHandler) configureMonitoring() error {
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
	eh.KeptnHandler = keptnHandler

	var shipyard *keptnv2.Shipyard
	if e.Project != "" {
		shipyard, err = keptnHandler.GetShipyard()
		if err != nil {
			msg := fmt.Sprintf("failed to retrieve shipyard for project %s: %v", e.Project, err)
			return eh.handleError(e, msg)
		}
	}

	keptnEvent := adapter.NewConfigureMonitoringAdapter(*e, keptnHandler.KeptnContext, eh.Event.Source())

	dynatraceConfig, err := config.GetDynatraceConfig(keptnEvent, eh.Logger)
	if err != nil {
		msg := fmt.Sprintf("failed to load Dynatrace config: %v", err)
		return eh.handleError(e, msg)
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		msg := fmt.Sprintf("failed to load Dynatrace credentials: %v", err)
		return eh.handleError(e, msg)
	}
	dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

	configuredEntities, err := dtHelper.ConfigureMonitoring(e.Project, shipyard)
	if err != nil {
		return eh.handleError(e, err.Error())
	}

	eh.Logger.Info("Dynatrace Monitoring setup done")

	if err := eh.sendConfigureMonitoringFinishedEvent(e, keptnv2.StatusSucceeded, keptnv2.ResultPass, getConfigureMonitoringResultMessage(configuredEntities)); err != nil {
		eh.Logger.Error(err.Error())
	}
	return nil
}

func getConfigureMonitoringResultMessage(entities *lib.ConfiguredEntities) string {
	if entities == nil {
		return ""
	}
	msg := "Dynatrace monitoring setup done.\nThe following entities have been configured:\n\n"

	if entities.ManagementZonesEnabled && len(entities.ManagementZones) > 0 {
		msg = msg + "---Management Zones:--- \n"
		for _, mz := range entities.ManagementZones {
			if mz.Success {
				msg = msg + "  - " + mz.Name + ": Created successfully \n"
			} else {
				msg = msg + "  - " + mz.Name + ": Error: " + mz.Message + "\n"
			}
		}
		msg = msg + "\n\n"
	}

	if entities.TaggingRulesEnabled && len(entities.TaggingRules) > 0 {
		msg = msg + "---Automatic Tagging Rules:--- \n"
		for _, mz := range entities.TaggingRules {
			if mz.Success {
				msg = msg + "  - " + mz.Name + ": Created successfully \n"
			} else {
				msg = msg + "  - " + mz.Name + ": Error: " + mz.Message + "\n"
			}
		}
		msg = msg + "\n\n"
	}

	if entities.ProblemNotificationsEnabled {
		msg = msg + "---Problem Notification:--- \n"
		msg = msg + "  - " + entities.ProblemNotifications.Message
		msg = msg + "\n\n"
	}

	if entities.MetricEventsEnabled && len(entities.MetricEvents) > 0 {
		msg = msg + "---Metric Events:--- \n"
		for _, mz := range entities.MetricEvents {
			if mz.Success {
				msg = msg + "  - " + mz.Name + ": Created successfully \n"
			} else {
				msg = msg + "  - " + mz.Name + ": Error: " + mz.Message + "\n"
			}
		}
		msg = msg + "\n\n"
	}

	if entities.DashboardEnabled && entities.Dashboard.Message != "" {
		msg = msg + "---Dashboard:--- \n"
		msg = msg + "  - " + entities.Dashboard.Message
		msg = msg + "\n"
	}

	return msg
}

func (eh *ConfigureMonitoringEventHandler) handleError(e *keptn.ConfigureMonitoringEventData, msg string) error {
	eh.Logger.Error(msg)
	if err := eh.sendConfigureMonitoringFinishedEvent(e, keptnv2.StatusErrored, keptnv2.ResultFailed, msg); err != nil {
		eh.Logger.Error(err.Error())
	}
	return errors.New(msg)
}

func (eh *ConfigureMonitoringEventHandler) sendConfigureMonitoringFinishedEvent(configureMonitoringData *keptn.ConfigureMonitoringEventData, status keptnv2.StatusType, result keptnv2.ResultType, message string) error {

	cmFinishedEvent := &keptnv2.ConfigureMonitoringFinishedEventData{
		EventData: keptnv2.EventData{
			Project: configureMonitoringData.Project,
			Service: configureMonitoringData.Service,
			Status:  status,
			Result:  result,
			Message: message,
		},
	}

	keptnContext, _ := eh.Event.Context.GetExtension("shkeptncontext")

	event := cloudevents.NewEvent()
	event.SetSource("dynatrace-service")
	event.SetDataContentType(cloudevents.ApplicationJSON)
	event.SetType(keptnv2.GetFinishedEventType(keptnv2.ConfigureMonitoringTaskName))
	event.SetData(cloudevents.ApplicationJSON, cmFinishedEvent)
	event.SetExtension("shkeptncontext", keptnContext)
	event.SetExtension("triggeredid", eh.Event.Context.GetID())

	if err := eh.KeptnHandler.SendCloudEvent(event); err != nil {
		return fmt.Errorf("could not send %s event: %s", keptnv2.GetFinishedEventType(keptnv2.ConfigureMonitoringTaskName), err.Error())
	}

	return nil
}

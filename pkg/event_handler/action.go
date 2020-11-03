package event_handler

import (
	"errors"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"os"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"

	"github.com/mitchellh/mapstructure"

	"github.com/keptn-contrib/dynatrace-service/pkg/lib"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ActionHandler struct {
	Logger *keptncommon.Logger
	Event  cloudevents.Event
}

func (eh ActionHandler) HandleEvent() error {

	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptncommon.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not initialize Keptn handler: " + err.Error())
		return err
	}

	var comment string
	var pid string

	if eh.Event.Type() == keptn.ActionTriggeredEventType {
		actionTriggeredData := &keptn.ActionTriggeredEventData{}

		err = eh.Event.DataAs(actionTriggeredData)
		if err != nil {
			eh.Logger.Error("Cannot parse incoming event: " + err.Error())
			return err
		}

		if actionTriggeredData.Problem.PID == "" {
			eh.Logger.Error("Cannot send DT problem comment: No problem ID is included in the event.")
			return errors.New("cannot send DT problem comment: No problem ID is included in the event")
		}

		comment = "Keptn triggered action " + actionTriggeredData.Action.Action
		if actionTriggeredData.Action.Description != "" {
			comment = comment + ": " + actionTriggeredData.Action.Description
		}
		pid = actionTriggeredData.Problem.PID

		keptnEvent := adapter.NewActionTriggeredAdapter(*actionTriggeredData, keptnHandler.KeptnContext, eh.Event.Source())

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

		// https://github.com/keptn-contrib/dynatrace-service/issues/174
		// Additionall to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
		dtInfoEvent := createInfoEvent(keptnEvent, dynatraceConfig, eh.Logger)
		dtInfoEvent.Title = "Keptn Remediation Action Triggered"
		dtInfoEvent.Description = actionTriggeredData.Action.Action
		dtHelper.SendEvent(dtInfoEvent)

		// this is posting the Event on the problem as a comment
		err = dtHelper.SendProblemComment(pid, comment)
	} else if eh.Event.Type() == keptn.ActionFinishedEventType {
		actionFinishedData := &keptn.ActionFinishedEventData{}

		err = eh.Event.DataAs(actionFinishedData)
		if err != nil {
			eh.Logger.Error("Cannot parse incoming Event: " + err.Error())
			return err
		}

		keptnEvent := adapter.NewActionFinishedAdapter(*actionFinishedData, keptnHandler.KeptnContext, eh.Event.Source())

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

		eventHandler := keptnapi.NewEventHandler(os.Getenv("DATASTORE"))

		events, errObj := eventHandler.GetEvents(&keptnapi.EventFilter{
			Project:      keptnHandler.KeptnBase.Event.GetProject(),
			EventType:    keptn.ProblemOpenEventType,
			KeptnContext: keptnHandler.KeptnContext,
		})

		if errObj != nil {
			msg := "cannot send DT problem comment: Could not retrieve problem.open event for incoming event: " + *errObj.Message
			eh.Logger.Error(msg)
			return errors.New(msg)
		}

		if len(events) == 0 {
			msg := "cannot send DT problem comment: Could not retrieve problem.open event for incoming event: no events returned"
			eh.Logger.Error(msg)
			return errors.New(msg)
		}

		problemOpenEvent := &keptn.ProblemEventData{}
		err = mapstructure.Decode(events[0].Data, problemOpenEvent)

		if err != nil {
			msg := "could not decode problem.open event: " + err.Error()
			eh.Logger.Error(msg)
			return errors.New(msg)
		}

		if problemOpenEvent.PID == "" {
			eh.Logger.Error("Cannot send DT problem comment: No problem ID is included in the event.")
			return errors.New("cannot send DT problem comment: No problem ID is included in the event")
		}

		comment = "Keptn finished execution of action"

		pid = problemOpenEvent.PID

		// https://github.com/keptn-contrib/dynatrace-service/issues/174
		// Additionall to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
		if actionFinishedData.Action.Status == keptn.ActionStatusSucceeded {
			dtConfigEvent := createConfigurationEvent(keptnEvent, dynatraceConfig, eh.Logger)
			dtConfigEvent.Description = "Keptn Remediation Action Finished"
			dtConfigEvent.Configuration = "successful"
			dtHelper.SendEvent(dtConfigEvent)
		} else {
			dtInfoEvent := createInfoEvent(keptnEvent, dynatraceConfig, eh.Logger)
			dtInfoEvent.Title = "Keptn Remediation Action Finished"
			dtInfoEvent.Description = "error during execution"
			dtHelper.SendEvent(dtInfoEvent)
		}

		// this is posting the Event on the problem as a comment
		err = dtHelper.SendProblemComment(pid, comment)
	} else {
		return errors.New("invalid event type")
	}

	return nil
}

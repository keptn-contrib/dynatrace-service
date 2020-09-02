package event_handler

import (
	"errors"
	"os"

	"github.com/mitchellh/mapstructure"

	"github.com/keptn-contrib/dynatrace-service/pkg/lib"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ActionHandler struct {
	Logger *keptn.Logger
	Event  cloudevents.Event
}

func (eh ActionHandler) HandleEvent() error {

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not initialize Keptn handler: " + err.Error())
		return err
	}
	dynatraceHelper, err := lib.NewDynatraceHelper(keptnHandler)
	dynatraceHelper.Logger = eh.Logger
	if err != nil {
		eh.Logger.Error("could not initialize Dynatrace helper: " + err.Error())
		return err
	}

	keptnBase := &baseKeptnEvent{
		project: keptnHandler.KeptnBase.Project,
		stage:   keptnHandler.KeptnBase.Stage,
		service: keptnHandler.KeptnBase.Service,
	}

	dynatraceConfig, _ := getDynatraceConfig(keptnBase, eh.Logger)

	dtCreds := ""
	if dynatraceConfig != nil {
		dtCreds = dynatraceConfig.DtCreds
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

		// https://github.com/keptn-contrib/dynatrace-service/issues/174
		// Additionall to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
		dtInfoEvent := CreateInfoEvent(keptnBase, dynatraceConfig, eh.Logger)
		dtInfoEvent.Title = "Keptn Remediation Action Triggered"
		dtInfoEvent.Description = actionTriggeredData.Action.Action
		dynatraceHelper.SendEvent(dtInfoEvent, dtCreds)

	} else if eh.Event.Type() == keptn.ActionFinishedEventType {
		actionFinishedData := &keptn.ActionFinishedEventData{}

		err = eh.Event.DataAs(actionFinishedData)
		if err != nil {
			eh.Logger.Error("Cannot parse incoming event: " + err.Error())
			return err
		}

		eventHandler := keptnapi.NewEventHandler(os.Getenv("DATASTORE"))

		events, errObj := eventHandler.GetEvents(&keptnapi.EventFilter{
			Project:      keptnHandler.KeptnBase.Project,
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
			dtConfigEvent := CreateConfigurationEvent(keptnBase, dynatraceConfig, eh.Logger)
			dtConfigEvent.Description = "Keptn Remediation Action Finished"
			dtConfigEvent.Configuration = "successful"
			dynatraceHelper.SendEvent(dtConfigEvent, dtCreds)
		} else {
			dtInfoEvent := CreateInfoEvent(keptnBase, dynatraceConfig, eh.Logger)
			dtInfoEvent.Title = "Keptn Remediation Action Finished"
			dtInfoEvent.Description = "error during execution"
			dynatraceHelper.SendEvent(dtInfoEvent, dtCreds)
		}
	} else {
		return errors.New("invalid event type")
	}

	// this is posting the event on the problem as a comment
	err = dynatraceHelper.SendProblemComment(pid, comment, dtCreds)

	return nil
}

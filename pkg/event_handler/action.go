package event_handler

import (
	"errors"

	"github.com/keptn-contrib/dynatrace-service/pkg/lib"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
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
	} else if eh.Event.Type() == keptn.ActionFinishedEventType {
		actionFinishedData := &keptn.ActionFinishedEventData{}

		err = eh.Event.DataAs(actionFinishedData)
		if err != nil {
			eh.Logger.Error("Cannot parse incoming event: " + err.Error())
			return err
		}

		if actionFinishedData.Problem.PID == "" {
			eh.Logger.Error("Cannot send DT problem comment: No problem ID is included in the event.")
			return errors.New("cannot send DT problem comment: No problem ID is included in the event")
		}

		comment = "Keptn finished execution of action"

		pid = actionFinishedData.Problem.PID
	} else {
		return errors.New("invalid event type")
	}

	err = dynatraceHelper.SendProblemComment(pid, comment, dtCreds)

	return nil
}

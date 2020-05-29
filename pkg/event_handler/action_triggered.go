package event_handler

import (
	"errors"

	"github.com/keptn-contrib/dynatrace-service/pkg/lib"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ActionTriggeredHandler struct {
	Logger *keptn.Logger
	Event  cloudevents.Event
}

func (eh ActionTriggeredHandler) HandleEvent() error {

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

	actionTriggeredData := &keptn.ActionTriggeredEventData{}

	if actionTriggeredData.Problem.PID == "" {
		eh.Logger.Error("Cannot send DT problem comment: No problem ID is included in the event.")
		return errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	comment := "Keptn triggering action " + actionTriggeredData.Action.Action
	if actionTriggeredData.Action.Description != "" {
		comment = comment + ": " + actionTriggeredData.Action.Description
	}
	err = dynatraceHelper.SendProblemComment(actionTriggeredData.Problem.PID, comment, dtCreds)

	return nil
}

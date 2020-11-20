package event_handler

import (
	"errors"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"

	"github.com/keptn-contrib/dynatrace-service/pkg/lib"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ActionHandler struct {
	Logger *keptn.Logger
	Event  cloudevents.Event
}

/**
 * Retrieves Dynatrace Credential information
 */
func (eh ActionHandler) GetDynatraceCredentials(keptnEvent adapter.EventContentAdapter) (*config.DynatraceConfigFile, *credentials.DTCredentials, error) {
	dynatraceConfig, err := config.GetDynatraceConfig(keptnEvent, eh.Logger)
	if err != nil {
		eh.Logger.Error("failed to load Dynatrace config: " + err.Error())
		return nil, nil, err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		eh.Logger.Error("failed to load Dynatrace credentials: " + err.Error())
		return nil, nil, err
	}

	return dynatraceConfig, creds, nil
}

func (eh ActionHandler) HandleEvent() error {

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
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
		comment = fmt.Sprintf("[Keptn triggered action](%s) %s", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], actionTriggeredData.Action.Action)
		if actionTriggeredData.Action.Description != "" {
			comment = comment + ": " + actionTriggeredData.Action.Description
		}
		pid = actionTriggeredData.Problem.PID

		err = dtHelper.SendProblemComment(pid, comment)
	} else if eh.Event.Type() == keptn.ActionStartedEventType {
		actionStartedData := &keptn.ActionStartedEventData{}

		err = eh.Event.DataAs(actionStartedData)
		if err != nil {
			eh.Logger.Error("Cannot parse incoming Event: " + err.Error())
			return err
		}

		keptnEvent := adapter.NewActionStartedAdapter(*actionStartedData, keptnHandler.KeptnContext, eh.Event.Source())

		pid, err := common.FindProblemIDForEvent(keptnHandler, keptnEvent.GetLabels())
		if err != nil {
			eh.Logger.Error(err.Error())
			return err
		}

		// lets get our dynatrace credentials - if we have none - no need to continue
		/*dynatraceConfig*/
		_, creds, err := eh.GetDynatraceCredentials(keptnEvent)
		if err != nil {
			return err
		}

		// Create our DTHelper
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

		// Comment we push over
		comment = fmt.Sprintf("[Keptn remediation action](%s) started execution by: %s", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.Event.Source())

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

		// lets get our dynatrace credentials - if we have none - no need to continue
		dynatraceConfig, creds, err := eh.GetDynatraceCredentials(keptnEvent)
		if err != nil {
			return err
		}

		// lets find our dynatrace problem details for this remediaiton workflow
		pid, err := common.FindProblemIDForEvent(keptnHandler, keptnEvent.GetLabels())
		if err != nil {
			eh.Logger.Error(err.Error())
			return err
		}

		// Create our DTHelper
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

		// Comment text we want to push over
		comment = fmt.Sprintf("[Keptn finished execution](%s) of action by: %s\nResult: %s\nStatus: %s",
			keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL],
			eh.Event.Source(),
			actionFinishedData.Action.Result,
			actionFinishedData.Action.Status)

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

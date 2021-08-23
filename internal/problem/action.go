package problem

import (
	"errors"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type ActionHandler struct {
	Event          cloudevents.Event
	DTConfigGetter adapter.DynatraceConfigGetterInterface
}

/**
 * Retrieves Dynatrace Credential information
 */
func (eh ActionHandler) GetDynatraceCredentials(keptnEvent adapter.EventContentAdapter) (*config.DynatraceConfigFile, *credentials.DTCredentials, error) {
	dynatraceConfig, err := eh.DTConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace config")
		return nil, nil, err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace credentials")
		return nil, nil, err
	}

	return dynatraceConfig, creds, nil
}

func (eh ActionHandler) HandleEvent() error {

	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptncommon.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not initialize Keptn handler")
		return err
	}

	var comment string

	if eh.Event.Type() == keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName) {
		actionTriggeredData := &keptnv2.ActionTriggeredEventData{}

		err = eh.Event.DataAs(actionTriggeredData)
		if err != nil {
			log.WithError(err).Error("Cannot parse incoming event")
			return err
		}

		keptnEvent := adapter.NewActionTriggeredAdapter(*actionTriggeredData, keptnHandler.KeptnContext, eh.Event.Source())

		pid, err := common.FindProblemIDForEvent(keptnHandler, keptnEvent.GetLabels())
		if err != nil {
			log.WithError(err).Error("Could not find problem ID for event")
			return err
		}

		if pid == "" {
			log.Error("Cannot send DT problem comment: No problem ID is included in the event.")
			return errors.New("cannot send DT problem comment: No problem ID is included in the event")
		}

		comment = "Keptn triggered action " + actionTriggeredData.Action.Action
		if actionTriggeredData.Action.Description != "" {
			comment = comment + ": " + actionTriggeredData.Action.Description
		}

		dynatraceConfig, err := eh.DTConfigGetter.GetDynatraceConfig(keptnEvent)
		if err != nil {
			log.WithError(err).Error("Failed to load Dynatrace config")
			return err
		}

		creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
		if err != nil {
			log.WithError(err).Error("failed to load Dynatrace credentials")
			return err
		}
		dtHelper := dynatrace.NewDynatraceHelper(keptnHandler, creds)

		// https://github.com/keptn-contrib/dynatrace-service/issues/174
		// Additionall to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
		dtInfoEvent := event.CreateInfoEvent(keptnEvent, dynatraceConfig)
		dtInfoEvent.Title = "Keptn Remediation Action Triggered"
		dtInfoEvent.Description = actionTriggeredData.Action.Action
		dtHelper.SendEvent(dtInfoEvent)

		// this is posting the Event on the problem as a comment
		comment = fmt.Sprintf("[Keptn triggered action](%s) %s", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], actionTriggeredData.Action.Action)
		if actionTriggeredData.Action.Description != "" {
			comment = comment + ": " + actionTriggeredData.Action.Description
		}

		AddProblemComment(dtHelper, pid, comment)
	} else if eh.Event.Type() == keptnv2.GetStartedEventType(keptnv2.ActionTaskName) {
		actionStartedData := &keptnv2.ActionStartedEventData{}

		err = eh.Event.DataAs(actionStartedData)
		if err != nil {
			log.WithError(err).Error("Cannot parse incoming Event")
			return err
		}

		keptnEvent := adapter.NewActionStartedAdapter(*actionStartedData, keptnHandler.KeptnContext, eh.Event.Source())

		pid, err := common.FindProblemIDForEvent(keptnHandler, keptnEvent.GetLabels())
		if err != nil {
			log.WithError(err).Error("Could not find problem ID for event")
			return err
		}

		// lets get our dynatrace credentials - if we have none - no need to continue
		/*dynatraceConfig*/
		_, creds, err := eh.GetDynatraceCredentials(keptnEvent)
		if err != nil {
			return err
		}

		// Comment we push over
		comment = fmt.Sprintf("[Keptn remediation action](%s) started execution by: %s", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.Event.Source())

		AddProblemComment(dynatrace.NewDynatraceHelper(keptnHandler, creds), pid, comment)
	} else if eh.Event.Type() == keptnv2.GetFinishedEventType(keptnv2.ActionTaskName) {
		actionFinishedData := &keptnv2.ActionFinishedEventData{}

		err = eh.Event.DataAs(actionFinishedData)
		if err != nil {
			log.WithError(err).Error("Cannot parse incoming Event")
			return err
		}

		keptnEvent := adapter.NewActionFinishedAdapter(*actionFinishedData, keptnHandler.KeptnContext, eh.Event.Source())

		// lets get our dynatrace credentials - if we have none - no need to continue
		dynatraceConfig, creds, err := eh.GetDynatraceCredentials(keptnEvent)
		if err != nil {
			log.WithError(err).Error("Failed to load Dynatrace config")
			return err
		}

		// lets find our dynatrace problem details for this remediaiton workflow
		pid, err := common.FindProblemIDForEvent(keptnHandler, keptnEvent.GetLabels())
		if err != nil {
			log.WithError(err).Error("Could not find problem ID for event")
			return err
		}
		dtHelper := dynatrace.NewDynatraceHelper(keptnHandler, creds)

		// Comment text we want to push over
		comment = fmt.Sprintf("[Keptn finished execution](%s) of action by: %s\nResult: %s\nStatus: %s",
			keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL],
			eh.Event.Source(),
			actionFinishedData.Result,
			actionFinishedData.Status)

		// https://github.com/keptn-contrib/dynatrace-service/issues/174
		// Additionally to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
		if actionFinishedData.Status == keptnv2.StatusSucceeded {
			dtConfigEvent := event.CreateConfigurationEvent(keptnEvent, dynatraceConfig)
			dtConfigEvent.Description = "Keptn Remediation Action Finished"
			dtConfigEvent.Configuration = "successful"
			dtHelper.SendEvent(dtConfigEvent)
		} else {
			dtInfoEvent := event.CreateInfoEvent(keptnEvent, dynatraceConfig)
			dtInfoEvent.Title = "Keptn Remediation Action Finished"
			dtInfoEvent.Description = "error during execution"
			dtHelper.SendEvent(dtInfoEvent)
		}

		AddProblemComment(dtHelper, pid, comment)
	} else {
		return errors.New("invalid event type")
	}

	return nil
}

func AddProblemComment(dtHelper *dynatrace.DynatraceHelper, pid string, comment string) {
	log.WithField("comment", comment).Info("Adding problem comment")
	problemClient := dynatrace.NewProblemsClient(dtHelper)
	response, err := problemClient.SendProblemComment(pid, comment)
	if err != nil {
		log.WithError(err).Error("Error adding problem comment")
		return
	}

	log.WithField("response", response).Info("Received problem comment response")
}

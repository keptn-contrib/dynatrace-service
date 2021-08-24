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
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type ActionHandler struct {
	Event          cloudevents.Event
	DTConfigGetter config.DynatraceConfigGetterInterface
}

// Retrieves Dynatrace Credential information
func (eh ActionHandler) getDynatraceCredentials(keptnEvent adapter.EventContentAdapter) (*config.DynatraceConfigFile, *credentials.DTCredentials, error) {
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

	switch eh.Event.Type() {
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		return eh.handleActionTriggeredEvent()
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName):
		return eh.handleActionStartedEvent()
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName):
		return eh.handleActionFinishedEvent()
	default:
		return fmt.Errorf("invalid event type: %s", eh.Event.Type())
	}
}

func (eh *ActionHandler) handleActionTriggeredEvent() error {
	keptnEvent, err := NewActionTriggeredAdapterFromEvent(eh.Event)
	if err != nil {
		return err
	}

	pid, err := common.FindProblemIDForEvent(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	if pid == "" {
		log.Error("Cannot send DT problem comment: No problem ID is included in the event.")
		return errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	comment := "Keptn triggered action " + keptnEvent.GetAction()
	if keptnEvent.GetActionDescription() != "" {
		comment = comment + ": " + keptnEvent.GetActionDescription()
	}

	dynatraceConfig, creds, err := eh.getDynatraceCredentials(keptnEvent)
	if err != nil {
		return err
	}

	dtHelper := dynatrace.NewClient(creds)

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// In addition to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	dtInfoEvent := event.CreateInfoEvent(keptnEvent, dynatraceConfig)
	dtInfoEvent.Title = "Keptn Remediation Action Triggered"
	dtInfoEvent.Description = keptnEvent.GetAction()

	dynatrace.NewEventsClient(dtHelper).SendEvent(dtInfoEvent)

	// this is posting the Event on the problem as a comment
	comment = fmt.Sprintf("[Keptn triggered action](%s) %s", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], keptnEvent.GetAction())
	if keptnEvent.GetActionDescription() != "" {
		comment = comment + ": " + keptnEvent.GetActionDescription()
	}

	dynatrace.NewProblemsClient(dtHelper).AddProblemComment(pid, comment)

	return nil
}

func (eh *ActionHandler) handleActionStartedEvent() error {
	keptnEvent, err := NewActionStartedAdapterFromEvent(eh.Event)
	if err != nil {
		return err
	}

	pid, err := common.FindProblemIDForEvent(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}

	// lets get our dynatrace credentials - if we have none - no need to continue
	/*dynatraceConfig*/
	_, creds, err := eh.getDynatraceCredentials(keptnEvent)
	if err != nil {
		return err
	}

	// Comment we push over
	comment := fmt.Sprintf("[Keptn remediation action](%s) started execution by: %s", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], eh.Event.Source())

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewProblemsClient(dtHelper).AddProblemComment(pid, comment)

	return nil
}

func (eh *ActionHandler) handleActionFinishedEvent() error {
	keptnEvent, err := NewActionFinishedAdapterFromEvent(eh.Event)
	if err != nil {
		return err
	}

	// lets get our dynatrace credentials - if we have none - no need to continue
	dynatraceConfig, creds, err := eh.getDynatraceCredentials(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace config")
		return err
	}

	// lets find our dynatrace problem details for this remediaiton workflow
	pid, err := common.FindProblemIDForEvent(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Could not find problem ID for event")
		return err
	}
	dtHelper := dynatrace.NewClient(creds)

	// Comment text we want to push over
	comment := fmt.Sprintf("[Keptn finished execution](%s) of action by: %s\nResult: %s\nStatus: %s",
		keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL],
		eh.Event.Source(),
		keptnEvent.GetResult(),
		keptnEvent.GetStatus())

	// https://github.com/keptn-contrib/dynatrace-service/issues/174
	// Additionally to the problem comment, send Info and Configuration Change Event to the entities in Dynatrace to indicate that remediation actions have been executed
	if keptnEvent.GetStatus() == keptnv2.StatusSucceeded {
		dtConfigEvent := event.CreateConfigurationEvent(keptnEvent, dynatraceConfig)
		dtConfigEvent.Description = "Keptn Remediation Action Finished"
		dtConfigEvent.Configuration = "successful"

		dynatrace.NewEventsClient(dtHelper).SendEvent(dtConfigEvent)
	} else {
		dtInfoEvent := event.CreateInfoEvent(keptnEvent, dynatraceConfig)
		dtInfoEvent.Title = "Keptn Remediation Action Finished"
		dtInfoEvent.Description = "error during execution"

		dynatrace.NewEventsClient(dtHelper).SendEvent(dtInfoEvent)
	}

	dynatrace.NewProblemsClient(dtHelper).AddProblemComment(pid, comment)

	return nil
}

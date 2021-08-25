package problem

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
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

	keptnEvent, err := eh.getEventAdapter()
	if err != nil {
		return err
	}

	dynatraceConfig, dynatraceCredentials, err := eh.getDynatraceCredentials(keptnEvent)
	if err != nil {
		return err
	}

	client := dynatrace.NewClient(dynatraceCredentials)

	switch keptnEvent.(type) {
	case ActionTriggeredAdapter:
		handler := NewActionTriggeredEventHandler(keptnEvent.(*ActionTriggeredAdapter), client, dynatraceConfig)
		return handler.HandleEvent()
	case ActionStartedAdapter:
		handler := NewActionStartedEventHandler(keptnEvent.(*ActionStartedAdapter), client, eh.Event.Source())
		return handler.HandleEvent()
	case ActionFinishedAdapter:
		handler := NewActionFinishedEventHandler(keptnEvent.(*ActionFinishedAdapter), client, dynatraceConfig, eh.Event.Source())
		return handler.HandleEvent()
	default:
		return fmt.Errorf("invalid event type: %s", eh.Event.Type())
	}
}

func (eh ActionHandler) getEventAdapter() (adapter.EventContentAdapter, error) {

	switch eh.Event.Type() {
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		keptnEvent, err := NewActionTriggeredAdapterFromEvent(eh.Event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName):
		keptnEvent, err := NewActionStartedAdapterFromEvent(eh.Event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName):
		keptnEvent, err := NewActionFinishedAdapterFromEvent(eh.Event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	default:
		return nil, fmt.Errorf("invalid event type: %s", eh.Event.Type())
	}
}

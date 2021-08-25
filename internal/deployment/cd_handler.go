package deployment

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/config"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

type CDEventHandler struct {
	event          cloudevents.Event
	dtConfigGetter config.DynatraceConfigGetterInterface
}

func NewCDEventHandler(event cloudevents.Event, configGetter config.DynatraceConfigGetterInterface) CDEventHandler {
	return CDEventHandler{
		event:          event,
		dtConfigGetter: configGetter,
	}
}

// Retrieves Dynatrace Credential information
func (eh CDEventHandler) getDynatraceCredentials(keptnEvent adapter.EventContentAdapter) (*config.DynatraceConfigFile, *credentials.DTCredentials, error) {
	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent)
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

func (eh CDEventHandler) HandleEvent() error {

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
	case DeploymentFinishedAdapter:
		handler := NewDeploymentFinishedEventHandler(keptnEvent.(*DeploymentFinishedAdapter), client, dynatraceConfig)
		return handler.HandleEvent()
	case TestTriggeredAdapter:
		handler := NewTestTriggeredEventHandler(keptnEvent.(*TestTriggeredAdapter), client, dynatraceConfig)
		return handler.HandleEvent()
	case TestFinishedAdapter:
		handler := NewTestFinishedEventHandler(keptnEvent.(*TestFinishedAdapter), client, dynatraceConfig)
		return handler.HandleEvent()
	case EvaluationFinishedAdapter:
		handler := NewEvaluationFinishedEventHandler(keptnEvent.(*EvaluationFinishedAdapter), client, dynatraceConfig)
		return handler.HandleEvent()
	case ReleaseTriggeredAdapter:
		handler := NewReleaseTriggeredEventHandler(keptnEvent.(*ReleaseTriggeredAdapter), client, dynatraceConfig)
		return handler.HandleEvent()
	case nil:
		// in case 'getEventAdapter()' would not return a known type
		log.WithField("EventType", eh.event.Type()).Info("Ignoring event")
	default:
		return fmt.Errorf("invalid event type: %s", eh.event.Type())
	}

	return nil
}

func (eh CDEventHandler) getEventAdapter() (adapter.EventContentAdapter, error) {

	switch eh.event.Type() {
	case keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName):
		keptnEvent, err := NewDeploymentFinishedAdapterFromEvent(eh.event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetTriggeredEventType(keptnv2.TestTaskName):
		keptnEvent, err := NewTestTriggeredAdapterFromEvent(eh.event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.TestTaskName):
		keptnEvent, err := NewTestFinishedAdapterFromEvent(eh.event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName):
		keptnEvent, err := NewEvaluationFinishedAdapterFromEvent(eh.event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetTriggeredEventType(keptnv2.ReleaseTaskName):
		keptnEvent, err := NewReleaseTriggeredAdapterFromEvent(eh.event)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName):
		//do nothing
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid event type: %s", eh.event.Type())
	}
}

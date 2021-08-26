package event_handler

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/deployment"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"
	"github.com/keptn-contrib/dynatrace-service/internal/problem"
	"github.com/keptn-contrib/dynatrace-service/internal/sli"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DynatraceEventHandler interface {
	HandleEvent() error
}

// Retrieves Dynatrace Credential information
func getDynatraceCredentialsAndConfig(keptnEvent adapter.EventContentAdapter, dtConfigGetter config.DynatraceConfigGetterInterface) (*config.DynatraceConfigFile, *credentials.DTCredentials, error) {
	dynatraceConfig, err := dtConfigGetter.GetDynatraceConfig(keptnEvent)
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

func NewEventHandler(event cloudevents.Event) (DynatraceEventHandler, error) {
	log.WithField("eventType", event.Type()).Debug("Received event")
	dtConfigGetter := &config.DynatraceConfigGetter{}

	keptnEvent, err := getEventAdapter(event)
	if err != nil {
		log.WithError(err).Error("Could not create event adapter")
		return ErrorHandler{err: err}, nil
	}

	dynatraceConfig, dynatraceCredentials, err := getDynatraceCredentialsAndConfig(keptnEvent, dtConfigGetter)
	if err != nil {
		log.WithError(err).Error("Could not get dynatrace credentials and config")
		return ErrorHandler{err: err}, nil
	}

	client := dynatrace.NewClient(dynatraceCredentials)

	switch aType := keptnEvent.(type) {
	case monitoring.ConfigureMonitoringAdapter:
		return monitoring.NewConfigureMonitoringEventHandler(keptnEvent.(*monitoring.ConfigureMonitoringAdapter), client, event), nil
	case monitoring.ProjectCreateAdapter:
		return monitoring.NewCreateProjectEventHandler(keptnEvent.(*monitoring.ProjectCreateAdapter), client, event), nil
	case problem.ProblemAdapter:
		return problem.NewProblemEventHandler(keptnEvent.(*problem.ProblemAdapter)), nil
	case problem.ActionTriggeredAdapter:
		return problem.NewActionTriggeredEventHandler(keptnEvent.(*problem.ActionTriggeredAdapter), client, dynatraceConfig), nil
	case problem.ActionStartedAdapter:
		return problem.NewActionStartedEventHandler(keptnEvent.(*problem.ActionStartedAdapter), client, event.Source()), nil
	case problem.ActionFinishedAdapter:
		return problem.NewActionFinishedEventHandler(keptnEvent.(*problem.ActionFinishedAdapter), client, dynatraceConfig, event.Source()), nil
	case sli.GetSLITriggeredAdapter:
		// TODO 2021-08-25: consolidate dynatrace client and config file retrieval in GetSLIEventHandler
		return sli.NewGetSLITriggeredHandler(keptnEvent.(*sli.GetSLITriggeredAdapter), event), nil
	case deployment.DeploymentFinishedAdapter:
		return deployment.NewDeploymentFinishedEventHandler(keptnEvent.(*deployment.DeploymentFinishedAdapter), client, dynatraceConfig), nil
	case deployment.TestTriggeredAdapter:
		return deployment.NewTestTriggeredEventHandler(keptnEvent.(*deployment.TestTriggeredAdapter), client, dynatraceConfig), nil
	case deployment.TestFinishedAdapter:
		return deployment.NewTestFinishedEventHandler(keptnEvent.(*deployment.TestFinishedAdapter), client, dynatraceConfig), nil
	case deployment.EvaluationFinishedAdapter:
		return deployment.NewEvaluationFinishedEventHandler(keptnEvent.(*deployment.EvaluationFinishedAdapter), client, dynatraceConfig), nil
	case deployment.ReleaseTriggeredAdapter:
		return deployment.NewReleaseTriggeredEventHandler(keptnEvent.(*deployment.ReleaseTriggeredAdapter), client, dynatraceConfig), nil
	case nil:
		// in case 'getEventAdapter()' would return a type we would ignore
		return NoOpHandler{event: event}, nil
	default:
		return ErrorHandler{err: fmt.Errorf("this should not have happened, we are missing an implementation for: %T", aType)}, nil
	}
}

func getEventAdapter(e cloudevents.Event) (adapter.EventContentAdapter, error) {
	switch e.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		keptnEvent, err := monitoring.NewConfigureMonitoringAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName):
		keptnEvent, err := monitoring.NewProjectCreateAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnevents.ProblemEventType:
		keptnEvent, err := problem.NewProblemAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		keptnEvent, err := problem.NewActionTriggeredAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName):
		keptnEvent, err := problem.NewActionStartedAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName):
		keptnEvent, err := problem.NewActionFinishedAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetTriggeredEventType(keptnv2.GetSLITaskName):
		keptnEvent, err := sli.NewGetSLITriggeredAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName):
		keptnEvent, err := deployment.NewDeploymentFinishedAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetTriggeredEventType(keptnv2.TestTaskName):
		keptnEvent, err := deployment.NewTestTriggeredAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.TestTaskName):
		keptnEvent, err := deployment.NewTestFinishedAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName):
		keptnEvent, err := deployment.NewEvaluationFinishedAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetTriggeredEventType(keptnv2.ReleaseTaskName):
		keptnEvent, err := deployment.NewReleaseTriggeredAdapterFromEvent(e)
		if err != nil {
			return nil, err
		}
		return keptnEvent, nil
	case keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName):
		//do nothing, ignore the type, don't even log
		return nil, nil
	default:
		log.WithField("EventType", e.Type()).Debug("Ignoring event")
		return nil, nil
	}
}

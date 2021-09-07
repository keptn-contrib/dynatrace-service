package event_handler

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/deployment"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"
	"github.com/keptn-contrib/dynatrace-service/internal/problem"
	"github.com/keptn-contrib/dynatrace-service/internal/sli"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnapi "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DynatraceEventHandler interface {
	HandleEvent() error
}

// Retrieves Dynatrace Credential information
func getDynatraceCredentialsAndConfig(keptnEvent adapter.EventContentAdapter, dtConfigGetter config.DynatraceConfigGetterInterface) (*config.DynatraceConfigFile, *credentials.DTCredentials, string, error) {
	dynatraceConfig, err := dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace config")
		return nil, nil, "", err
	}

	cm, err := credentials.NewCredentialManager(nil)
	if err != nil {
		return nil, nil, "", err
	}

	// TODO 2021-09-01: remove temporary fallback behaviour later on
	var fallbackDecorator *credentials.CredentialManagerFallbackDecorator
	switch keptnEvent.(type) {
	case *sli.GetSLITriggeredAdapter:
		fallbackDecorator = credentials.NewCredentialManagerSLIServiceFallbackDecorator(cm, keptnEvent.GetProject())
	default:
		fallbackDecorator = credentials.NewCredentialManagerDefaultFallbackDecorator(cm)
	}

	creds, err := fallbackDecorator.GetDynatraceCredentials(dynatraceConfig.DtCreds)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace credentials")
		return nil, nil, "", err
	}

	return dynatraceConfig, creds, fallbackDecorator.GetSecretName(), nil
}

func NewEventHandler(event cloudevents.Event) (DynatraceEventHandler, error) {
	log.WithField("eventType", event.Type()).Debug("Received event")
	dtConfigGetter := config.NewDynatraceConfigGetter(keptn.NewResourceClient())

	keptnEvent, err := getEventAdapter(event)
	if err != nil {
		log.WithError(err).Error("Could not create event adapter")
		return ErrorHandler{err: err}, nil
	}

	// in case 'getEventAdapter()' would return a type we would ignore, handle it explicitly here
	if keptnEvent == nil {
		return NoOpHandler{}, nil
	}

	dynatraceConfig, dynatraceCredentials, secretName, err := getDynatraceCredentialsAndConfig(keptnEvent, dtConfigGetter)
	if err != nil {
		log.WithError(err).Error("Could not get dynatrace credentials and config")
		return ErrorHandler{err: err}, nil
	}

	dtClient := dynatrace.NewClient(dynatraceCredentials)
	kClient, err := keptnv2.NewKeptn(&event, keptnapi.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not get create Keptn client")
		return ErrorHandler{err: err}, nil
	}

	switch aType := keptnEvent.(type) {
	case *monitoring.ConfigureMonitoringAdapter:
		return monitoring.NewConfigureMonitoringEventHandler(keptnEvent.(*monitoring.ConfigureMonitoringAdapter), dtClient, keptn.NewClient(kClient)), nil
	case *monitoring.ProjectCreateFinishedAdapter:
		return monitoring.NewProjectCreateFinishedEventHandler(keptnEvent.(*monitoring.ProjectCreateFinishedAdapter), dtClient, keptn.NewClient(kClient)), nil
	case *problem.ProblemAdapter:
		return problem.NewProblemEventHandler(keptnEvent.(*problem.ProblemAdapter), keptn.NewClient(kClient)), nil
	case *problem.ActionTriggeredAdapter:
		return problem.NewActionTriggeredEventHandler(keptnEvent.(*problem.ActionTriggeredAdapter), dtClient, dynatraceConfig.AttachRules), nil
	case *problem.ActionStartedAdapter:
		return problem.NewActionStartedEventHandler(keptnEvent.(*problem.ActionStartedAdapter), dtClient), nil
	case *problem.ActionFinishedAdapter:
		return problem.NewActionFinishedEventHandler(keptnEvent.(*problem.ActionFinishedAdapter), dtClient, dynatraceConfig.AttachRules), nil
	case *sli.GetSLITriggeredAdapter:
		return sli.NewGetSLITriggeredHandler(keptnEvent.(*sli.GetSLITriggeredAdapter), dtClient, keptn.NewClient(kClient), secretName, dynatraceConfig.Dashboard), nil
	case *deployment.DeploymentFinishedAdapter:
		return deployment.NewDeploymentFinishedEventHandler(keptnEvent.(*deployment.DeploymentFinishedAdapter), dtClient, dynatraceConfig.AttachRules), nil
	case *deployment.TestTriggeredAdapter:
		return deployment.NewTestTriggeredEventHandler(keptnEvent.(*deployment.TestTriggeredAdapter), dtClient, dynatraceConfig.AttachRules), nil
	case *deployment.TestFinishedAdapter:
		return deployment.NewTestFinishedEventHandler(keptnEvent.(*deployment.TestFinishedAdapter), dtClient, dynatraceConfig.AttachRules), nil
	case *deployment.EvaluationFinishedAdapter:
		return deployment.NewEvaluationFinishedEventHandler(keptnEvent.(*deployment.EvaluationFinishedAdapter), dtClient, dynatraceConfig.AttachRules), nil
	case *deployment.ReleaseTriggeredAdapter:
		return deployment.NewReleaseTriggeredEventHandler(keptnEvent.(*deployment.ReleaseTriggeredAdapter), dtClient, dynatraceConfig.AttachRules), nil
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
		keptnEvent, err := monitoring.NewProjectCreateFinishedAdapterFromEvent(e)
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

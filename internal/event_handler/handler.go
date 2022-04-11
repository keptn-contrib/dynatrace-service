package event_handler

import (
	"errors"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	api "github.com/keptn/go-utils/pkg/api/utils"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/deployment"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"
	"github.com/keptn-contrib/dynatrace-service/internal/problem"
	"github.com/keptn-contrib/dynatrace-service/internal/sli"
)

// DynatraceEventHandler is the common interface for all event handlers.
type DynatraceEventHandler interface {
	HandleEvent() error
}

// NewEventHandler creates a new DynatraceEventHandler for the specified event or returns an error.
func NewEventHandler(event cloudevents.Event) (DynatraceEventHandler, error) {
	keptnAPISet, err := api.New(common.GetShipyardControllerURL())
	if err != nil {
		return nil, fmt.Errorf("could not create Keptn API set: %w", err)
	}

	eventHandler, err := getEventHandler(event, *keptnAPISet)
	if err != nil {
		err = fmt.Errorf("cannot handle event: %w", err)
		log.Error(err.Error())
		return NewErrorHandler(err, event, keptn.NewUniformClient(keptnAPISet.UniformV1())), nil
	}

	return eventHandler, nil
}

func getEventHandler(event cloudevents.Event, keptnAPISet api.APISet) (DynatraceEventHandler, error) {
	log.WithField("eventType", event.Type()).Debug("Received event")

	keptnEvent, err := getEventAdapter(event)
	if err != nil {
		return nil, fmt.Errorf("could not create event adapter: %w", err)
	}

	// in case 'getEventAdapter()' would return a type we would ignore, handle it explicitly here
	if keptnEvent == nil {
		return NoOpHandler{}, nil
	}

	if keptnEvent.GetProject() == "" {
		return nil, errors.New("event has no project")
	}

	dynatraceConfigGetter := config.NewDynatraceConfigGetter(keptn.NewResourceClient(keptn.NewConfigResourceClient(keptnAPISet.ResourcesV1())))
	dynatraceConfig, err := dynatraceConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		return nil, fmt.Errorf("could not get configuration: %w", err)
	}

	credentialsProvider, err := credentials.NewDefaultDynatraceK8sSecretReader()
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes secret reader: %w", err)
	}

	dynatraceCredentials, err := credentialsProvider.GetDynatraceCredentials(dynatraceConfig.DtCreds)
	if err != nil {
		return nil, fmt.Errorf("could not get Dynatrace credentials: %w", err)
	}

	dtClient := dynatrace.NewClient(dynatraceCredentials)

	kClient, err := keptn.NewDefaultClient(event)
	if err != nil {
		return nil, fmt.Errorf("could not create Keptn client: %w", err)
	}

	switch aType := keptnEvent.(type) {
	case *monitoring.ConfigureMonitoringAdapter:
		return monitoring.NewConfigureMonitoringEventHandler(keptnEvent.(*monitoring.ConfigureMonitoringAdapter), dtClient, kClient, keptn.NewResourceClient(keptn.NewConfigResourceClient(keptnAPISet.ResourcesV1())), keptn.NewServiceClient(keptnAPISet.ServicesV1(), keptnAPISet.APIV1())), nil
	case *problem.ProblemAdapter:
		return problem.NewProblemEventHandler(keptnEvent.(*problem.ProblemAdapter), kClient), nil
	case *problem.ActionTriggeredAdapter:
		return problem.NewActionTriggeredEventHandler(keptnEvent.(*problem.ActionTriggeredAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1()), dynatraceConfig.AttachRules), nil
	case *problem.ActionStartedAdapter:
		return problem.NewActionStartedEventHandler(keptnEvent.(*problem.ActionStartedAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1())), nil
	case *problem.ActionFinishedAdapter:
		return problem.NewActionFinishedEventHandler(keptnEvent.(*problem.ActionFinishedAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1()), dynatraceConfig.AttachRules), nil
	case *sli.GetSLITriggeredAdapter:
		return sli.NewGetSLITriggeredHandler(keptnEvent.(*sli.GetSLITriggeredAdapter), dtClient, kClient, keptn.NewResourceClient(keptn.NewConfigResourceClient(keptnAPISet.ResourcesV1())), dynatraceConfig.DtCreds, dynatraceConfig.Dashboard), nil
	case *deployment.DeploymentFinishedAdapter:
		return deployment.NewDeploymentFinishedEventHandler(keptnEvent.(*deployment.DeploymentFinishedAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1()), dynatraceConfig.AttachRules), nil
	case *deployment.TestTriggeredAdapter:
		return deployment.NewTestTriggeredEventHandler(keptnEvent.(*deployment.TestTriggeredAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1()), dynatraceConfig.AttachRules), nil
	case *deployment.TestFinishedAdapter:
		return deployment.NewTestFinishedEventHandler(keptnEvent.(*deployment.TestFinishedAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1()), dynatraceConfig.AttachRules), nil
	case *deployment.EvaluationFinishedAdapter:
		return deployment.NewEvaluationFinishedEventHandler(keptnEvent.(*deployment.EvaluationFinishedAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1()), dynatraceConfig.AttachRules), nil
	case *deployment.ReleaseTriggeredAdapter:
		return deployment.NewReleaseTriggeredEventHandler(keptnEvent.(*deployment.ReleaseTriggeredAdapter), dtClient, keptn.NewEventClient(keptnAPISet.EventsV1()), dynatraceConfig.AttachRules), nil
	default:
		return NewErrorHandler(fmt.Errorf("this should not have happened, we are missing an implementation for: %T", aType), event, keptn.NewUniformClient(keptnAPISet.UniformV1())), nil
	}
}

func getEventAdapter(e cloudevents.Event) (adapter.EventContentAdapter, error) {
	switch e.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return monitoring.NewConfigureMonitoringAdapterFromEvent(e)
	case keptnevents.ProblemEventType:
		return problem.NewProblemAdapterFromEvent(e)
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		return problem.NewActionTriggeredAdapterFromEvent(e)
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName):
		return problem.NewActionStartedAdapterFromEvent(e)
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName):
		return problem.NewActionFinishedAdapterFromEvent(e)
	case keptnv2.GetTriggeredEventType(keptnv2.GetSLITaskName):
		return sli.NewGetSLITriggeredAdapterFromEvent(e)
	case keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName):
		return deployment.NewDeploymentFinishedAdapterFromEvent(e)
	case keptnv2.GetTriggeredEventType(keptnv2.TestTaskName):
		return deployment.NewTestTriggeredAdapterFromEvent(e)
	case keptnv2.GetFinishedEventType(keptnv2.TestTaskName):
		return deployment.NewTestFinishedAdapterFromEvent(e)
	case keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName):
		return deployment.NewEvaluationFinishedAdapterFromEvent(e)
	case keptnv2.GetTriggeredEventType(keptnv2.ReleaseTaskName):
		return deployment.NewReleaseTriggeredAdapterFromEvent(e)
	case keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName):
		//do nothing, ignore the type, don't even log
		return nil, nil
	default:
		log.WithField("EventType", e.Type()).Debug("Ignoring event")
		return nil, nil
	}
}

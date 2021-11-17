package event_handler

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/deployment"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"
	"github.com/keptn-contrib/dynatrace-service/internal/problem"
	"github.com/keptn-contrib/dynatrace-service/internal/sli"
)

type DynatraceEventHandler interface {
	HandleEvent() error
}

// Retrieves Dynatrace Credential information
func getDynatraceCredentialsAndConfig(keptnEvent adapter.EventContentAdapter, dtConfigGetter config.DynatraceConfigGetterInterface) (*config.DynatraceConfigFile, *credentials.DynatraceCredentials, string, error) {
	dynatraceConfig, err := dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Warn("Failed to load Dynatrace config - will use a default one!")

		// TODO 2021-09-08: think about a better way of handling it on a use-case per use-case basis
		dynatraceConfig = &config.DynatraceConfigFile{
			SpecVersion: "0.1.0",
			DtCreds:     "dynatrace",
			Dashboard:   "",
			AttachRules: nil,
		}
	}

	credentialsProvider, err := credentials.NewDefaultDynatraceK8sSecretReader()
	if err != nil {
		return nil, nil, "", err
	}

	var credentialsProviderWithDefault = credentials.NewDynatraceCredentialsProviderWithDefault(credentialsProvider)
	creds, err := credentialsProviderWithDefault.GetDynatraceCredentials(dynatraceConfig.DtCreds)
	if err != nil {
		return nil, nil, "", err
	}

	return dynatraceConfig, creds, credentialsProviderWithDefault.GetSecretName(), nil
}

func NewEventHandler(event cloudevents.Event) DynatraceEventHandler {
	log.WithField("eventType", event.Type()).Debug("Received event")
	dtConfigGetter := config.NewDynatraceConfigGetter(keptn.NewDefaultResourceClient())

	keptnEvent, err := getEventAdapter(event)
	if err != nil {
		log.WithError(err).Error("Could not create event adapter")
		return NewErrorHandler(err, event)
	}

	// in case 'getEventAdapter()' would return a type we would ignore, handle it explicitly here
	if keptnEvent == nil {
		return NoOpHandler{}
	}

	dynatraceConfig, dynatraceCredentials, secretName, err := getDynatraceCredentialsAndConfig(keptnEvent, dtConfigGetter)
	if err != nil {
		log.WithError(err).Error("Could not get dynatrace credentials and config")
		return NewErrorHandler(err, event)
	}

	dtClient := dynatrace.NewClient(dynatraceCredentials)
	kClient, err := keptn.NewDefaultClient(event)
	if err != nil {
		log.WithError(err).Error("Could not get create Keptn client")
		return NewErrorHandler(err, event)
	}

	switch aType := keptnEvent.(type) {
	case *monitoring.ConfigureMonitoringAdapter:
		return monitoring.NewConfigureMonitoringEventHandler(keptnEvent.(*monitoring.ConfigureMonitoringAdapter), dtClient, kClient, keptn.NewDefaultResourceClient(), keptn.NewDefaultServiceClient())
	case *monitoring.ProjectCreateFinishedAdapter:
		return monitoring.NewProjectCreateFinishedEventHandler(keptnEvent.(*monitoring.ProjectCreateFinishedAdapter), dtClient, kClient, keptn.NewDefaultResourceClient(), keptn.NewDefaultServiceClient())
	case *problem.ProblemAdapter:
		return problem.NewProblemEventHandler(keptnEvent.(*problem.ProblemAdapter), kClient)
	case *problem.ActionTriggeredAdapter:
		return problem.NewActionTriggeredEventHandler(keptnEvent.(*problem.ActionTriggeredAdapter), dtClient, keptn.NewDefaultEventClient(), dynatraceConfig.AttachRules)
	case *problem.ActionStartedAdapter:
		return problem.NewActionStartedEventHandler(keptnEvent.(*problem.ActionStartedAdapter), dtClient, keptn.NewDefaultEventClient())
	case *problem.ActionFinishedAdapter:
		return problem.NewActionFinishedEventHandler(keptnEvent.(*problem.ActionFinishedAdapter), dtClient, keptn.NewDefaultEventClient(), dynatraceConfig.AttachRules)
	case *sli.GetSLITriggeredAdapter:
		return sli.NewGetSLITriggeredHandler(keptnEvent.(*sli.GetSLITriggeredAdapter), dtClient, kClient, keptn.NewDefaultResourceClient(), secretName, dynatraceConfig.Dashboard)
	case *deployment.DeploymentFinishedAdapter:
		return deployment.NewDeploymentFinishedEventHandler(keptnEvent.(*deployment.DeploymentFinishedAdapter), dtClient, keptn.NewDefaultEventClient(), dynatraceConfig.AttachRules)
	case *deployment.TestTriggeredAdapter:
		return deployment.NewTestTriggeredEventHandler(keptnEvent.(*deployment.TestTriggeredAdapter), dtClient, keptn.NewDefaultEventClient(), dynatraceConfig.AttachRules)
	case *deployment.TestFinishedAdapter:
		return deployment.NewTestFinishedEventHandler(keptnEvent.(*deployment.TestFinishedAdapter), dtClient, keptn.NewDefaultEventClient(), dynatraceConfig.AttachRules)
	case *deployment.EvaluationFinishedAdapter:
		return deployment.NewEvaluationFinishedEventHandler(keptnEvent.(*deployment.EvaluationFinishedAdapter), dtClient, keptn.NewDefaultEventClient(), dynatraceConfig.AttachRules)
	case *deployment.ReleaseTriggeredAdapter:
		return deployment.NewReleaseTriggeredEventHandler(keptnEvent.(*deployment.ReleaseTriggeredAdapter), dtClient, keptn.NewDefaultEventClient(), dynatraceConfig.AttachRules)
	default:
		return NewErrorHandler(fmt.Errorf("this should not have happened, we are missing an implementation for: %T", aType), event)
	}
}

func getEventAdapter(e cloudevents.Event) (adapter.EventContentAdapter, error) {
	switch e.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return monitoring.NewConfigureMonitoringAdapterFromEvent(e)
	case keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName):
		return monitoring.NewProjectCreateFinishedAdapterFromEvent(e)
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

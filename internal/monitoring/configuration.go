package monitoring

import (
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

// ConfiguredEntities contains information about the entities configures in Dynatrace
type ConfiguredEntities struct {
	TaggingRules         []ConfigResult
	ProblemNotifications *ConfigResult
	ManagementZones      []ConfigResult
	Dashboard            *ConfigResult
	MetricEvents         []ConfigResult
}

type ConfigResult struct {
	Name    string
	Success bool
	Message string
}

type Configuration struct {
	dtClient      dynatrace.ClientInterface
	kClient       keptn.ClientInterface
	sloReader     keptn.SLOReaderInterface
	serviceClient keptn.ServiceClientInterface
}

func NewConfiguration(dynatraceClient dynatrace.ClientInterface, keptnClient keptn.ClientInterface, sloReader keptn.SLOReaderInterface, serviceClient keptn.ServiceClientInterface) *Configuration {
	return &Configuration{
		dtClient:      dynatraceClient,
		kClient:       keptnClient,
		sloReader:     sloReader,
		serviceClient: serviceClient,
	}
}

// ConfigureMonitoring configures Dynatrace for a Keptn project
func (mc *Configuration) ConfigureMonitoring(project string, shipyard keptnv2.Shipyard) (*ConfiguredEntities, error) {

	configuredEntities := &ConfiguredEntities{}

	if env.IsTaggingRulesGenerationEnabled() {
		configuredEntities.TaggingRules = NewAutoTagCreation(mc.dtClient).Create()
	}

	if env.IsProblemNotificationsGenerationEnabled() {
		configuredEntities.ProblemNotifications = NewProblemNotificationCreation(mc.dtClient).Create(project)
	}

	if env.IsManagementZonesGenerationEnabled() {
		configuredEntities.ManagementZones = NewManagementZoneCreation(mc.dtClient).Create(project, shipyard)
	}

	if env.IsDashboardsGenerationEnabled() {
		configuredEntities.Dashboard = NewDashboardCreation(mc.dtClient).Create(project, shipyard)
	}

	if env.IsMetricEventsGenerationEnabled() {
		var metricEvents []ConfigResult
		for _, stage := range shipyard.Spec.Stages {
			metricEvents = append(metricEvents, mc.createMetricEventsForStage(project, stage)...)
		}
		configuredEntities.MetricEvents = metricEvents
	}

	return configuredEntities, nil
}

func (mc *Configuration) createMetricEventsForStage(project string, stage keptnv2.Stage) []ConfigResult {
	if isStageMissingRemediationSequence(stage) {
		return nil
	}

	serviceNames, err := mc.serviceClient.GetServiceNames(project, stage.Name)
	if err != nil {
		return []ConfigResult{{
			Success: false,
			Message: err.Error(),
		}}
	}

	var metricEvents []ConfigResult
	for _, serviceName := range serviceNames {
		metricEvents = append(
			metricEvents,
			NewMetricEventCreation(mc.dtClient, mc.kClient, mc.sloReader).Create(project, stage.Name, serviceName)...)
	}
	return metricEvents
}

func isStageMissingRemediationSequence(stage keptnv2.Stage) bool {
	for _, taskSequence := range stage.Sequences {
		if taskSequence.Name == "remediation" {
			return false
		}
	}
	return true
}

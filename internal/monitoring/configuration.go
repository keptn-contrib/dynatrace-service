package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// ConfiguredEntities contains information about the entities configures in Dynatrace
type ConfiguredEntities struct {
	TaggingRulesEnabled         bool
	TaggingRules                []ConfigResult
	ProblemNotificationsEnabled bool
	ProblemNotifications        ConfigResult
	ManagementZonesEnabled      bool
	ManagementZones             []ConfigResult
	DashboardEnabled            bool
	Dashboard                   ConfigResult
	MetricEventsEnabled         bool
	MetricEvents                []ConfigResult
}

type ConfigResult struct {
	Name    string
	Success bool
	Message string
}

type Configuration struct {
	dtClient       dynatrace.ClientInterface
	kClient        keptn.ClientInterface
	resourceClient keptn.ResourceClientInterface
	serviceClient  keptn.ServiceClientInterface
}

func NewConfiguration(dynatraceClient dynatrace.ClientInterface, keptnClient keptn.ClientInterface, resourceClient keptn.ResourceClientInterface, serviceClient keptn.ServiceClientInterface) *Configuration {
	return &Configuration{
		dtClient:       dynatraceClient,
		kClient:        keptnClient,
		resourceClient: resourceClient,
		serviceClient:  serviceClient,
	}
}

// ConfigureMonitoring configures Dynatrace for a Keptn project
func (mc *Configuration) ConfigureMonitoring(project string, shipyard *keptnv2.Shipyard) (*ConfiguredEntities, error) {

	configuredEntities := &ConfiguredEntities{
		TaggingRulesEnabled:         env.IsTaggingRulesGenerationEnabled(),
		TaggingRules:                NewAutoTagCreation(mc.dtClient).Create(),
		ProblemNotificationsEnabled: env.IsProblemNotificationsGenerationEnabled(),
		ProblemNotifications:        NewProblemNotificationCreation(mc.dtClient).Create(),
		ManagementZonesEnabled:      env.IsManagementZonesGenerationEnabled(),
		ManagementZones:             []ConfigResult{},
		DashboardEnabled:            env.IsDashboardsGenerationEnabled(),
		Dashboard:                   ConfigResult{},
		MetricEventsEnabled:         env.IsMetricEventsGenerationEnabled(),
		MetricEvents:                []ConfigResult{},
	}

	if project != "" && shipyard != nil {
		configuredEntities.ManagementZones = NewManagementZoneCreation(mc.dtClient).Create(project, *shipyard)
		configuredEntities.Dashboard = NewDashboardCreation(mc.dtClient).Create(project, *shipyard)

		var metricEvents []ConfigResult
		// try to create metric events - if one fails, don't fail the whole setup
		for _, stage := range shipyard.Spec.Stages {
			if shouldCreateMetricEvents(stage) {
				serviceNames, err := mc.serviceClient.GetServiceNames(project, stage.Name)
				if err != nil {
					return nil, err
				}
				for _, serviceName := range serviceNames {
					metricEvents = append(
						metricEvents,
						NewMetricEventCreation(mc.dtClient, mc.kClient, mc.resourceClient).Create(project, stage.Name, serviceName)...)
				}
			}
		}
		configuredEntities.MetricEvents = metricEvents
	}
	return configuredEntities, nil
}

// shouldCreateMetricEvents checks if a task sequence with the name 'remediation' is available - this would be the equivalent of remediation_strategy: automated of Keptn < 0.8.x
func shouldCreateMetricEvents(stage keptnv2.Stage) bool {
	for _, taskSequence := range stage.Sequences {
		if taskSequence.Name == "remediation" {
			return true
		}
	}
	return false
}

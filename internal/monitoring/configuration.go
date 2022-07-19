package monitoring

import (
	"context"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

// configuredEntities contains information about the entities configures in Dynatrace
type configuredEntities struct {
	TaggingRules         []configResult
	ProblemNotifications *configResult
	ManagementZones      []configResult
	Dashboard            *configResult
	MetricEvents         []configResult
}

type configResult struct {
	Name    string
	Success bool
	Message string
}

type configuration struct {
	dtClient          dynatrace.ClientInterface
	eventSenderClient keptn.EventSenderClientInterface
	sliAndSLOReader   keptn.SLIAndSLOReaderInterface
	serviceClient     keptn.ServiceClientInterface
}

func newConfiguration(dynatraceClient dynatrace.ClientInterface, eventSenderClient keptn.EventSenderClientInterface, sliAndSLOReader keptn.SLIAndSLOReaderInterface, serviceClient keptn.ServiceClientInterface) *configuration {
	return &configuration{
		dtClient:          dynatraceClient,
		eventSenderClient: eventSenderClient,
		sliAndSLOReader:   sliAndSLOReader,
		serviceClient:     serviceClient,
	}
}

// configureMonitoring configures Dynatrace for a Keptn project
func (mc *configuration) configureMonitoring(ctx context.Context, project string, shipyard keptnv2.Shipyard) (*configuredEntities, error) {

	configuredEntities := &configuredEntities{}

	if env.IsTaggingRulesGenerationEnabled() {
		configuredEntities.TaggingRules = newAutoTagCreation(mc.dtClient).create(ctx)
	}

	if env.IsProblemNotificationsGenerationEnabled() {
		configuredEntities.ProblemNotifications = newProblemNotificationCreation(mc.dtClient).create(ctx, project)
	}

	if env.IsManagementZonesGenerationEnabled() {
		configuredEntities.ManagementZones = newManagementZoneCreation(mc.dtClient).create(ctx, project, shipyard)
	}

	if env.IsDashboardsGenerationEnabled() {
		configuredEntities.Dashboard = newDashboardCreation(mc.dtClient).create(ctx, project, shipyard)
	}

	if env.IsMetricEventsGenerationEnabled() {
		var metricEvents []configResult
		for _, stage := range shipyard.Spec.Stages {
			metricEvents = append(metricEvents, mc.createMetricEventsForStage(ctx, project, stage)...)
		}
		configuredEntities.MetricEvents = metricEvents
	}

	return configuredEntities, nil
}

func (mc *configuration) createMetricEventsForStage(ctx context.Context, project string, stage keptnv2.Stage) []configResult {
	if isStageMissingRemediationSequence(stage) {
		return nil
	}

	serviceNames, err := mc.serviceClient.GetServiceNames(ctx, project, stage.Name)
	if err != nil {
		return []configResult{{
			Success: false,
			Message: err.Error(),
		}}
	}

	var metricEvents []configResult
	for _, serviceName := range serviceNames {
		metricEvents = append(
			metricEvents,
			newMetricEventCreation(mc.dtClient, mc.eventSenderClient, mc.sliAndSLOReader).create(ctx, project, stage.Name, serviceName)...)
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

package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type ManagementZoneCreation struct {
	client *dynatrace.DynatraceHelper
}

func NewManagementZoneCreation(client *dynatrace.DynatraceHelper) *ManagementZoneCreation {
	return &ManagementZoneCreation{
		client: client,
	}
}

// Create creates a new management zone for the project
func (mzc *ManagementZoneCreation) Create(project string, shipyard keptnv2.Shipyard) []dynatrace.ConfigResult {
	var managementZones []dynatrace.ConfigResult
	if !lib.IsManagementZonesGenerationEnabled() {
		return managementZones
	}

	// get existing management zones
	managementZoneClient := dynatrace.NewManagementZonesClient(mzc.client)
	managementZoneNames, err := managementZoneClient.GetAll()
	if err != nil {
		// continue
		log.WithError(err).Error("Could not retrieve management zones")
	}

	managementZoneResult := getOrCreateManagementZone(
		managementZoneClient,
		GetManagementZoneNameForProject(project),
		func() *dynatrace.ManagementZone {
			return createManagementZoneForProject(project)
		},
		managementZoneNames)
	managementZones = append(managementZones, managementZoneResult)

	for _, stage := range shipyard.Spec.Stages {
		managementZone := getOrCreateManagementZone(
			managementZoneClient,
			GetManagementZoneNameForProjectAndStage(project, stage.Name),
			func() *dynatrace.ManagementZone {
				return createManagementZoneForStage(project, stage.Name)
			},
			managementZoneNames)
		managementZones = append(managementZones, managementZone)
	}

	return managementZones
}

func getOrCreateManagementZone(
	managementZoneClient *dynatrace.ManagementZonesClient,
	managementZoneName string,
	managementZoneFunc func() *dynatrace.ManagementZone,
	managementZoneNames *dynatrace.ManagementZones) dynatrace.ConfigResult {
	if managementZoneNames != nil && managementZoneNames.Contains(managementZoneName) {
		return dynatrace.ConfigResult{
			Name:    managementZoneName,
			Success: true,
			Message: "Management Zone '" + managementZoneName + "' was already available in your Tenant",
		}
	}

	_, err := managementZoneClient.Create(managementZoneFunc())
	if err != nil {
		log.WithError(err).Error("Failed to create management zone")
		return dynatrace.ConfigResult{
			Name:    managementZoneName,
			Success: false,
			Message: "failed to create management zone: " + err.Error(),
		}
	}

	return dynatrace.ConfigResult{
		Name:    managementZoneName,
		Success: true,
	}
}

func GetManagementZoneNameForProjectAndStage(project string, stage string) string {
	return GetManagementZoneNameForProject(project) + " " + stage
}

func GetManagementZoneNameForProject(project string) string {
	return "Keptn: " + project
}

func createManagementZoneForProject(project string) *dynatrace.ManagementZone {
	managementZone := &dynatrace.ManagementZone{
		Name: GetManagementZoneNameForProject(project),
		Rules: []dynatrace.MZRules{
			{
				Type:             dynatrace.ServiceEntityType,
				Enabled:          true,
				PropagationTypes: []string{},
				Conditions: []dynatrace.MZConditions{
					createManagementZoneConditions(dynatrace.KeptnProject, project),
				},
			},
		},
	}

	return managementZone
}

func createManagementZoneForStage(project string, stage string) *dynatrace.ManagementZone {
	managementZone := &dynatrace.ManagementZone{
		Name: GetManagementZoneNameForProjectAndStage(project, stage),
		Rules: []dynatrace.MZRules{
			{
				Type:             dynatrace.ServiceEntityType,
				Enabled:          true,
				PropagationTypes: []string{},
				Conditions: []dynatrace.MZConditions{
					createManagementZoneConditions(dynatrace.KeptnProject, project),
					createManagementZoneConditions(dynatrace.KeptnStage, stage),
				},
			},
		},
	}

	return managementZone
}

func createManagementZoneConditions(key string, value string) dynatrace.MZConditions {
	return dynatrace.MZConditions{
		Key: dynatrace.MZKey{
			Attribute: "SERVICE_TAGS",
		},
		ComparisonInfo: dynatrace.MZComparisonInfo{
			Type:     "TAG",
			Operator: "EQUALS",
			Value: dynatrace.MZValue{
				Context: "CONTEXTLESS",
				Key:     key,
				Value:   value,
			},
			Negate: false,
		},
	}
}

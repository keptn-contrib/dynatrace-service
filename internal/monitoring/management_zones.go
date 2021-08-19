package monitoring

import (
	"encoding/json"
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

// CreateFor creates a new management zone for the project
func (mzc *ManagementZoneCreation) CreateFor(project string, shipyard keptnv2.Shipyard) []dynatrace.ConfigResult {
	var managementZones []dynatrace.ConfigResult
	if !lib.IsManagementZonesGenerationEnabled() {
		return managementZones
	}
	// get existing management zones
	mzs := mzc.getManagementZones()

	found := false
	for _, mz := range mzs {
		if mz.Name == "Keptn: "+project {
			found = true
		}
	}

	if !found {
		managementZone := dynatrace.CreateManagementZoneForProject(project)
		mzPayload, err := json.Marshal(managementZone)
		if err == nil {
			_, err := mzc.client.SendDynatraceAPIRequest("/api/config/v1/managementZones", "POST", mzPayload)
			if err != nil {
				// Error occurred but continue

				log.WithError(err).Error("Failed to create management zone")

				managementZones = append(
					managementZones,
					dynatrace.ConfigResult{
						Name:    "Keptn: " + project,
						Success: false,
						Message: "failed to create management zone: " + err.Error(),
					})
			} else {
				managementZones = append(
					managementZones,
					dynatrace.ConfigResult{
						Name:    "Keptn: " + project,
						Success: true,
					})
			}
		} else {
			// Error occurred but continue
			log.WithError(err).Warn("Failed to marshal management zone for project")
		}
	} else {
		managementZones = append(
			managementZones,
			dynatrace.ConfigResult{
				Name:    "Keptn: " + project,
				Success: true,
				Message: "Management Zone 'Keptn:" + project + "' was already available in your Tenant",
			})
	}

	for _, stage := range shipyard.Spec.Stages {
		found := false
		for _, mz := range mzs {
			if mz.Name == getManagementZoneNameForStage(project, stage.Name) {
				found = true
			}
		}

		if !found {
			managementZone := dynatrace.CreateManagementZoneForStage(project, stage.Name)
			mzPayload, err := json.Marshal(managementZone)
			if err == nil {
				_, err = mzc.client.SendDynatraceAPIRequest("/api/config/v1/managementZones", "POST", mzPayload)

				if err != nil {
					log.WithError(err).Error("Could not create management zone")
					managementZones = append(
						managementZones,
						dynatrace.ConfigResult{
							Name:    managementZone.Name,
							Success: false,
							Message: "Could not create management zone: " + err.Error(),
						})
				} else {
					managementZones = append(
						managementZones,
						dynatrace.ConfigResult{
							Name:    managementZone.Name,
							Success: true,
						})
				}
			} else {
				log.WithError(err).Warn("Failed to marshal management zone for stage")
			}
		} else {
			managementZones = append(
				managementZones,
				dynatrace.ConfigResult{
					Name:    "Keptn: " + project + " " + stage.Name,
					Success: true,
					Message: "Management Zone 'Keptn:" + project + " " + stage.Name + "' was already available in your Tenant",
				})
		}
	}

	return managementZones
}

func getManagementZoneNameForStage(project string, stage string) string {
	return "Keptn: " + project + " " + stage
}

func (mzc *ManagementZoneCreation) getManagementZones() []dynatrace.Values {
	response, err := mzc.client.SendDynatraceAPIRequest("/api/config/v1/managementZones", "GET", nil)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve management zones")
		return nil
	}
	mzs := &dynatrace.DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), mzs)
	if err != nil {
		log.WithError(err).Error("Failed to parse management zones list")
		return nil
	}
	return mzs.Values
}

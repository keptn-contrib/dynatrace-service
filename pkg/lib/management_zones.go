package lib

import (
	"encoding/json"
	"fmt"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// CreateManagementZones creates a new management zone for the project
func (dt *DynatraceHelper) CreateManagementZones(project string, shipyard keptnv2.Shipyard) {
	if !IsManagementZonesGenerationEnabled() {
		return
	}
	// get existing management zones
	mzs := dt.getManagementZones()

	found := false
	for _, mz := range mzs {
		if mz.Name == "Keptn: "+project {
			found = true
		}
	}

	if !found {
		managementZone := CreateManagementZoneForProject(project)
		mzPayload, err := json.Marshal(managementZone)
		if err == nil {
			_, err := dt.sendDynatraceAPIRequest("/api/config/v1/managementZones", "POST", mzPayload)

			if err != nil {
				// Error occurred but continue
				msg := "failed to create management zone: " + err.Error()
				dt.Logger.Error(msg)

				dt.configuredEntities.ManagementZones = append(dt.configuredEntities.ManagementZones, ConfigResult{
					Name:    "Keptn: " + project,
					Success: false,
					Message: msg,
				})
			} else {
				dt.configuredEntities.ManagementZones = append(dt.configuredEntities.ManagementZones, ConfigResult{
					Name:    "Keptn: " + project,
					Success: true,
				})
			}
		} else {
			// Error occurred but continue
			fmt.Errorf("failed to marshal management zone: %v", err)
		}
	}

	for _, stage := range shipyard.Spec.Stages {
		found := false
		for _, mz := range mzs {
			if mz.Name == getManagementZoneNameForStage(project, stage.Name) {
				found = true
			}
		}

		if !found {
			managementZone := CreateManagementZoneForStage(project, stage.Name)
			mzPayload, _ := json.Marshal(managementZone)
			_, err := dt.sendDynatraceAPIRequest("/api/config/v1/managementZones", "POST", mzPayload)
			if err != nil {
				msg := "Could not create management zone: " + err.Error()
				dt.Logger.Error(msg)
				dt.configuredEntities.ManagementZones = append(dt.configuredEntities.ManagementZones, ConfigResult{
					Name:    managementZone.Name,
					Success: false,
					Message: msg,
				})
			} else {
				dt.configuredEntities.ManagementZones = append(dt.configuredEntities.ManagementZones, ConfigResult{
					Name:    managementZone.Name,
					Success: true,
				})
			}
		}
	}

	return
}

func getManagementZoneNameForStage(project string, stage string) string {
	return "Keptn: " + project + " " + stage
}

func (dt *DynatraceHelper) getManagementZones() []Values {
	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/managementZones", "GET", nil)
	if err != nil {
		dt.Logger.Error("failed not retrieve management zones: " + err.Error())
		return nil
	}
	mzs := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), mzs)
	if err != nil {
		dt.Logger.Error("failed not parse management zones list: " + err.Error())
		return nil
	}
	return mzs.Values
}

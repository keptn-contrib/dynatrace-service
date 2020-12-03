package lib

import (
	"encoding/json"
	"fmt"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// CreateDashboard creates a new dashboard for the provided project
func (dt *DynatraceHelper) CreateDashboard(project string, shipyard keptnv2.Shipyard) {
	if !IsDashboardsGenerationEnabled() {
		return
	}

	// first, check if dashboard for this project already exists and delete that
	err := dt.DeleteExistingDashboard(project)
	if err != nil {
		msg := "Could not delete existing dashboard: " + err.Error()
		dt.Logger.Error(msg)
		dt.configuredEntities.Dashboard.Success = false
		dt.configuredEntities.Dashboard.Message = msg
		return
	}

	dt.Logger.Info("Creating Dashboard for project " + project)
	dashboard := createDynatraceDashboard(project, shipyard)
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal Dynatrace dashboards: %v", err)
		dt.Logger.Error(msg)
		dt.configuredEntities.Dashboard.Success = false
		dt.configuredEntities.Dashboard.Message = msg
		return
	}

	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/dashboards", "POST", dashboardPayload)
	if err != nil {
		msg := fmt.Sprintf("failed to create Dynatrace dashboards: %v", err)
		dt.Logger.Error(msg)
		dt.configuredEntities.Dashboard.Success = false
		dt.configuredEntities.Dashboard.Message = msg
		return
	}
	msg := "Dynatrace dashboard created successfully. You can view it here: https://" + dt.DynatraceCreds.Tenant + "/#dashboards"
	dt.Logger.Info(msg)
	dt.configuredEntities.Dashboard.Success = false
	dt.configuredEntities.Dashboard.Message = msg
	return
}

const dashboardNameSuffix = "@keptn: Digital Delivery & Operations Dashboard"

// DeleteExistingDashboard deletes an existing dashboard for the provided project
func (dt *DynatraceHelper) DeleteExistingDashboard(project string) error {
	res, err := dt.sendDynatraceAPIRequest("/api/config/v1/dashboards", "GET", nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve list of existing Dynatrace dashboards: %v", err)
	}

	dtDashboardsResponse := &DTDashboardsResponse{}
	err = json.Unmarshal([]byte(res), dtDashboardsResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal list of existing Dynatrace dashboards: %v", err)
	}

	for _, dashboardItem := range dtDashboardsResponse.Dashboards {
		if dashboardItem.Name == project+dashboardNameSuffix {
			res, err = dt.sendDynatraceAPIRequest("/api/config/v1/dashboards/"+dashboardItem.ID, "DELETE", nil)
			if err != nil {
				return fmt.Errorf("Could not delete dashboard for project %s: %v", project, err)
			}
		}
	}
	return nil
}

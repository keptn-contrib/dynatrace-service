package lib

import (
	"encoding/json"
	"fmt"

	keptn "github.com/keptn/go-utils/pkg/lib"
)

func (dt *DynatraceHelper) CreateDashboard(project string, shipyard keptn.Shipyard) {
	if !GetGenerateDashboardsConfig() {
		return
	}

	// first, check if dashboard for this project already exists and delete that
	err := dt.DeleteExistingDashboard(project)
	if err != nil {
		dt.Logger.Error(err.Error())
		return
	}

	dt.Logger.Info("Creating Dashboard for project " + project)
	dashboard := CreateDynatraceDashboard(project, shipyard)
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		dt.Logger.Error(fmt.Sprintf("failed to unmarshal Dynatrace dashboards: %v", err))
		return
	}

	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/dashboards", "POST", dashboardPayload)
	if err != nil {
		dt.Logger.Error(fmt.Sprintf("failed to create Dynatrace dashboards: %v", err))
		return
	}
	dt.Logger.Info("Dynatrace dashboard created successfully. You can view it here: https://" + dt.DynatraceCreds.Tenant + "/#dashboards")
	return
}

const dashboardNameSuffix = "@keptn: Digital Delivery & Operations Dashboard"

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

package lib

import (
	"encoding/json"

	keptn "github.com/keptn/go-utils/pkg/lib"
)

func (dt *DynatraceHelper) CreateDashboard(project string, shipyard keptn.Shipyard, services []string) error {
	// first, check if dashboard for this project already exists and delete that
	err := dt.DeleteExistingDashboard(project)
	if err != nil {
		return err
	}

	dt.Logger.Info("Creating Dashboard for project " + project)
	dashboard, err := CreateDynatraceDashboard(project, shipyard, services)
	if err != nil {
		return err
	}

	dashboardPayload, _ := json.Marshal(dashboard)

	_, err = dt.sendDynatraceAPIRequest("", "/api/config/v1/dashboards", "POST", string(dashboardPayload))

	if err != nil {
		return err
	}
	dt.Logger.Info("Dynatrace dashboard created successfully. You can view it here: https://" + dt.DynatraceCreds.Tenant + "/#dashboards")
	return nil
}

func (dt *DynatraceHelper) DeleteExistingDashboard(project string) error {
	res, err := dt.sendDynatraceAPIRequest("", "/api/config/v1/dashboards", "GET", "")
	if err != nil {
		dt.Logger.Error("Could not retrieve list of existing Dynatrace dashboards: " + err.Error())
		return err
	}

	dtDashboardsResponse := &DTDashboardsResponse{}
	err = json.Unmarshal([]byte(res), dtDashboardsResponse)

	if err != nil {
		dt.Logger.Error("Could not parse list of existing Dynatrace dashboards: " + err.Error())
		return err
	}

	for _, dashboardItem := range dtDashboardsResponse.Dashboards {
		if dashboardItem.Name == project+"@keptn: Digital Delivery & Operations Dashboard" {
			res, err = dt.sendDynatraceAPIRequest("", "/api/config/v1/dashboards/"+dashboardItem.ID, "DELETE", "")
			if err != nil {
				dt.Logger.Error("Could not delete previous dashboard for project " + project + ": " + err.Error())
				return err
			}
		}
	}
	return nil
}

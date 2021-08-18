package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

// CreateDashboard creates a new dashboard for the provided project
func (dt *DynatraceHelper) CreateDashboard(project string, shipyard keptnv2.Shipyard) {
	if !IsDashboardsGenerationEnabled() {
		return
	}

	// first, check if dashboard for this project already exists and delete that
	err := dt.DeleteExistingDashboard(project)
	if err != nil {
		log.WithError(err).Error("Could not delete existing dashboard")
		dt.configuredEntities.Dashboard.Success = false
		dt.configuredEntities.Dashboard.Message = "Could not delete existing dashboard: " + err.Error()
		return
	}

	log.WithField("project", project).Info("Creating Dashboard for project")
	dashboard := createDynatraceDashboard(project, shipyard)
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal Dynatrace dashboards")
		dt.configuredEntities.Dashboard.Success = false
		dt.configuredEntities.Dashboard.Message = fmt.Sprintf("failed to unmarshal Dynatrace dashboards: %v", err)
		return
	}

	_, err = dt.SendDynatraceAPIRequest("/api/config/v1/dashboards", "POST", dashboardPayload)
	if err != nil {
		log.WithError(err).Error("Failed to create Dynatrace dashboards")
		dt.configuredEntities.Dashboard.Success = false
		dt.configuredEntities.Dashboard.Message = fmt.Sprintf("failed to create Dynatrace dashboards: %v", err)
		return
	}
	log.WithField("dashboardUrl", "https://"+dt.DynatraceCreds.Tenant+"/#dashboards").Info("Dynatrace dashboard created successfully")
	dt.configuredEntities.Dashboard.Success = false
	dt.configuredEntities.Dashboard.Message = "Dynatrace dashboard created successfully. You can view it here: https://" + dt.DynatraceCreds.Tenant + "/#dashboards"
	return
}

const dashboardNameSuffix = "@keptn: Digital Delivery & Operations Dashboard"

// DeleteExistingDashboard deletes an existing dashboard for the provided project
func (dt *DynatraceHelper) DeleteExistingDashboard(project string) error {
	res, err := dt.SendDynatraceAPIRequest("/api/config/v1/dashboards", "GET", nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve list of existing Dynatrace dashboards: %v", err)
	}

	dtDashboardsResponse := &DTDashboardsResponse{}
	err = json.Unmarshal([]byte(res), dtDashboardsResponse)
	if err != nil {
		err = checkForUnexpectedHTMLResponseError(err)
		return fmt.Errorf("failed to unmarshal list of existing Dynatrace dashboards: %v", err)
	}

	for _, dashboardItem := range dtDashboardsResponse.Dashboards {
		if dashboardItem.Name == project+dashboardNameSuffix {
			res, err = dt.SendDynatraceAPIRequest("/api/config/v1/dashboards/"+dashboardItem.ID, "DELETE", nil)
			if err != nil {
				return fmt.Errorf("Could not delete dashboard for project %s: %v", project, err)
			}
		}
	}
	return nil
}

func checkForUnexpectedHTMLResponseError(err error) error {
	// in some cases, e.g. when the DT API has a problem, or the request URL is malformed, we do get a 200 response coded, but with an HTML error page instead of JSON
	// this function checks for the resulting error in that case and generates an error message that is more user friendly
	if strings.Contains(err.Error(), "invalid character '<'") {
		err = errors.New("received invalid response from Dynatrace API")
	}
	return err
}

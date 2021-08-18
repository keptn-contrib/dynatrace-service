package monitoring

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	"strings"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

const dashboardNameSuffix = "@keptn: Digital Delivery & Operations Dashboard"

type DashboardCreation struct {
	client *lib.DynatraceHelper
}

func NewDashboardCreation(client *lib.DynatraceHelper) *DashboardCreation {
	return &DashboardCreation{
		client: client,
	}
}

// CreateFor creates a new dashboard for the provided project
func (dc *DashboardCreation) CreateFor(project string, shipyard keptnv2.Shipyard) lib.ConfigResult {
	if !lib.IsDashboardsGenerationEnabled() {
		return lib.ConfigResult{}
	}

	// first, check if dashboard for this project already exists and delete that
	err := dc.deleteExistingDashboard(project)
	if err != nil {
		log.WithError(err).Error("Could not delete existing dashboard")
		return lib.ConfigResult{
			Success: false,
			Message: "Could not delete existing dashboard: " + err.Error(),
		}
	}

	log.WithField("project", project).Info("Creating Dashboard for project")
	dashboard := lib.CreateDynatraceDashboard(project, shipyard, dashboardNameSuffix)
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal Dynatrace dashboards")
		return lib.ConfigResult{
			Success: false,
			Message: fmt.Sprintf("failed to unmarshal Dynatrace dashboards: %v", err),
		}
	}

	_, err = dc.client.SendDynatraceAPIRequest("/api/config/v1/dashboards", "POST", dashboardPayload)
	if err != nil {
		log.WithError(err).Error("Failed to create Dynatrace dashboards")
		return lib.ConfigResult{
			Success: false,
			Message: fmt.Sprintf("failed to create Dynatrace dashboards: %v", err),
		}
	}
	log.WithField("dashboardUrl", "https://"+dc.client.DynatraceCreds.Tenant+"/#dashboards").Info("Dynatrace dashboard created successfully")
	return lib.ConfigResult{
		Success: true, // I guess this should be true not false?
		Message: "Dynatrace dashboard created successfully. You can view it here: https://" + dc.client.DynatraceCreds.Tenant + "/#dashboards",
	}
}

// deleteExistingDashboard deletes an existing dashboard for the provided project
func (dc *DashboardCreation) deleteExistingDashboard(project string) error {
	res, err := dc.client.SendDynatraceAPIRequest("/api/config/v1/dashboards", "GET", nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve list of existing Dynatrace dashboards: %v", err)
	}

	dtDashboardsResponse := &lib.DTDashboardsResponse{}
	err = json.Unmarshal([]byte(res), dtDashboardsResponse)
	if err != nil {
		err = checkForUnexpectedHTMLResponseError(err)
		return fmt.Errorf("failed to unmarshal list of existing Dynatrace dashboards: %v", err)
	}

	for _, dashboardItem := range dtDashboardsResponse.Dashboards {
		if dashboardItem.Name == project+dashboardNameSuffix {
			res, err = dc.client.SendDynatraceAPIRequest("/api/config/v1/dashboards/"+dashboardItem.ID, "DELETE", nil)
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

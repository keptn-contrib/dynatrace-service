package dynatrace

import (
	"encoding/json"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const dashboardsPath = "/api/config/v1/dashboards"

type DashboardsClient struct {
	client ClientInterface
}

func NewDashboardsClient(client ClientInterface) *DashboardsClient {
	return &DashboardsClient{
		client: client,
	}
}

func (dc *DashboardsClient) GetAll() (*Dashboards, error) {
	res, err := dc.client.Get(dashboardsPath)
	if err != nil {
		return nil, err
	}

	dashboards := &Dashboards{}
	err = json.Unmarshal(res, dashboards)
	if err != nil {
		err = CheckForUnexpectedHTMLResponseError(err)
		return nil, common.NewUnmarshalJSONError("Dynatrace dashboards", err)
	}

	return dashboards, nil
}

func (dc *DashboardsClient) GetByID(dashboardID string) (*Dashboard, error) {
	body, err := dc.client.Get(dashboardsPath + "/" + dashboardID)
	if err != nil {
		return nil, err
	}

	// parse json
	dynatraceDashboard := &Dashboard{}
	err = json.Unmarshal(body, &dynatraceDashboard)
	if err != nil {
		return nil, common.NewUnmarshalJSONError("Dynatrace dashboard", err)
	}

	return dynatraceDashboard, nil
}

func (dc *DashboardsClient) Create(dashboard *Dashboard) error {
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		return common.NewMarshalJSONError("Dynatrace dashboards", err)
	}

	_, err = dc.client.Post(dashboardsPath, dashboardPayload)
	if err != nil {
		return err
	}

	return nil
}

func (dc *DashboardsClient) Delete(dashboardID string) error {
	_, err := dc.client.Delete(dashboardsPath + "/" + dashboardID)
	if err != nil {
		return err
	}

	return nil
}

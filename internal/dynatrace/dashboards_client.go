package dynatrace

import (
	"encoding/json"
	"fmt"
)

const dashboardsPath = "/api/config/v1/dashboards"

type DashboardsClient struct {
	client *Client
}

func NewDashboardsClient(client *Client) *DashboardsClient {
	return &DashboardsClient{
		client: client,
	}
}

func (dc *DashboardsClient) GetAll() (*Dashboards, error) {
	res, err := dc.client.Get(dashboardsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve list of existing Dynatrace dashboards: %v", err)
	}

	dashboards := &Dashboards{}
	err = json.Unmarshal(res, dashboards)
	if err != nil {
		err = CheckForUnexpectedHTMLResponseError(err)
		return nil, fmt.Errorf("failed to unmarshal list of existing Dynatrace dashboards: %v", err)
	}

	return dashboards, nil
}

func (dc *DashboardsClient) Create(dashboard *Dashboard) error {
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Dynatrace dashboards: %v", err)
	}

	_, err = dc.client.Post(dashboardsPath, dashboardPayload)
	if err != nil {
		return fmt.Errorf("failed to create Dynatrace dashboards: %v", err)
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

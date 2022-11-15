package dynatrace

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

// DashboardsPath is the base endpoint for dashboards Config API
const DashboardsPath = "/api/config/v1/dashboards"

// DashboardsClient is a client for interacting with the dashboards configuration endpoint.
type DashboardsClient struct {
	client ClientInterface
}

// NewDashboardsClient creates a new DashboardsClient.
func NewDashboardsClient(client ClientInterface) *DashboardsClient {
	return &DashboardsClient{
		client: client,
	}
}

// GetAll gets a list of DashboardStubs detailling all accessible dashboards or returns an error.
func (dc *DashboardsClient) GetAll(ctx context.Context) (*DashboardList, error) {
	res, err := dc.client.Get(ctx, DashboardsPath)
	if err != nil {
		return nil, err
	}

	dashboards := &DashboardList{}
	err = json.Unmarshal(res, dashboards)
	if err != nil {
		err = CheckForUnexpectedHTMLResponseError(err)
		return nil, common.NewUnmarshalJSONError("Dynatrace dashboards", err)
	}

	return dashboards, nil
}

// GetByID gets a dashboard by ID or returns an error.
func (dc *DashboardsClient) GetByID(ctx context.Context, dashboardID string) (*Dashboard, error) {
	body, err := dc.client.Get(ctx, DashboardsPath+"/"+url.PathEscape(dashboardID))
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

// Create creates the specified dashboard or returns an error.
func (dc *DashboardsClient) Create(ctx context.Context, dashboard *Dashboard) error {
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		return common.NewMarshalJSONError("Dynatrace dashboard", err)
	}

	_, err = dc.client.Post(ctx, DashboardsPath, dashboardPayload)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the dashboard referenced by the specified ID or returns an error.
func (dc *DashboardsClient) Delete(ctx context.Context, dashboardID string) error {
	_, err := dc.client.Delete(ctx, DashboardsPath+"/"+url.PathEscape(dashboardID))
	if err != nil {
		return err
	}

	return nil
}

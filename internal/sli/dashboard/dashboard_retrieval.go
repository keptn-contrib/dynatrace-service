package dashboard

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

const queryDashboardProperty = "query"

type Retrieval struct {
	client    dynatrace.ClientInterface
	eventData adapter.EventContentAdapter
}

func NewRetrieval(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter) *Retrieval {
	return &Retrieval{
		client:    client,
		eventData: eventData,
	}
}

// Retrieve Depending on the dashboard parameter which is pulled from dynatrace.conf.yaml:dashboard this method either
//   - query:        queries all dashboards on the Dynatrace Tenant and returns the one that matches project/service/stage, or
//   - dashboard-ID: if this is a valid dashboard ID it will query the dashboard with this ID, e.g: ddb6a571-4bda-4e8b-a9c0-4a3e02c2e14a, or
//
// It returns a parsed Dynatrace Dashboard and the actual dashboard ID in case we queried a dashboard.
func (r *Retrieval) Retrieve(ctx context.Context, dashboardProperty string) (*dynatrace.Dashboard, error) {
	dashboard, err := r.retrieve(ctx, dashboardProperty)
	if err != nil {
		return nil, NewRetrievalError(err)
	}

	return dashboard, nil
}

func (r *Retrieval) retrieve(ctx context.Context, dashboardProperty string) (*dynatrace.Dashboard, error) {
	dashboardID, err := r.convertDashboardPropertyToID(ctx, dashboardProperty)
	if err != nil {
		return nil, err
	}

	return dynatrace.NewDashboardsClient(r.client).GetByID(ctx, dashboardID)
}

func (r *Retrieval) convertDashboardPropertyToID(ctx context.Context, dashboardProperty string) (string, error) {
	if dashboardProperty == "" {
		return "", fmt.Errorf("invalid 'dashboard' property - either specify a dashboard ID or use 'query'")
	}

	if dashboardProperty == queryDashboardProperty {
		return r.findDynatraceDashboard(ctx)
	}

	// assume what is left is a dashboard UUID
	return dashboardProperty, nil
}

func (r *Retrieval) findDynatraceDashboard(ctx context.Context) (string, error) {
	dashboardList, err := dynatrace.NewDashboardsClient(r.client).GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("could not get dashboard list: %w", err)
	}

	return dashboardList.SearchForDashboardMatching(r.eventData.GetProject(), r.eventData.GetStage(), r.eventData.GetService())
}

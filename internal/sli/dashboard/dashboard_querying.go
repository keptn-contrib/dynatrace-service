package dashboard

import (
	"context"
	"fmt"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

// Querying interacts with a dynatrace API endpoint
type Querying struct {
	eventData        adapter.EventContentAdapter
	customSLIFilters []*keptnv2.SLIFilter
	dtClient         dynatrace.ClientInterface
}

// NewQuerying returns a new dynatrace handler that interacts with the Dynatrace REST API
func NewQuerying(eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, dtClient dynatrace.ClientInterface) *Querying {
	return &Querying{
		eventData:        eventData,
		customSLIFilters: customFilters,
		dtClient:         dtClient,
	}
}

// GetSLIValues implements - https://github.com/keptn-contrib/dynatrace-sli-service/issues/60
// Queries Dynatrace for the existence of a dashboard tagged with keptn_project:project, keptn_stage:stage, keptn_service:service, SLI
// if this dashboard exists it will be parsed and a custom SLI_dashboard.yaml and an SLO_dashboard.yaml will be created
// Returns a QueryResult or an error
func (q *Querying) GetSLIValues(ctx context.Context, dashboardID string, timeframe common.Timeframe) (*QueryResult, error) {
	// let's load the dashboard if needed
	dashboard, dashboardID, err := NewRetrieval(q.dtClient, q.eventData).Retrieve(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("error while processing dashboard config '%s' - %w", dashboardID, err)
	}

	return NewProcessing(q.dtClient, q.eventData, q.customSLIFilters, timeframe).Process(ctx, dashboard), nil
}

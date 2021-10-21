package dashboard

import (
	"errors"
	"fmt"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"

	log "github.com/sirupsen/logrus"
)

// Querying interacts with a dynatrace API endpoint
type Querying struct {
	eventData        adapter.EventContentAdapter
	customSLIFilters []*keptnv2.SLIFilter
	dtClient         dynatrace.ClientInterface
	dashboardReader  keptn.DashboardResourceReaderInterface
}

// NewQuerying returns a new dynatrace handler that interacts with the Dynatrace REST API
func NewQuerying(eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, dtClient dynatrace.ClientInterface, dashboardReader keptn.DashboardResourceReaderInterface) *Querying {
	return &Querying{
		eventData:        eventData,
		customSLIFilters: customFilters,
		dtClient:         dtClient,
		dashboardReader:  dashboardReader,
	}
}

// GetSLIValues implements - https://github.com/keptn-contrib/dynatrace-sli-service/issues/60
// Queries Dynatrace for the existence of a dashboard tagged with keptn_project:project, keptn_stage:stage, keptn_service:service, SLI
// if this dashboard exists it will be parsed and a custom SLI_dashboard.yaml and an SLO_dashboard.yaml will be created
// Returns a QueryResult, a bool indicating whether the dashboard was processed, or an error
func (q *Querying) GetSLIValues(dashboardID string, startUnix time.Time, endUnix time.Time) (*QueryResult, bool, error) {
	existingDashboardContent, err := q.dashboardReader.GetDashboard(q.eventData.GetProject(), q.eventData.GetStage(), q.eventData.GetService())
	if err != nil {
		// only fail if there is a problem with dashboard. Having no dashboard stored is not a problem, so continue
		var rnfErr *keptn.ResourceNotFoundError
		if !errors.As(err, &rnfErr) {
			return nil, false, err
		}
	}

	// let's load the dashboard if needed
	dashbd, dashboardID, err := NewRetrieval(q.dtClient, q.eventData).Retrieve(dashboardID)
	if err != nil {
		return nil, false, fmt.Errorf("error while processing dashboard config '%s' - %w", dashboardID, err)
	}

	// Lets validate if we really need to process this dashboard as it might be the same (without change) from the previous runs
	// see https://github.com/keptn-contrib/dynatrace-sli-service/issues/92 for more details
	if dashbd.IsTheSameAs(existingDashboardContent) {
		log.Debug("Dashboard hasn't changed: skipping parsing of dashboard")
		return NewQueryResultFrom(
				NewLink(
					q.dtClient.Credentials().GetTenant(),
					startUnix,
					endUnix,
					dashbd.ID,
					dashbd.GetFilter())),
			false,
			nil
	}

	return NewProcessing(q.dtClient, q.eventData, q.customSLIFilters, startUnix, endUnix).Process(dashbd), true, nil
}

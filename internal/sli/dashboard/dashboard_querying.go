package dashboard

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
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
// Queries Dynatrace for the existance of a dashboard tagged with keptn_project:project, keptn_stage:stage, keptn_service:service, SLI
// if this dashboard exists it will be parsed and a custom SLI_dashboard.yaml and an SLO_dashboard.yaml will be created
// Returns:
//  #1: Link to Dashboard
//  #2: SLI
//  #3: ServiceLevelObjectives
//  #4: SLIResult
//  #5: Error
func (q *Querying) GetSLIValues(dashboardID string, startUnix time.Time, endUnix time.Time) (*QueryResult, error) {

	// Lets see if there is a dashboard.json already in the configuration repo - if so its an indicator that we should query the dashboard
	// This check is especially important for backward compatibility as the new dynatrace.conf.yaml:dashboard property is changing the default behavior
	// If a dashboard.json exists and dashboard property is empty we default to QUERY - which is the old default behavior
	existingDashboardContent, err := q.dashboardReader.GetDashboard(q.eventData.GetProject(), q.eventData.GetStage(), q.eventData.GetService())
	if err == nil && existingDashboardContent != "" && dashboardID == "" {
		log.Debug("Set dashboard=query for backward compatibility as dashboard.json was present!")
		dashboardID = common.DynatraceConfigDashboardQUERY
	}

	// lets load the dashboard if needed
	dashbd, dashboardID, err := NewRetrieval(q.dtClient, q.eventData).Retrieve(dashboardID)
	if err != nil {
		return nil, fmt.Errorf("error while processing dashboard config '%s' - %w", dashboardID, err)
	}

	if dashbd == nil {
		return nil, nil
	}

	// Lets validate if we really need to process this dashboard as it might be the same (without change) from the previous runs
	// see https://github.com/keptn-contrib/dynatrace-sli-service/issues/92 for more details
	if dashbd.IsTheSameAs(existingDashboardContent) {
		log.Debug("Dashboard hasn't changed: skipping parsing of dashboard")
		return NewQueryResultFrom(
				NewLink(
					q.dtClient.Credentials().Tenant,
					startUnix,
					endUnix,
					dashbd.ID,
					dashbd.GetFilter())),
			nil
	}

	return NewProcessing(q.dtClient, q.eventData, q.customSLIFilters, startUnix, endUnix).Process(dashbd), nil
}

package dashboard

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type TileResult struct {
	sliResult *keptnv2.SLIResult
	objective *keptnapi.SLO
	sliName   string
	sliQuery  string
}

// QueryResult is the object returned by querying a Dynatrace dashboard for SLIs
type QueryResult struct {
	dashboardLink *DashboardLink
	dashboard     *dynatrace.Dashboard
	sli           *dynatrace.SLI
	slo           *keptnapi.ServiceLevelObjectives
	sliResults    []*keptnv2.SLIResult
}

// NewQueryResultFrom creates a new QueryResult object just from a DashboardLink
func NewQueryResultFrom(dashboardLink *DashboardLink) *QueryResult {
	return &QueryResult{
		dashboardLink: dashboardLink,
	}
}

func (r *QueryResult) DashboardLink() *DashboardLink {
	return r.dashboardLink
}

func (r *QueryResult) Dashboard() *dynatrace.Dashboard {
	return r.dashboard
}

func (r *QueryResult) SLI() *dynatrace.SLI {
	return r.sli
}

// HasSLIs checks whether any indicators are available
func (r *QueryResult) HasSLIs() bool {
	return r.sli != nil && len(r.sli.Indicators) > 0
}

func (r *QueryResult) SLO() *keptnapi.ServiceLevelObjectives {
	return r.slo
}

// HasSLOs checks whether any objectives are available
func (r *QueryResult) HasSLOs() bool {
	return r.slo != nil && len(r.slo.Objectives) > 0
}

func (r *QueryResult) SLIResults() []*keptnv2.SLIResult {
	return r.sliResults
}

// addTileResult adds a TileResult to the QueryResult, also allows nil values for convenience
func (r *QueryResult) addTileResult(result *TileResult) {
	if result == nil {
		return
	}

	r.sli.Indicators[result.sliName] = result.sliQuery
	r.slo.Objectives = append(r.slo.Objectives, result.objective)
	r.sliResults = append(r.sliResults, result.sliResult)
}

// addTileResult adds multiple TileResult to the QueryResult,
func (r *QueryResult) addTileResults(results []*TileResult) {
	for _, result := range results {
		r.addTileResult(result)
	}
}

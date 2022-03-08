package dashboard

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
)

// QueryResult is the object returned by querying a Dynatrace dashboard for SLIs
type QueryResult struct {
	dashboardLink *DashboardLink
	dashboard     *dynatrace.Dashboard
	sli           *dynatrace.SLI
	slo           *keptnapi.ServiceLevelObjectives
	sliResults    []result.SLIResult
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

// SLIs gets the SLIs.
func (r *QueryResult) SLIs() *dynatrace.SLI {
	return r.sli
}

// HasSLIs checks whether any indicators are available
func (r *QueryResult) HasSLIs() bool {
	return r.sli != nil && len(r.sli.Indicators) > 0
}

// SLOs gets the SLOs.
func (r *QueryResult) SLOs() *keptnapi.ServiceLevelObjectives {
	return r.slo
}

// HasSLOs checks whether any objectives are available
func (r *QueryResult) HasSLOs() bool {
	return r.slo != nil && len(r.slo.Objectives) > 0
}

// SLIResults gets the SLI results.
func (r *QueryResult) SLIResults() []result.SLIResult {
	return r.sliResults
}

// addTileResult adds a TileResult to the QueryResult, also allows nil values for convenience
func (r *QueryResult) addTileResult(result *TileResult) {
	if result == nil {
		return
	}

	r.sli.Indicators[result.sliName] = result.sliQuery

	if result.objective != nil {
		r.slo.Objectives = append(r.slo.Objectives, result.objective)
	}

	r.sliResults = append(r.sliResults, result.sliResult)
}

// addTileResult adds multiple TileResult to the QueryResult,
func (r *QueryResult) addTileResults(results []*TileResult) {
	for _, result := range results {
		r.addTileResult(result)
	}
}

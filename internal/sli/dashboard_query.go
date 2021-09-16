package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type tileResult struct {
	sliResult *keptnv2.SLIResult
	objective *keptnapi.SLO
	sliName   string
	sliQuery  string
}

// dashboardQueryResult is the object returned by querying a Dynatrace dashboard for SLIs
type dashboardQueryResult struct {
	dashboardLink *DashboardLink
	dashboard     *dynatrace.Dashboard
	sli           *dynatrace.SLI
	slo           *keptnapi.ServiceLevelObjectives
	sliResults    []*keptnv2.SLIResult
}

// newDashboardQueryResultFrom creates a new dashboardQueryResult object just from a DashboardLink
func newDashboardQueryResultFrom(dashboardLink *DashboardLink) *dashboardQueryResult {
	return &dashboardQueryResult{
		dashboardLink: dashboardLink,
	}
}

func (r *dashboardQueryResult) DashboardLink() *DashboardLink {
	return r.dashboardLink
}

func (r *dashboardQueryResult) Dashboard() *dynatrace.Dashboard {
	return r.dashboard
}

func (r *dashboardQueryResult) SLI() *dynatrace.SLI {
	return r.sli
}

func (r *dashboardQueryResult) SLO() *keptnapi.ServiceLevelObjectives {
	return r.slo
}

func (r *dashboardQueryResult) SLIResults() []*keptnv2.SLIResult {
	return r.sliResults
}

// addTileResult adds a tileResult to the dashboardQueryResult, also allows nil values for convenience
func (r *dashboardQueryResult) addTileResult(result *tileResult) {
	if result == nil {
		return
	}

	r.sli.Indicators[result.sliName] = result.sliQuery
	r.slo.Objectives = append(r.slo.Objectives, result.objective)
	r.sliResults = append(r.sliResults, result.sliResult)
}

// addTileResult adds multiple tileResult to the dashboardQueryResult,
func (r *dashboardQueryResult) addTileResults(results []*tileResult) {
	for _, result := range results {
		r.addTileResult(result)
	}
}

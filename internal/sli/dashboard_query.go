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

// DashboardQueryResult is the object returned by querying a Dynatrace dashboard for SLIs
type DashboardQueryResult struct {
	dashboardLink *DashboardLink
	dashboard     *dynatrace.Dashboard
	sli           *dynatrace.SLI
	slo           *keptnapi.ServiceLevelObjectives
	sliResults    []*keptnv2.SLIResult
}

// NewDashboardQueryResultFrom creates a new DashboardQueryResult object just from a DashboardLink
func NewDashboardQueryResultFrom(dashboardLink *DashboardLink) *DashboardQueryResult {
	return &DashboardQueryResult{
		dashboardLink: dashboardLink,
	}
}

func (r *DashboardQueryResult) DashboardLink() *DashboardLink {
	return r.dashboardLink
}

func (r *DashboardQueryResult) Dashboard() *dynatrace.Dashboard {
	return r.dashboard
}

func (r *DashboardQueryResult) SLI() *dynatrace.SLI {
	return r.sli
}

func (r *DashboardQueryResult) SLO() *keptnapi.ServiceLevelObjectives {
	return r.slo
}

func (r *DashboardQueryResult) SLIResults() []*keptnv2.SLIResult {
	return r.sliResults
}

func (r *DashboardQueryResult) addTileResult(result tileResult) {
	r.sli.Indicators[result.sliName] = result.sliQuery
	r.slo.Objectives = append(r.slo.Objectives, result.objective)
	r.sliResults = append(r.sliResults, result.sliResult)
}

func (r *DashboardQueryResult) addTileResults(results []tileResult) {
	for _, result := range results {
		r.addTileResult(result)
	}
}

package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// DashboardQueryResult is the object returned by querying a Dynatrace dashboard for SLIs
type DashboardQueryResult struct {
	dashboardLink *DashboardLink
	dashboard     *dynatrace.Dashboard
	sli           *dynatrace.SLI
	slo           *keptn.ServiceLevelObjectives
	sliResults    []*keptnv2.SLIResult
}

// NewDashboardQueryResultFrom creates a new DashboardQueryResult object just from a DashboardLink
func NewDashboardQueryResultFrom(dashboardLink *DashboardLink) *DashboardQueryResult {
	return &DashboardQueryResult{
		dashboardLink: dashboardLink,
	}
}

func (result *DashboardQueryResult) DashboardLink() *DashboardLink {
	return result.dashboardLink
}

func (result *DashboardQueryResult) Dashboard() *dynatrace.Dashboard {
	return result.dashboard
}

func (result *DashboardQueryResult) SLI() *dynatrace.SLI {
	return result.sli
}

func (result *DashboardQueryResult) SLO() *keptn.ServiceLevelObjectives {
	return result.slo
}

func (result *DashboardQueryResult) SLIResults() []*keptnv2.SLIResult {
	return result.sliResults
}

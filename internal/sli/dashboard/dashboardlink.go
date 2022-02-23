package dashboard

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type DashboardLink struct {
	apiURL          string
	timeframe       common.Timeframe
	dashboardID     string
	dashboardFilter *dynatrace.DashboardFilter
}

func NewLink(
	apiURL string,
	timeframe common.Timeframe,
	dashboardID string,
	dashboardFilter *dynatrace.DashboardFilter) *DashboardLink {
	return &DashboardLink{
		apiURL:          apiURL,
		timeframe:       timeframe,
		dashboardID:     dashboardID,
		dashboardFilter: dashboardFilter,
	}
}

func (dashboardLink *DashboardLink) String() string {
	managementZone := ""
	if dashboardLink.dashboardFilter != nil && dashboardLink.dashboardFilter.ManagementZone != nil {
		managementZone = ";gf=" + dashboardLink.dashboardFilter.ManagementZone.ID
	}

	return fmt.Sprintf("%s#dashboard;id=%s;gtf=c_%s_%s%s",
		dashboardLink.apiURL,
		dashboardLink.dashboardID,
		common.TimestampToUnixMillisecondsString(dashboardLink.timeframe.Start()),
		common.TimestampToUnixMillisecondsString(dashboardLink.timeframe.End()),
		managementZone)
}

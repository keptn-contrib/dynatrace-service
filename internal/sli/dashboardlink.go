package sli

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"time"
)

type DashboardLink struct {
	apiURL          string
	startTimestamp  time.Time
	endTimestamp    time.Time
	dashboardID     string
	dashboardFilter *dynatrace.DashboardFilter
}

func NewDashboardLink(
	apiURL string,
	startTimestamp time.Time,
	endTimestamp time.Time,
	dashboardID string,
	dashboardFilter *dynatrace.DashboardFilter) *DashboardLink {
	return &DashboardLink{
		apiURL:          apiURL,
		startTimestamp:  startTimestamp,
		endTimestamp:    endTimestamp,
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
		common.TimestampToString(dashboardLink.startTimestamp),
		common.TimestampToString(dashboardLink.endTimestamp),
		managementZone)
}

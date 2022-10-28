package dashboard

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type ManagementZoneFilter struct {
	dashboardFilter    *dynatrace.DashboardFilter
	tileManagementZone *dynatrace.ManagementZoneEntry
}

func NewManagementZoneFilter(
	dashboardManagementZone *dynatrace.DashboardFilter,
	tileManagementZone *dynatrace.ManagementZoneEntry,
) *ManagementZoneFilter {
	return &ManagementZoneFilter{
		dashboardFilter:    dashboardManagementZone,
		tileManagementZone: tileManagementZone,
	}
}

// ForProblemSelector returns the ID of the ManagementZone in a valid representation for the problemSelector.
// If a ManagementZone for a Dashboard tile is given, then it will take precedence over the ManagementZone of the DashboardFilter
// If none of both are given it will return an empty string
func (filter *ManagementZoneFilter) ForProblemSelector() string {
	return filter.forSelector(createFilterQueryForProblemSelector)
}

// ForMZSelector returns the ID of the ManagementZone in a valid representation for the mzSelector.
// If a ManagementZone for a Dashboard tile is given, then it will take precedence over the ManagementZone of the DashboardFilter
// If none of both are given it will return an empty string
func (filter *ManagementZoneFilter) ForMZSelector() string {
	return filter.forSelector(createFilterQueryForMZSelector)
}

func (filter *ManagementZoneFilter) forSelector(mapper func(string) string) string {
	if filter.tileManagementZone != nil {
		return mapper(filter.tileManagementZone.ID)
	}

	if filter.dashboardFilter != nil && filter.dashboardFilter.ManagementZone != nil {
		return mapper(filter.dashboardFilter.ManagementZone.ID)
	}

	return ""
}

func createFilterQueryForProblemSelector(managementZoneID string) string {
	return fmt.Sprintf(",managementZoneIds(%s)", managementZoneID)
}

func createFilterQueryForMZSelector(managementZoneID string) string {
	return fmt.Sprintf("mzId(%s)", managementZoneID)
}

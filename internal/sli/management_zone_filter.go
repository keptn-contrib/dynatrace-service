package sli

import "fmt"

type ManagementZoneFilter struct {
	dashboardFilter    *DashboardFilter
	tileManagementZone *ManagementZone
}

func NewManagementZoneFilter(
	dashboardManagementZone *DashboardFilter,
	tileManagementZone *ManagementZone,
) *ManagementZoneFilter {
	return &ManagementZoneFilter{
		dashboardFilter:    dashboardManagementZone,
		tileManagementZone: tileManagementZone,
	}
}

// ForEntitySelector returns the ID of the ManagementZone in a valid representation for the entitySelector.
// If a ManagementZone for a Dashboard tile is given, then it will take precedence over the ManagementZone of the DashboardFilter
// If none of both are given it will return an empty string
func (filter *ManagementZoneFilter) ForEntitySelector() string {
	return filter.forSelector(createFilterQueryForEntitySelector)
}

// ForProblemSelector returns the ID of the ManagementZone in a valid representation for the problemSelector.
// If a ManagementZone for a Dashboard tile is given, then it will take precedence over the ManagementZone of the DashboardFilter
// If none of both are given it will return an empty string
func (filter *ManagementZoneFilter) ForProblemSelector() string {
	return filter.forSelector(createFilterQueryForProblemSelector)
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

func createFilterQueryForEntitySelector(managementZoneID string) string {
	return fmt.Sprintf(",mzId(%s)", managementZoneID)
}

func createFilterQueryForProblemSelector(managementZoneID string) string {
	return fmt.Sprintf(",managementZoneIds(%s)", managementZoneID)
}

package dynatrace

import (
	"fmt"
	"strings"
	"testing"
)

type dashboardTestConfig struct {
	testDescription     string
	dashboards          Dashboards
	expectedDashboardID string
}

func TestDynatraceDashboards_SearchForDashboardMatching(t *testing.T) {
	const project = "sockshop"
	const service = "carts"
	const stage = "staging"

	const desiredDashboardID = "311f4aa7-5257-41d7-abd1-70420500e1c8"

	exactNameMatchForEvent := createDashboardNameFor(project, service, stage)
	matchingDashboard := createDashboard(desiredDashboardID, exactNameMatchForEvent)

	configs := []dashboardTestConfig{
		{
			testDescription:     "full match, single dashboard",
			dashboards:          createDashboards(matchingDashboard),
			expectedDashboardID: desiredDashboardID,
		},
		{
			testDescription: "full match, multiple dashboards for same project and service",
			dashboards: createDashboards(
				createDashboardWith("dashboard-1", project, service, "dev"),
				matchingDashboard,
				createDashboardWith("dashboard-3", project, service, "production")),
			expectedDashboardID: desiredDashboardID,
		},
		{
			testDescription: "full match, multiple dashboards for same project and stage",
			dashboards: createDashboards(
				createDashboardWith("dashboard-1", project, "carts-v1", stage),
				createDashboardWith("dashboard-2", project, "carts-v2", stage),
				matchingDashboard),
			expectedDashboardID: desiredDashboardID,
		},
		{
			testDescription: "full match, multiple dashboards for same service and stage",
			dashboards: createDashboards(
				matchingDashboard,
				createDashboardWith("dashboard-2", "sockshop-v2", service, stage),
				createDashboardWith("dashboard-3", "sockshop-v3", service, stage)),
			expectedDashboardID: desiredDashboardID,
		},
		{
			testDescription: "no match, but multiple dashboards for same subsets of project, service and stage",
			dashboards: createDashboards(
				createDashboardWith("dashboard-1", project, service, "production"),
				createDashboardWith("dashboard-2", "sockshop-v2", service, stage),
				createDashboardWith("dashboard-3", project, "carts-v2", stage)),
			expectedDashboardID: "",
		},
		{
			testDescription: "no match, because only a subset of project, service and/or stage are given and would match",
			dashboards: createDashboards(
				createDashboardWith("dashboard-1", project, service, ""),
				createDashboardWith("dashboard-2", "", service, stage),
				createDashboardWith("dashboard-3", project, "", stage),
				createDashboardWith("dashboard-4", project, "", ""),
				createDashboardWith("dashboard-5", "", service, ""),
				createDashboardWith("dashboard-6", "", "", stage),
				createDashboardWith("dashboard-7", "", "", "")),
			expectedDashboardID: "",
		},
		{
			testDescription: "no match, and multiple dashboards without matching subsets of project, service and stage",
			dashboards: createDashboards(
				createDashboardWith("dashboard-1", "sockshop-v2", "carts-v2", "production"),
				createDashboardWith("dashboard-2", "sockshop-v2", "carts-v1", "dev"),
				createDashboardWith("dashboard-3", "sockshop-v2", "carts-v3", "hardening")),
			expectedDashboardID: "",
		},
		{
			testDescription: "no match, single dashboards with nearly matching name",
			dashboards: createDashboards(
				createDashboard("dashboard-1", strings.TrimPrefix(exactNameMatchForEvent, "KQG;"))),
			expectedDashboardID: "",
		},
		{
			testDescription: "no match, multiple dashboards with standard names",
			dashboards: createDashboards(
				createDashboard("dashboard-1", "Dashboard 1"),
				createDashboard("dashboard-2", "Dashboard 2")),
			expectedDashboardID: "",
		},
		{
			testDescription:     "no match, because there are no dashboards",
			dashboards:          createDashboards(),
			expectedDashboardID: "",
		},
	}

	for _, config := range configs {
		actualDashboardID := config.dashboards.SearchForDashboardMatching(project, stage, service)
		if actualDashboardID != config.expectedDashboardID {
			t.Errorf(
				"Test: %s - expected: %s, but got: %s",
				config.testDescription,
				config.expectedDashboardID,
				actualDashboardID)
		}
	}
}

func createDashboardNameFor(project string, service string, stage string) string {
	dashboardName := "KQG;"
	if project != "" {
		dashboardName = fmt.Sprintf("%sproject=%s;", dashboardName, project)
	}
	if service != "" {
		dashboardName = fmt.Sprintf("%sservice=%s;", dashboardName, service)
	}
	if stage != "" {
		dashboardName = fmt.Sprintf("%sstage=%s;", dashboardName, stage)
	}

	return dashboardName + "something-else"
}

func createDashboardWith(dashboardID string, project string, service string, stage string) DashboardEntry {
	return createDashboard(
		dashboardID,
		createDashboardNameFor(project, service, stage))
}

func createDashboard(dashboardID string, dashboardName string) DashboardEntry {
	return DashboardEntry{
		ID:   dashboardID,
		Name: dashboardName,
	}
}

func createDashboards(dashboards ...DashboardEntry) Dashboards {
	return Dashboards{
		Dashboards: dashboards,
	}
}

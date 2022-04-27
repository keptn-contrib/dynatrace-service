package dynatrace

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboardList_SearchForDashboardMatching(t *testing.T) {

	const project = "sockshop"
	const service = "carts"
	const stage = "staging"

	const desiredDashboardID = "311f4aa7-5257-41d7-abd1-70420500e1c8"

	exactNameMatchForEvent := createDashboardNameFor(project, service, stage)
	matchingDashboard := createDashboardStub(desiredDashboardID, exactNameMatchForEvent)

	tests := []struct {
		name                string
		dashboardList       DashboardList
		expectedDashboardID string
		expectError         bool
		partialErrorMessage string
	}{
		{
			name:                "full match, single dashboard",
			dashboardList:       createDashboardList(matchingDashboard),
			expectedDashboardID: desiredDashboardID,
		},
		{
			name: "full match, multiple dashboards for same project and service",
			dashboardList: createDashboardList(
				createDashboardStubWith("dashboard-1", project, service, "dev"),
				matchingDashboard,
				createDashboardStubWith("dashboard-3", project, service, "production")),
			expectedDashboardID: desiredDashboardID,
		},
		{
			name: "full match, multiple dashboards for same project and stage",
			dashboardList: createDashboardList(
				createDashboardStubWith("dashboard-1", project, "carts-v1", stage),
				createDashboardStubWith("dashboard-2", project, "carts-v2", stage),
				matchingDashboard),
			expectedDashboardID: desiredDashboardID,
		},
		{
			name: "full match, multiple dashboards for same service and stage",
			dashboardList: createDashboardList(
				matchingDashboard,
				createDashboardStubWith("dashboard-2", "sockshop-v2", service, stage),
				createDashboardStubWith("dashboard-3", "sockshop-v3", service, stage)),
			expectedDashboardID: desiredDashboardID,
		},
		{
			name: "no match, but multiple dashboards for same subsets of project, service and stage",
			dashboardList: createDashboardList(
				createDashboardStubWith("dashboard-1", project, service, "production"),
				createDashboardStubWith("dashboard-2", "sockshop-v2", service, stage),
				createDashboardStubWith("dashboard-3", project, "carts-v2", stage)),
			expectError:         true,
			partialErrorMessage: "no dashboard name matches the name specification",
		},
		{
			name: "no match, because only a subset of project, service and/or stage are given and would match",
			dashboardList: createDashboardList(
				createDashboardStubWith("dashboard-1", project, service, ""),
				createDashboardStubWith("dashboard-2", "", service, stage),
				createDashboardStubWith("dashboard-3", project, "", stage),
				createDashboardStubWith("dashboard-4", project, "", ""),
				createDashboardStubWith("dashboard-5", "", service, ""),
				createDashboardStubWith("dashboard-6", "", "", stage),
				createDashboardStubWith("dashboard-7", "", "", "")),
			expectError:         true,
			partialErrorMessage: "no dashboard name matches the name specification",
		},
		{
			name: "no match, and multiple dashboards without matching subsets of project, service and stage",
			dashboardList: createDashboardList(
				createDashboardStubWith("dashboard-1", "sockshop-v2", "carts-v2", "production"),
				createDashboardStubWith("dashboard-2", "sockshop-v2", "carts-v1", "dev"),
				createDashboardStubWith("dashboard-3", "sockshop-v2", "carts-v3", "hardening")),
			expectError:         true,
			partialErrorMessage: "no dashboard name matches the name specification",
		},
		{
			name: "no match, single dashboards with nearly matching name",
			dashboardList: createDashboardList(
				createDashboardStub("dashboard-1", strings.TrimPrefix(exactNameMatchForEvent, "KQG;"))),
			expectError:         true,
			partialErrorMessage: "no dashboard name matches the name specification",
		},
		{
			name: "no match, multiple dashboards with standard names",
			dashboardList: createDashboardList(
				createDashboardStub("dashboard-1", "Dashboard 1"),
				createDashboardStub("dashboard-2", "Dashboard 2")),
			expectError:         true,
			partialErrorMessage: "no dashboard name matches the name specification",
		},
		{
			name:                "no match, because there are no dashboards",
			dashboardList:       createDashboardList(),
			expectError:         true,
			partialErrorMessage: "no dashboard name matches the name specification",
		},
		{
			name: "multiple dashboards match",
			dashboardList: createDashboardList(
				matchingDashboard,
				createDashboardStubWith("dashboard-1", project, "carts-v1", stage),
				createDashboardStubWith("dashboard-2", project, "carts-v2", stage),
				matchingDashboard),
			expectError:         true,
			partialErrorMessage: "2 dashboards match the name specification",
		},
		{
			name: "multiple dashboards match - different key order",
			dashboardList: createDashboardList(
				matchingDashboard,
				createDashboardStub(desiredDashboardID, fmt.Sprintf("kqg;project=%s;stage=%s;service=%s", project, stage, service))),
			expectError:         true,
			partialErrorMessage: "2 dashboards match the name specification",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dashboardID, err := tt.dashboardList.SearchForDashboardMatching(project, stage, service)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.partialErrorMessage)
				assert.Empty(t, dashboardID)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.expectedDashboardID, dashboardID)
			}
		})
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

func createDashboardStubWith(dashboardID string, project string, service string, stage string) DashboardStub {
	return createDashboardStub(
		dashboardID,
		createDashboardNameFor(project, service, stage))
}

func createDashboardStub(dashboardID string, dashboardName string) DashboardStub {
	return DashboardStub{
		ID:   dashboardID,
		Name: dashboardName,
	}
}

func createDashboardList(dashboardStubs ...DashboardStub) DashboardList {
	return DashboardList{
		Dashboards: dashboardStubs,
	}
}

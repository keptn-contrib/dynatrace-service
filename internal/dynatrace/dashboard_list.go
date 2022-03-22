package dynatrace

import (
	"fmt"
	"strings"
)

// DashboardList is a list of short representations of dashboards returned by the /dashboards endpoint
type DashboardList struct {
	Dashboards []DashboardStub `json:"dashboards"`
}

// DashboardStub is a short representation of a dashboard
type DashboardStub struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// SearchForDashboardMatching searches for a dashboard that has the prefix "KQG;" and criteria "project=PROJECT", "service=SERVICE" and "stage=STAGE"
// It returns the ID of the dashboard if exactly one dashboard matches or an error otherwise
func (dashboards *DashboardList) SearchForDashboardMatching(project string, stage string, service string) (string, error) {
	namePrefix := "kqg;"
	projectKeyValuePair := strings.ToLower("project=" + project)
	stageKeyValuePair := strings.ToLower("stage=" + stage)
	serviceKeyValuePair := strings.ToLower("service=" + service)

	var matchingDashboardIds []string
	for _, dashboardStub := range dashboards.Dashboards {
		dashboardNameLowerCase := strings.ToLower(dashboardStub.Name)

		if !strings.HasPrefix(dashboardNameLowerCase, namePrefix) {
			continue
		}

		nameSplits := strings.Split(dashboardNameLowerCase, ";")

		if !sliceContainsString(nameSplits, projectKeyValuePair) {
			continue
		}
		if !sliceContainsString(nameSplits, stageKeyValuePair) {
			continue
		}
		if !sliceContainsString(nameSplits, serviceKeyValuePair) {
			continue
		}

		matchingDashboardIds = append(matchingDashboardIds, dashboardStub.ID)
	}

	switch len(matchingDashboardIds) {
	case 0:
		return "", fmt.Errorf("No dashboard name matches the name specification with prefix '%s' and criteria '%s', '%s', '%s'", namePrefix, projectKeyValuePair, stageKeyValuePair, serviceKeyValuePair)
	case 1:
		return matchingDashboardIds[0], nil
	default:
		return "", fmt.Errorf("%d dashboards match the name specification with prefix '%s' and criteria '%s', '%s', '%s'", len(matchingDashboardIds), namePrefix, projectKeyValuePair, stageKeyValuePair, serviceKeyValuePair)
	}

}

func sliceContainsString(slice []string, wantedValue string) bool {
	for _, value := range slice {
		if value == wantedValue {
			return true
		}
	}
	return false
}

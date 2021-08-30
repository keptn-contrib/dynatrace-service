package sli

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"strings"
)

type NestedFilterDataExplorer struct {
	Filter         string                     `json:"filter"`
	FilterType     string                     `json:"filterType"`
	FilterOperator string                     `json:"filterOperator"`
	NestedFilters  []NestedFilterDataExplorer `json:"nestedFilters"`
	Criteria       []struct {
		Value     string `json:"value"`
		Evaluator string `json:"evaluator"`
	} `json:"criteria"`
}

// DataExplorerQuery Query Definition for DATA_EXPLORER dashboard tile
type DataExplorerQuery struct {
	ID               string   `json:"id"`
	Metric           string   `json:"metric"`
	SpaceAggregation string   `json:"spaceAggregation"`
	TimeAggregation  string   `json:"timeAggregation"`
	SplitBy          []string `json:"splitBy"`
	FilterBy         *struct {
		FilterOperator string                     `json:"filterOperator"`
		NestedFilters  []NestedFilterDataExplorer `json:"nestedFilters"`
		Criteria       []struct {
			Value     string `json:"value"`
			Evaluator string `json:"evaluator"`
		} `json:"criteria"`
	} `json:"filterBy,omitempty"`
}

// ChartSeries Chart Series for a regular Chart
type ChartSeries struct {
	Metric      string      `json:"metric"`
	Aggregation string      `json:"aggregation"`
	Percentile  interface{} `json:"percentile"`
	Type        string      `json:"type"`
	EntityType  string      `json:"entityType"`
	Dimensions  []struct {
		ID              string   `json:"id"`
		Name            string   `json:"name"`
		Values          []string `json:"values"`
		EntityDimension bool     `json:"entitiyDimension"`
	} `json:"dimensions"`
	SortAscending   bool   `json:"sortAscending"`
	SortColumn      bool   `json:"sortColumn"`
	AggregationRate string `json:"aggregationRate"`
}

type ManagementZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DashboardFilter struct {
	Timeframe      string          `json:"timeframe"`
	ManagementZone *ManagementZone `json:"managementZone,omitempty"`
}

type DashboardMetadata struct {
	Name           string `json:"name"`
	Shared         bool   `json:"shared"`
	Owner          string `json:"owner"`
	SharingDetails struct {
		LinkShared bool `json:"linkShared"`
		Published  bool `json:"published"`
	} `json:"sharingDetails"`
	DashboardFilter *DashboardFilter `json:"dashboardFilter,omitempty"`
	Tags            []string         `json:"tags"`
}

// DynatraceDashboard is struct for /dashboards/<dashboardID> endpoint
type DynatraceDashboard struct {
	Metadata struct {
		ConfigurationVersions []int  `json:"configurationVersions"`
		ClusterVersion        string `json:"clusterVersion"`
	} `json:"metadata"`
	ID                string            `json:"id"`
	DashboardMetadata DashboardMetadata `json:"dashboardMetadata"`
	Tiles             []Tile            `json:"tiles"`
}

// isTheSameAs Will validate if the this dashboard is the same as the one passed as parameter
func (dashboard *DynatraceDashboard) isTheSameAs(existingDashboardContent string) bool {

	jsonAsByteArray, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		log.WithError(err).Warn("Could not marshal dashboard")
	}
	newDashboardContent := string(jsonAsByteArray)

	// If ParseOnChange is not specified we consider this as a dashboard with a change
	if strings.Index(newDashboardContent, "KQG.QueryBehavior=ParseOnChange") == -1 {
		return false
	}

	// now lets compare the dashboard from the config repo and the one passed to this function
	if strings.Compare(newDashboardContent, existingDashboardContent) == 0 {
		return true
	}

	return false
}

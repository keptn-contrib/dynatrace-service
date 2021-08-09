package dynatrace

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

// DynatraceDashboard is struct for /dashboards/<dashboardID> endpoint
type DynatraceDashboard struct {
	Metadata struct {
		ConfigurationVersions []int  `json:"configurationVersions"`
		ClusterVersion        string `json:"clusterVersion"`
	} `json:"metadata"`
	ID                string `json:"id"`
	DashboardMetadata struct {
		Name           string `json:"name"`
		Shared         bool   `json:"shared"`
		Owner          string `json:"owner"`
		SharingDetails struct {
			LinkShared bool `json:"linkShared"`
			Published  bool `json:"published"`
		} `json:"sharingDetails"`
		DashboardFilter *struct {
			Timeframe      string `json:"timeframe"`
			ManagementZone *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"managementZone,omitempty"`
		} `json:"dashboardFilter,omitempty"`
		Tags []string `json:"tags"`
	} `json:"dashboardMetadata"`
	Tiles []struct {
		Name       string `json:"name"`
		TileType   string `json:"tileType"`
		Configured bool   `json:"configured"`
		Query      string `json:"query"`
		Type       string `json:"type"`
		CustomName string `json:"customName"`
		Markdown   string `json:"markdown"`
		Bounds     struct {
			Top    int `json:"top"`
			Left   int `json:"left"`
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"bounds"`
		TileFilter struct {
			Timeframe      string `json:"timeframe"`
			ManagementZone *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"managementZone,omitempty"`
		} `json:"tileFilter"`
		Queries          []DataExplorerQuery `json:"queries"`
		AssignedEntities []string            `json:"assignedEntities"`
		FilterConfig     struct {
			Type        string `json:"type"`
			CustomName  string `json:"customName"`
			DefaultName string `json:"defaultName"`
			ChartConfig struct {
				LegendShown    bool          `json:"legendShown"`
				Type           string        `json:"type"`
				Series         []ChartSeries `json:"series"`
				ResultMetadata struct {
				} `json:"resultMetadata"`
			} `json:"chartConfig"`
			FiltersPerEntityType map[string]map[string][]string `json:"filtersPerEntityType"`
			/* FiltersPerEntityType struct {
				HOST struct {
					SPECIFIC_ENTITIES    []string `json:"SPECIFIC_ENTITIES"`
					HOST_DATACENTERS     []string `json:"HOST_DATACENTERS"`
					AUTO_TAGS            []string `json:"AUTO_TAGS"`
					HOST_SOFTWARE_TECH   []string `json:"HOST_SOFTWARE_TECH"`
					HOST_VIRTUALIZATION  []string `json:"HOST_VIRTUALIZATION"`
					HOST_MONITORING_MODE []string `json:"HOST_MONITORING_MODE"`
					HOST_STATE           []string `json:"HOST_STATE"`
					HOST_HOST_GROUPS     []string `json:"HOST_HOST_GROUPS"`
				} `json:"HOST"`
				PROCESS_GROUP struct {
					SPECIFIC_ENTITIES     []string `json:"SPECIFIC_ENTITIES"`
					HOST_TAG_OF_PROCESS   []string `json:"HOST_TAG_OF_PROCESS"`
					AUTO_TAGS             []string `json:"AUTO_TAGS"`
					PROCESS_SOFTWARE_TECH []string `json:"PROCESS_SOFTWARE_TECH"`
				} `json:"PROCESS_GROUP"`
				PROCESS_GROUP_INSTANCE struct {
					SPECIFIC_ENTITIES     []string `json:"SPECIFIC_ENTITIES"`
					HOST_TAG_OF_PROCESS   []string `json:"HOST_TAG_OF_PROCESS"`
					AUTO_TAGS             []string `json:"AUTO_TAGS"`
					PROCESS_SOFTWARE_TECH []string `json:"PROCESS_SOFTWARE_TECH"`
				} `json:"PROCESS_GROUP_INSTANCE"`
				SERVICE struct {
					SPECIFIC_ENTITIES     []string `json:"SPECIFIC_ENTITIES"`
					SERVICE_SOFTWARE_TECH []string `json:"SERVICE_SOFTWARE_TECH"`
					AUTO_TAGS             []string `json:"AUTO_TAGS"`
					SERVICE_TYPE          []string `json:"SERVICE_TYPE"`
					SERVICE_TO_PG         []string `json:"SERVICE_TO_PG"`
				} `json:"SERVICE"`
				APPLICATION struct {
					SPECIFIC_ENTITIES          []string `json:"SPECIFIC_ENTITIES"`
					APPLICATION_TYPE           []string `json:"APPLICATION_TYPE"`
					AUTO_TAGS                  []string `json:"AUTO_TAGS"`
					APPLICATION_INJECTION_TYPE []string `json:"PROCESS_SOFTWARE_TECH"`
					APPLICATION_STATUS         []string `json:"APPLICATION_STATUS"`
				} `json:"APPLICATION"`
				APPLICATION_METHOD struct {
					SPECIFIC_ENTITIES []string `json:"SPECIFIC_ENTITIES"`
				} `json:"APPLICATION_METHOD"`
			} `json:"filtersPerEntityType"`*/
		} `json:"filterConfig"`
	} `json:"tiles"`
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

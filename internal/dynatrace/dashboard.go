package dynatrace

type Dashboard struct {
	Metadata          *Metadata         `json:"metadata,omitempty"`
	ID                string            `json:"id,omitempty"`
	DashboardMetadata DashboardMetadata `json:"dashboardMetadata"`
	Tiles             []Tile            `json:"tiles"`
}

type Metadata struct {
	ConfigurationVersions []int  `json:"configurationVersions,omitempty"`
	ClusterVersion        string `json:"clusterVersion,omitempty"`
}

type DashboardMetadata struct {
	Name            string           `json:"name"`
	Shared          bool             `json:"shared"`
	Owner           string           `json:"owner"`
	SharingDetails  SharingDetails   `json:"sharingDetails"`
	DashboardFilter *DashboardFilter `json:"dashboardFilter,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
}

type SharingDetails struct {
	LinkShared bool `json:"linkShared"`
	Published  bool `json:"published"`
}
type DashboardFilter struct {
	Timeframe      string               `json:"timeframe"`
	ManagementZone *ManagementZoneEntry `json:"managementZone,omitempty"`
}

type ManagementZoneEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Tile struct {
	Name                      string              `json:"name"`
	TileType                  string              `json:"tileType"`
	Configured                bool                `json:"configured"`
	Query                     string              `json:"query,omitempty"`
	Type                      string              `json:"type,omitempty"`
	CustomName                string              `json:"customName,omitempty"`
	Markdown                  string              `json:"markdown,omitempty"`
	ChartVisible              bool                `json:"chartVisible,omitempty"`
	Bounds                    Bounds              `json:"bounds"`
	TileFilter                TileFilter          `json:"tileFilter"`
	Queries                   []DataExplorerQuery `json:"queries,omitempty"`
	AssignedEntities          []string            `json:"assignedEntities,omitempty"`
	ExcludeMaintenanceWindows bool                `json:"excludeMaintenanceWindows,omitempty"`
	FilterConfig              *FilterConfig       `json:"filterConfig,omitempty"`
}

type Bounds struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type TileFilter struct {
	Timeframe      string               `json:"timeframe"`
	ManagementZone *ManagementZoneEntry `json:"managementZone,omitempty"`
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

type FilterConfig struct {
	Type                 string                         `json:"type"`
	CustomName           string                         `json:"customName"`
	DefaultName          string                         `json:"defaultName"`
	ChartConfig          ChartConfig                    `json:"chartConfig"`
	FiltersPerEntityType map[string]map[string][]string `json:"filtersPerEntityType"`
}

/*
FiltersPerEntityType struct {
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
} `json:"filtersPerEntityType"`
*/

type ChartConfig struct {
	LegendShown    bool           `json:"legendShown"`
	Type           string         `json:"type"`
	Series         []Series       `json:"series"`
	ResultMetadata ResultMetadata `json:"resultMetadata"`
}

type Series struct {
	Metric          string       `json:"metric"`
	Aggregation     string       `json:"aggregation"`
	Percentile      interface{}  `json:"percentile"`
	Type            string       `json:"type"`
	EntityType      string       `json:"entityType"`
	Dimensions      []Dimensions `json:"dimensions"`
	SortAscending   bool         `json:"sortAscending"`
	SortColumn      bool         `json:"sortColumn"`
	AggregationRate string       `json:"aggregationRate"`
}

type ResultMetadata struct {
}

type Dimensions struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Values          []string `json:"values"`
	EntityDimension bool     `json:"entityDimension"`
}

// Title custom chart and usql have different ways to define their tile names - so - lets figure it out by looking at the potential values
func (tile Tile) Title() string {
	if tile.FilterConfig != nil && tile.FilterConfig.CustomName != "" {
		return tile.FilterConfig.CustomName
	}

	if tile.CustomName != "" {
		return tile.CustomName
	}

	return tile.Name
}

// GetFilter returns the DashboardFilter
func (dashboard *Dashboard) GetFilter() *DashboardFilter {
	return dashboard.DashboardMetadata.DashboardFilter
}

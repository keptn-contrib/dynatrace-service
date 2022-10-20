package dynatrace

const (
	// CustomChartingTileType is the tile type for custom charting dashboard tiles
	CustomChartingTileType = "CUSTOM_CHARTING"

	// DataExplorerTileType is the tile type for data explorer dashboard tiles
	DataExplorerTileType = "DATA_EXPLORER"

	// MarkdownTileType is the tile type for markdown dashboard tiles
	MarkdownTileType = "MARKDOWN"

	// OpenProblemsTileType is the tile type for open problems dashboard tiles
	OpenProblemsTileType = "OPEN_PROBLEMS"

	// SLOTileType is the tile type for SLO dashboard tiles
	SLOTileType = "SLO"

	// USQLTileType is the tile type for USQL dashboard tiles
	USQLTileType = "DTAQL"
)

const (
	// ColumnChartVisualizationType is the column chart visualization type for USQL tiles
	ColumnChartVisualizationType = "COLUMN_CHART"

	// LineChartVisualizationType is the line chart visualization type for USQL tiles
	LineChartVisualizationType = "LINE_CHART"

	// PieChartVisualizationType is the pie chart visualization type for USQL tiles
	PieChartVisualizationType = "PIE_CHART"

	// SingleValueVisualizationType is the single value visualization type for USQL tiles
	SingleValueVisualizationType = "SINGLE_VALUE"

	// TableVisualizationType is the table visualization type for USQL tiles
	TableVisualizationType = "TABLE"
)

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
	Timeframe      string               `json:"timeframe,omitempty"`
	ManagementZone *ManagementZoneEntry `json:"managementZone,omitempty"`
}

type ManagementZoneEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Tile struct {
	Name                      string                      `json:"name"`
	TileType                  string                      `json:"tileType"`
	Configured                bool                        `json:"configured"`
	Query                     string                      `json:"query,omitempty"`
	Type                      string                      `json:"type,omitempty"`
	CustomName                string                      `json:"customName,omitempty"`
	Markdown                  string                      `json:"markdown,omitempty"`
	ChartVisible              bool                        `json:"chartVisible,omitempty"`
	Bounds                    Bounds                      `json:"bounds"`
	TileFilter                TileFilter                  `json:"tileFilter"`
	Queries                   []DataExplorerQuery         `json:"queries,omitempty"`
	AssignedEntities          []string                    `json:"assignedEntities,omitempty"`
	ExcludeMaintenanceWindows bool                        `json:"excludeMaintenanceWindows,omitempty"`
	FilterConfig              *FilterConfig               `json:"filterConfig,omitempty"`
	VisualConfig              *VisualizationConfiguration `json:"visualConfig,omitempty"`
	MetricExpressions         []string                    `json:"metricExpressions,omitempty"`
}

// VisualizationConfiguration is the visual configuration for a dashboard tile.
type VisualizationConfiguration struct {
	Type       string                   `json:"type,omitempty"`
	Thresholds []VisualizationThreshold `json:"thresholds,omitempty"`
	Rules      []VisualizationRule      `json:"rules,omitempty"`
}

// SingleValueVisualizationConfigurationType is the single value visualization type for VisualConfigs
const SingleValueVisualizationConfigurationType = "SINGLE_VALUE"

// VisualizationRule is a rule for the visual configuration.
type VisualizationRule struct {
	UnitTransform string `json:"unitTransform,omitempty"`
	Matcher       string `json:"matcher,omitempty"`
}

// VisualizationThreshold is a threshold configuration for a Data Explorer tile.
type VisualizationThreshold struct {
	Visible bool                         `json:"visible"`
	Rules   []VisualizationThresholdRule `json:"rules,omitempty"`
}

// VisualizationThresholdRule is a rule for a threshold.
type VisualizationThresholdRule struct {
	Value *float64 `json:"value,omitempty"`
	Color string   `json:"color"`
}

type Bounds struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type TileFilter struct {
	Timeframe      string               `json:"timeframe,omitempty"`
	ManagementZone *ManagementZoneEntry `json:"managementZone,omitempty"`
}

// DataExplorerQuery Query Definition for DATA_EXPLORER dashboard tile
type DataExplorerQuery struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type FilterConfig struct {
	Type                 string               `json:"type"`
	CustomName           string               `json:"customName"`
	DefaultName          string               `json:"defaultName"`
	ChartConfig          ChartConfig          `json:"chartConfig"`
	FiltersPerEntityType map[string]FilterMap `json:"filtersPerEntityType"`
}

type FilterMap map[string][]string

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

// GetFilter returns the DashboardFilter
func (dashboard *Dashboard) GetFilter() *DashboardFilter {
	return dashboard.DashboardMetadata.DashboardFilter
}

package lib

import (
	"math"

	keptnmodels "github.com/keptn/go-utils/pkg/models"
)

// CALCULATED METRIC TYPES
type CalculatedMetric struct {
	TsmMetricKey        string                       `json:"tsmMetricKey"`
	Name                string                       `json:"name"`
	Enabled             bool                         `json:"enabled"`
	MetricDefinition    MetricDefinition             `json:"metricDefinition"`
	Unit                string                       `json:"unit"`
	UnitDisplayName     string                       `json:"unitDisplayName"`
	Conditions          []CalculatedMetricConditions `json:"conditions"`
	DimensionDefinition DimensionDefinition          `json:"dimensionDefinition"`
}
type MetricDefinition struct {
	Metric           string      `json:"metric"`
	RequestAttribute interface{} `json:"requestAttribute"`
}
type Value struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}
type CalculatedMetricConditions struct {
	Attribute      string                         `json:"attribute"`
	ComparisonInfo CalculatedMetricComparisonInfo `json:"comparisonInfo"`
}
type CalculatedMetricComparisonInfo struct {
	Type             string `json:"type"`
	Comparison       string `json:"comparison"`
	Value            Value  `json:"value"`
	Negate           bool   `json:"negate"`
	RequestAttribute string `json:"requestAttribute"`
	CaseSensitive    bool   `json:"caseSensitive"`
}
type DimensionDefinition struct {
	Name            string   `json:"name"`
	Dimension       string   `json:"dimension"`
	Placeholders    []string `json:"placeholders"`
	TopX            int      `json:"topX"`
	TopXDirection   string   `json:"topXDirection"`
	TopXAggregation string   `json:"topXAggregation"`
}

// DASHBOARD TYPES
type DynatraceDashboard struct {
	DashboardMetadata DashboardMetadata `json:"dashboardMetadata"`
	Tiles             []Tiles           `json:"tiles"`
}
type SharingDetails struct {
	LinkShared bool `json:"linkShared"`
	Published  bool `json:"published"`
}
type DashboardFilter struct {
	Timeframe      string      `json:"timeframe"`
	ManagementZone interface{} `json:"managementZone"`
}
type DashboardMetadata struct {
	Name            string          `json:"name"`
	Shared          bool            `json:"shared"`
	Owner           string          `json:"owner"`
	SharingDetails  SharingDetails  `json:"sharingDetails"`
	DashboardFilter DashboardFilter `json:"dashboardFilter"`
}
type Bounds struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
type TileFilter struct {
	Timeframe      interface{} `json:"timeframe"`
	ManagementZone interface{} `json:"managementZone"`
}
type ResultMetadata struct {
}
type ChartConfig struct {
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
type Dimensions struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Values          []string `json:"values"`
	EntityDimension bool     `json:"entityDimension"`
}
type FiltersPerEntityType struct {
	NonDatabaseService EntityFilter `json:"NON_DATABASE_SERVICE,omitempty"`
	Service            EntityFilter `json:"SERVICE,omitempty"`
}
type EntityFilter struct {
	AutoTags []string `json:"AUTO_TAGS,omitempty"`
}
type FilterConfig struct {
	Type                 string               `json:"type"`
	CustomName           string               `json:"customName"`
	DefaultName          string               `json:"defaultName"`
	ChartConfig          ChartConfig          `json:"chartConfig"`
	FiltersPerEntityType FiltersPerEntityType `json:"filtersPerEntityType"`
}
type Tiles struct {
	Name                      string       `json:"name"`
	TileType                  string       `json:"tileType"`
	Configured                bool         `json:"configured"`
	Bounds                    Bounds       `json:"bounds"`
	TileFilter                TileFilter   `json:"tileFilter"`
	FilterConfig              FilterConfig `json:"filterConfig,omitempty"`
	ChartVisible              bool         `json:"chartVisible,omitempty"`
	AssignedEntities          []string     `json:"assignedEntities,omitempty"`
	ExcludeMaintenanceWindows bool         `json:"excludeMaintenanceWindows,omitempty"`
	Markdown                  string       `json:"markdown,omitempty"`
}

func CreateDynatraceDashboard(projectName string, shipyard keptnmodels.Shipyard, keptnDomain string, services []string) (*DynatraceDashboard, error) {
	dtDashboard := &DynatraceDashboard{
		DashboardMetadata: DashboardMetadata{
			Name:   projectName + "@keptn: Digital Delivery & Operations Dashboard",
			Shared: true,
			Owner:  "",
			SharingDetails: SharingDetails{
				LinkShared: true,
				Published:  false,
			},
			DashboardFilter: DashboardFilter{
				Timeframe:      "l_7_DAYS",
				ManagementZone: nil,
			},
		},
		Tiles: []Tiles{},
	}

	addTileToDashboard(createMarkdownTile("## Operations\n[Open Keptns Bridge](https://bridge.keptn."+keptnDomain+"/?#/)"), dtDashboard, true)

	// create stage service tiles
	for _, stage := range shipyard.Stages {
		addTileToDashboard(createHeaderTile(stage.Name), dtDashboard, true)
		addTileToDashboard(createStageServicesTile(projectName, stage.Name), dtDashboard, false)
		addTileToDashboard(createServiceThroughputTile(projectName, stage.Name), dtDashboard, false)
		addTileToDashboard(createServiceErrorRateTile(projectName, stage.Name), dtDashboard, false)
		addTileToDashboard(createServiceResponseTimeTile(projectName, stage.Name), dtDashboard, false)
		/*
			addTileToDashboard(createServiceTopAPICallsTile(projectName, stage.Name), dtDashboard, false)
			addTileToDashboard(createServiceTestStepTopAPICallsTile(projectName, stage.Name), dtDashboard, false)

		*/

		if len(services) > 0 {
			servicesMarkdown := "### Services: \n"
			for _, service := range services {
				servicesMarkdown = servicesMarkdown + "[" + service + "](http://" + service + "." + projectName + "-" + stage.Name + "." + keptnDomain + ")\n"
			}
			addTileToDashboard(createMarkdownTile(servicesMarkdown), dtDashboard, true)
		}
	}

	addTileToDashboard(createHeaderTile("Infrastructure"), dtDashboard, true)
	hostsTile := Tiles{
		Name:       "",
		TileType:   "HOSTS",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "HOST",
			CustomName:  "Hosts",
			DefaultName: "Hosts",
			ChartConfig: ChartConfig{
				Type:           "TIMESERIES",
				Series:         []Series{},
				ResultMetadata: ResultMetadata{},
			},
			FiltersPerEntityType: FiltersPerEntityType{},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
		Markdown:                  "",
	}

	addTileToDashboard(hostsTile, dtDashboard, false)
	addTileToDashboard(createHostCPULoadTile(), dtDashboard, false)

	return dtDashboard, nil
}

func CreateCalculatedMetric(key string, name string, baseMetric string, unit string, conditionContext string, conditionKey string, conditionValue string, dimensionName string, dimensionDefinition string, dimensionAggregate string) CalculatedMetric {
	return CalculatedMetric{
		TsmMetricKey:     key,
		Name:             name,
		Enabled:          true,
		MetricDefinition: MetricDefinition{},
		Unit:             unit,
		UnitDisplayName:  "",
		Conditions: []CalculatedMetricConditions{
			{
				Attribute: "SERVICE_TAG",
				ComparisonInfo: CalculatedMetricComparisonInfo{
					Type:       "TAG",
					Comparison: "TAG_KEY_EQUALS",
					Value: Value{
						Context: conditionContext,
						Key:     conditionKey,
						Value:   conditionValue,
					},
					Negate: false,
				},
			},
		},
		DimensionDefinition: DimensionDefinition{
			Name:            dimensionName,
			Dimension:       dimensionDefinition,
			Placeholders:    []string{},
			TopX:            10,
			TopXDirection:   "DESCENDING",
			TopXAggregation: dimensionAggregate,
		},
	}
}

func CreateCalculatedTestStepMetric(key string, name string, baseMetric string, unit string, conditionContext string, conditionKey string, conditionValue string, dimensionName string, dimensionDefinition string, dimensionAggregate string) CalculatedMetric {
	return CalculatedMetric{
		TsmMetricKey: key,
		Name:         name,
		Enabled:      true,
		MetricDefinition: MetricDefinition{
			Metric:           baseMetric,
			RequestAttribute: nil,
		},
		Unit:            unit,
		UnitDisplayName: "",
		Conditions: []CalculatedMetricConditions{
			{
				Attribute: "SERVICE_REQUEST_ATTRIBUTE",
				ComparisonInfo: CalculatedMetricComparisonInfo{
					Type:             "STRING_REQUEST_ATTRIBUTE",
					Comparison:       "EXISTS",
					Value:            Value{},
					Negate:           false,
					RequestAttribute: "TSN",
					CaseSensitive:    false,
				},
			},
			{
				Attribute: "SERVICE_TAG",
				ComparisonInfo: CalculatedMetricComparisonInfo{
					Type:       "TAG",
					Comparison: "TAG_KEY_EQUALS",
					Value: Value{
						Context: conditionContext,
						Key:     conditionKey,
						Value:   conditionValue,
					},
					Negate:           false,
					RequestAttribute: "TSN",
					CaseSensitive:    false,
				},
			},
		},
		DimensionDefinition: DimensionDefinition{
			Name:            "TestStep",
			Dimension:       "{RequestAttribute:TSN}",
			Placeholders:    []string{},
			TopX:            10,
			TopXDirection:   "DESCENDING",
			TopXAggregation: dimensionAggregate,
		},
	}
}

func createMarkdownTile(markdown string) Tiles {
	return Tiles{
		Name:       "Markdown",
		TileType:   "MARKDOWN",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		Markdown: markdown,
	}
}

func createHeaderTile(name string) Tiles {
	return Tiles{
		Name:       name,
		TileType:   "HEADER",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
	}
}

func createServiceResponseTimeTile(project string, stage string) Tiles {
	return Tiles{
		Name:       "Response Time " + stage,
		TileType:   "CUSTOM_CHARTING",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "MIXED",
			CustomName:  "Response Time " + stage,
			DefaultName: "Custom Chart",
			ChartConfig: ChartConfig{
				Type: "TIMESERIES",
				Series: []Series{
					{
						Metric:          "builtin:service.response.time",
						Aggregation:     "AVG",
						Percentile:      nil,
						Type:            "LINE",
						EntityType:      "SERVICE",
						Dimensions:      []Dimensions{},
						SortAscending:   false,
						SortColumn:      true,
						AggregationRate: "TOTAL",
					},
				},
			},
			FiltersPerEntityType: FiltersPerEntityType{
				Service: EntityFilter{
					AutoTags: []string{"keptn_project:" + project, "keptn_stage:" + stage},
				},
			},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
	}
}

func createHostCPULoadTile() Tiles {
	return Tiles{
		Name:       "Host CPU Load",
		TileType:   "CUSTOM_CHARTING",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "MIXED",
			CustomName:  "CPU",
			DefaultName: "Custom Chart",
			ChartConfig: ChartConfig{
				Type: "TIMESERIES",
				Series: []Series{
					{
						Metric:          "builtin:host.cpu.load",
						Aggregation:     "AVG",
						Percentile:      nil,
						Type:            "LINE",
						EntityType:      "HOST",
						Dimensions:      []Dimensions{},
						SortAscending:   false,
						SortColumn:      true,
						AggregationRate: "TOTAL",
					},
				},
			},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
	}
}

func createServiceErrorRateTile(project string, stage string) Tiles {
	return Tiles{
		Name:       "Failure Rate " + stage,
		TileType:   "CUSTOM_CHARTING",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "MIXED",
			CustomName:  "Failure Rate " + stage,
			DefaultName: "Custom Chart",
			ChartConfig: ChartConfig{
				Type: "TIMESERIES",
				Series: []Series{
					{
						Metric:          "builtin:service.errors.server.rate",
						Aggregation:     "AVG",
						Percentile:      nil,
						Type:            "BAR",
						EntityType:      "SERVICE",
						Dimensions:      []Dimensions{},
						SortAscending:   false,
						SortColumn:      true,
						AggregationRate: "TOTAL",
					},
				},
			},
			FiltersPerEntityType: FiltersPerEntityType{
				Service: EntityFilter{
					AutoTags: []string{"keptn_project:" + project, "keptn_stage:" + stage},
				},
			},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
	}
}

func createServiceTestStepTopAPICallsTile(project string, stage string) Tiles {
	return Tiles{
		Name:       "Service Calls per Test Name: " + stage,
		TileType:   "CUSTOM_CHARTING",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "MIXED",
			CustomName:  "Service Calls per Test Name: " + stage,
			DefaultName: "Custom Chart",
			ChartConfig: ChartConfig{
				Type: "TIMESERIES",
				Series: []Series{
					{
						Metric:      "calc:service.teststepservicecalls" + project,
						Aggregation: "NONE",
						Percentile:  nil,
						Type:        "BAR",
						EntityType:  "SERVICE",
						Dimensions: []Dimensions{
							{
								ID:              "1",
								Name:            "Test Step",
								Values:          []string{},
								EntityDimension: false,
							},
						},
						SortAscending:   false,
						SortColumn:      true,
						AggregationRate: "TOTAL",
					},
				},
			},
			FiltersPerEntityType: FiltersPerEntityType{
				NonDatabaseService: EntityFilter{
					AutoTags: []string{"keptn_project:" + project, "keptn_stage:" + stage},
				},
			},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
	}
}

func createServiceTopAPICallsTile(project string, stage string) Tiles {
	return Tiles{
		Name:       "Top Service Calls per API Endpoint: " + stage,
		TileType:   "CUSTOM_CHARTING",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "MIXED",
			CustomName:  "Top Service Calls per API Endpoint: " + stage,
			DefaultName: "Custom Chart",
			ChartConfig: ChartConfig{
				Type: "TIMESERIES",
				Series: []Series{
					{
						Metric:      "calc:service.topurlservicecalls" + project,
						Aggregation: "NONE",
						Percentile:  nil,
						Type:        "BAR",
						EntityType:  "SERVICE",
						Dimensions: []Dimensions{
							{
								ID:              "1",
								Name:            "URL",
								Values:          []string{},
								EntityDimension: false,
							},
						},
						SortAscending:   false,
						SortColumn:      true,
						AggregationRate: "TOTAL",
					},
				},
			},
			FiltersPerEntityType: FiltersPerEntityType{
				NonDatabaseService: EntityFilter{
					AutoTags: []string{"keptn_project:" + project, "keptn_stage:" + stage},
				},
			},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
	}
}

func createServiceThroughputTile(project string, stage string) Tiles {
	return Tiles{
		Name:       "Throughput " + stage,
		TileType:   "CUSTOM_CHARTING",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "MIXED",
			CustomName:  "Throughput " + stage,
			DefaultName: "Custom Chart",
			ChartConfig: ChartConfig{
				Type: "TIMESERIES",
				Series: []Series{
					{
						Metric:          "builtin:service.requestCount.total",
						Aggregation:     "NONE",
						Percentile:      nil,
						Type:            "BAR",
						EntityType:      "SERVICE",
						Dimensions:      []Dimensions{},
						SortAscending:   false,
						SortColumn:      true,
						AggregationRate: "TOTAL",
					},
				},
			},
			FiltersPerEntityType: FiltersPerEntityType{
				NonDatabaseService: EntityFilter{
					AutoTags: []string{"keptn_project:" + project, "keptn_stage:" + stage},
				},
			},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
	}
}

func createStageServicesTile(project string, stage string) Tiles {
	return Tiles{
		Name:       "Services: " + stage,
		TileType:   "SERVICES",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "NON_DATABASE_SERVICE",
			CustomName:  "Services: " + stage,
			DefaultName: "Services: " + stage,
			ChartConfig: ChartConfig{
				Type:           "TIMESERIES",
				Series:         []Series{},
				ResultMetadata: ResultMetadata{},
			},
			FiltersPerEntityType: FiltersPerEntityType{
				NonDatabaseService: EntityFilter{
					AutoTags: []string{"keptn_project:" + project, "keptn_stage:" + stage},
				},
			},
		},
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
		Markdown:                  "",
	}
}

func addTileToDashboard(tile Tiles, dashboard *DynatraceDashboard, useFullRow bool) {
	topOffset := 76.0
	numberOfColumns := 3
	tileWidth := 304.0
	tileHeight := 152.0

	usedSpace := 0.0

	for _, tile := range dashboard.Tiles {
		usedSpace += float64(tile.Bounds.Width)
	}

	numberOfTiles := int(usedSpace / tileWidth)

	if useFullRow {
		mod := numberOfTiles % numberOfColumns
		if mod != 0 {
			placesToFill := numberOfColumns - mod
			for i := 0; i < placesToFill; i++ {
				addTileToDashboard(createMarkdownTile(" "), dashboard, false)
			}
			addTileToDashboard(tile, dashboard, useFullRow)
			return
		}
	}

	top := math.Floor(usedSpace/(float64(numberOfColumns)*tileWidth))*tileHeight + topOffset

	left := float64(numberOfTiles%numberOfColumns) * tileWidth

	tile.Bounds = Bounds{
		Top:    int(top),
		Left:   int(left),
		Width:  int(tileWidth),
		Height: int(tileHeight),
	}

	if useFullRow {
		tile.Bounds.Width = int(float64(numberOfColumns) * tileWidth)
	}

	dashboard.Tiles = append(dashboard.Tiles, tile)
}

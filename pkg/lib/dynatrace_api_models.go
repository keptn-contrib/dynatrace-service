package lib

import (
	"math"

	keptnmodels "github.com/keptn/go-utils/pkg/models"
)

// PROBLEM NOTIFICATION

const PROBLEM_NOTIFICATION_PAYLOAD string = `{ 
      "type": "WEBHOOK", 
      "name": "Keptn Problem Notification", 
      "alertingProfile": "$ALERTING_PROFILE_ID", 
      "active": true, 
      "url": "$KEPTN_DNS/v1/event", 
      "acceptAnyCertificate": true, 
      "headers": [ 
        { "name": "x-token", "value": "$KEPTN_TOKEN" },
        { "name": "Content-Type", "value": "application/cloudevents+json" }
      ],
      "payload": "{\n    \"specversion\":\"0.2\",\n    \"type\":\"sh.keptn.events.problem\",\n    \"shkeptncontext\":\"{PID}\",\n    \"source\":\"dynatrace\",\n    \"id\":\"{PID}\",\n    \"time\":\"\",\n    \"contenttype\":\"application/json\",\n    \"data\": {\n        \"State\":\"{State}\",\n        \"ProblemID\":\"{ProblemID}\",\n        \"PID\":\"{PID}\",\n        \"ProblemTitle\":\"{ProblemTitle}\",\n        \"ProblemDetails\":{ProblemDetailsJSON},\n        \"Tags\":\"{Tags}\",\n        \"ImpactedEntities\":{ImpactedEntities},\n        \"ImpactedEntity\":\"{ImpactedEntity}\"\n    }\n}\n" 

      }`

const DASHBOARD_STAGE_WIDTH int = 456

// ALERTING PROFILE TYPES
type AlertingProfile struct {
	Metadata         AlertingProfileMetadata `json:"metadata"`
	ID               string                  `json:"id"`
	DisplayName      string                  `json:"displayName"`
	Rules            []AlertingProfileRules  `json:"rules"`
	ManagementZoneID interface{}             `json:"managementZoneId"`
	EventTypeFilters []string                `json:"eventTypeFilters"`
}
type AlertingProfileMetadata struct {
	ConfigurationVersions []int  `json:"configurationVersions"`
	ClusterVersion        string `json:"clusterVersion"`
}
type AlertingProfileTagFilter struct {
	IncludeMode string   `json:"includeMode"`
	TagFilters  []string `json:"tagFilters"`
}
type AlertingProfileRules struct {
	SeverityLevel  string                   `json:"severityLevel"`
	TagFilter      AlertingProfileTagFilter `json:"tagFilter"`
	DelayInMinutes int                      `json:"delayInMinutes"`
}

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

// MANAGEMENT ZONE TYPES
type ManagementZone struct {
	Name  string    `json:"name"`
	Rules []MZRules `json:"rules"`
}

type MZKey struct {
	Attribute string `json:"attribute"`
}
type MZValue struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}
type MZComparisonInfo struct {
	Type     string  `json:"type"`
	Operator string  `json:"operator"`
	Value    MZValue `json:"value"`
	Negate   bool    `json:"negate"`
}
type MZConditions struct {
	Key            MZKey            `json:"key"`
	ComparisonInfo MZComparisonInfo `json:"comparisonInfo"`
}
type MZRules struct {
	Type             string         `json:"type"`
	Enabled          bool           `json:"enabled"`
	PropagationTypes []string       `json:"propagationTypes"`
	Conditions       []MZConditions `json:"conditions"`
}

// AUTO TAGGING
type DTTaggingRule struct {
	Name  string  `json:"name"`
	Rules []Rules `json:"rules"`
}
type DynamicKey struct {
	Source string `json:"source"`
	Key    string `json:"key"`
}
type Key struct {
	Attribute  string     `json:"attribute"`
	DynamicKey DynamicKey `json:"dynamicKey"`
	Type       string     `json:"type"`
}
type ComparisonInfo struct {
	Type          string      `json:"type"`
	Operator      string      `json:"operator"`
	Value         interface{} `json:"value"`
	Negate        bool        `json:"negate"`
	CaseSensitive interface{} `json:"caseSensitive"`
}
type Conditions struct {
	Key            Key            `json:"key"`
	ComparisonInfo ComparisonInfo `json:"comparisonInfo"`
}
type Rules struct {
	Type             string       `json:"type"`
	Enabled          bool         `json:"enabled"`
	ValueFormat      string       `json:"valueFormat"`
	PropagationTypes []string     `json:"propagationTypes"`
	Conditions       []Conditions `json:"conditions"`
}

type DTDashboardsResponse struct {
	Dashboards []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Owner string `json:"owner"`
	} `json:"dashboards"`
}

func CreateKeptnAlertingProfile() *AlertingProfile {
	return &AlertingProfile{
		Metadata:    AlertingProfileMetadata{},
		DisplayName: "Keptn",
		Rules: []AlertingProfileRules{
			{
				SeverityLevel: "AVAILABILITY",
				TagFilter: AlertingProfileTagFilter{
					IncludeMode: "NONE",
					TagFilters:  nil,
				},
				DelayInMinutes: 0,
			},
			{
				SeverityLevel: "ERROR",
				TagFilter: AlertingProfileTagFilter{
					IncludeMode: "NONE",
					TagFilters:  nil,
				},
				DelayInMinutes: 0,
			},
			{
				SeverityLevel: "PERFORMANCE",
				TagFilter: AlertingProfileTagFilter{
					IncludeMode: "NONE",
					TagFilters:  nil,
				},
				DelayInMinutes: 0,
			},
			{
				SeverityLevel: "RESOURCE_CONTENTION",
				TagFilter: AlertingProfileTagFilter{
					IncludeMode: "NONE",
					TagFilters:  nil,
				},
				DelayInMinutes: 0,
			},
			{
				SeverityLevel: "CUSTOM_ALERT",
				TagFilter: AlertingProfileTagFilter{
					IncludeMode: "NONE",
					TagFilters:  nil,
				},
				DelayInMinutes: 0,
			},
			{
				SeverityLevel: "MONITORING_UNAVAILABLE",
				TagFilter: AlertingProfileTagFilter{
					IncludeMode: "NONE",
					TagFilters:  nil,
				},
				DelayInMinutes: 0,
			},
		},
		ManagementZoneID: nil,
		EventTypeFilters: nil,
	}
}

func CreateManagementZoneForProject(project string) *ManagementZone {
	managementZone := &ManagementZone{
		Name: "Keptn: " + project,
		Rules: []MZRules{
			{
				Type:             "SERVICE",
				Enabled:          true,
				PropagationTypes: []string{},
				Conditions: []MZConditions{
					{
						Key: MZKey{
							Attribute: "SERVICE_TAGS",
						},
						ComparisonInfo: MZComparisonInfo{
							Type:     "TAG",
							Operator: "EQUALS",
							Value: MZValue{
								Context: "CONTEXTLESS",
								Key:     "keptn_project",
								Value:   project,
							},
							Negate: false,
						},
					},
				},
			},
		},
	}

	return managementZone
}

func CreateManagementZoneForStage(project string, stage string) *ManagementZone {
	managementZone := &ManagementZone{
		Name: "Keptn: " + project + " " + stage,
		Rules: []MZRules{
			{
				Type:             "SERVICE",
				Enabled:          true,
				PropagationTypes: []string{},
				Conditions: []MZConditions{
					{
						Key: MZKey{
							Attribute: "SERVICE_TAGS",
						},
						ComparisonInfo: MZComparisonInfo{
							Type:     "TAG",
							Operator: "EQUALS",
							Value: MZValue{
								Context: "CONTEXTLESS",
								Key:     "keptn_project",
								Value:   project,
							},
							Negate: false,
						},
					},
					{
						Key: MZKey{
							Attribute: "SERVICE_TAGS",
						},
						ComparisonInfo: MZComparisonInfo{
							Type:     "TAG",
							Operator: "EQUALS",
							Value: MZValue{
								Context: "CONTEXTLESS",
								Key:     "keptn_stage",
								Value:   stage,
							},
							Negate: false,
						},
					},
				},
			},
		},
	}

	return managementZone
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

	infrastructureHeaderTile := createHeaderTile("Infrastructure")
	infrastructureHeaderTile.Bounds = Bounds{
		Top:    0,
		Left:   0,
		Width:  500,
		Height: 38,
	}

	dtDashboard.Tiles = append(dtDashboard.Tiles, infrastructureHeaderTile)

	hostsTile := Tiles{
		Name:       "",
		TileType:   "HOSTS",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig: FilterConfig{
			Type:        "MIXED",
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
		Bounds: Bounds{
			Top:    38,
			Left:   0,
			Width:  DASHBOARD_STAGE_WIDTH,
			Height: 152,
		},
	}
	dtDashboard.Tiles = append(dtDashboard.Tiles, hostsTile)

	networkTile := Tiles{
		Name:       "Network Status",
		TileType:   "NETWORK_MEDIUM",
		Configured: true,
		TileFilter: TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		AssignedEntities: nil,
		Bounds: Bounds{
			Top:    38,
			Left:   912,
			Width:  DASHBOARD_STAGE_WIDTH,
			Height: 152,
		},
	}
	dtDashboard.Tiles = append(dtDashboard.Tiles, networkTile)

	cpuLoadTile := createHostCPULoadTile()
	cpuLoadTile.Bounds = Bounds{
		Top:    38,
		Left:   DASHBOARD_STAGE_WIDTH,
		Width:  DASHBOARD_STAGE_WIDTH,
		Height: 152,
	}
	dtDashboard.Tiles = append(dtDashboard.Tiles, cpuLoadTile)

	bridgeTile := createMarkdownTile("## Operations\n[Open Keptns Bridge](https://bridge.keptn." + keptnDomain + "/?#/)")
	bridgeTile.Bounds = Bounds{
		Top:    190,
		Left:   0,
		Width:  912,
		Height: 76,
	}
	dtDashboard.Tiles = append(dtDashboard.Tiles, bridgeTile)

	// create stage service tiles
	for index, stage := range shipyard.Stages {

		headerTile := createHeaderTile(stage.Name)
		headerTile.Bounds = Bounds{
			Top:    266,
			Left:   index * DASHBOARD_STAGE_WIDTH,
			Width:  DASHBOARD_STAGE_WIDTH,
			Height: 38,
		}
		dtDashboard.Tiles = append(dtDashboard.Tiles, headerTile)

		servicesTile := createStageServicesTile(projectName, stage.Name)
		servicesTile.Bounds = Bounds{
			Top:    304,
			Left:   index * DASHBOARD_STAGE_WIDTH,
			Width:  DASHBOARD_STAGE_WIDTH,
			Height: 152,
		}
		dtDashboard.Tiles = append(dtDashboard.Tiles, servicesTile)

		throughputTile := createServiceThroughputTile(projectName, stage.Name)
		throughputTile.Bounds = Bounds{
			Top:    456,
			Left:   index * DASHBOARD_STAGE_WIDTH,
			Width:  DASHBOARD_STAGE_WIDTH,
			Height: 152,
		}
		dtDashboard.Tiles = append(dtDashboard.Tiles, throughputTile)

		errorRateTile := createServiceErrorRateTile(projectName, stage.Name)
		errorRateTile.Bounds = Bounds{
			Top:    608,
			Left:   index * DASHBOARD_STAGE_WIDTH,
			Width:  DASHBOARD_STAGE_WIDTH,
			Height: 152,
		}
		dtDashboard.Tiles = append(dtDashboard.Tiles, errorRateTile)

		responseTimeTile := createServiceResponseTimeTile(projectName, stage.Name)
		responseTimeTile.Bounds = Bounds{
			Top:    760,
			Left:   index * DASHBOARD_STAGE_WIDTH,
			Width:  DASHBOARD_STAGE_WIDTH,
			Height: 152,
		}
		dtDashboard.Tiles = append(dtDashboard.Tiles, responseTimeTile)

		if len(services) > 0 {
			servicesMarkdown := "### Services: \n"
			for _, service := range services {
				servicesMarkdown = servicesMarkdown + "[" + service + "](http://" + service + "." + projectName + "-" + stage.Name + "." + keptnDomain + ")\n"
			}
			servicesMdTile := createMarkdownTile(servicesMarkdown)
			servicesMdTile.Bounds = Bounds{
				Top:    912,
				Left:   index * DASHBOARD_STAGE_WIDTH,
				Width:  DASHBOARD_STAGE_WIDTH,
				Height: 1000,
			}
			dtDashboard.Tiles = append(dtDashboard.Tiles, servicesMdTile)
		}
	}

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
			Type:        "MIXED",
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

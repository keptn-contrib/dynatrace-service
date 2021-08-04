package dynatrace

type MetricQueryResultNumbers struct {
	Dimensions   []string          `json:"dimensions"`
	DimensionMap map[string]string `json:"dimensionMap,omitempty"`
	Timestamps   []int64           `json:"timestamps"`
	Values       []float64         `json:"values"`
}

type MetricQueryResultValues struct {
	MetricID string                     `json:"metricId"`
	Data     []MetricQueryResultNumbers `json:"data"`
}

// DTUSQLResult struct
type DTUSQLResult struct {
	ExtrapolationLevel int             `json:"extrapolationLevel"`
	ColumnNames        []string        `json:"columnNames"`
	Values             [][]interface{} `json:"values"`
}

// SLI struct for SLI.yaml
type SLI struct {
	SpecVersion string            `yaml:"spec_version"`
	Indicators  map[string]string `yaml:"indicators"`
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

// MetricDefinition defines the output of /metrics/<metricID>
type MetricDefinition struct {
	MetricID           string   `json:"metricId"`
	DisplayName        string   `json:"displayName"`
	Description        string   `json:"description"`
	Unit               string   `json:"unit"`
	AggregationTypes   []string `json:"aggregationTypes"`
	Transformations    []string `json:"transformations"`
	DefaultAggregation struct {
		Type string `json:"type"`
	} `json:"defaultAggregation"`
	DimensionDefinitions []struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Key         string `json:"key"`
		DisplayName string `json:"displayName"`
	} `json:"dimensionDefinitions"`
	EntityType []string `json:"entityType"`
}

type DynatraceSLOResult struct {
	ID                  string  `json:"id"`
	Enabled             bool    `json:"enabled"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	EvaluatedPercentage float64 `json:"evaluatedPercentage"`
	ErrorBudget         float64 `json:"errorBudget"`
	Status              string  `json:"status"`
	Error               string  `json:"error"`
	UseRateMetric       bool    `json:"useRateMetric"`
	MetricRate          string  `json:"metricRate"`
	MetricNumerator     string  `json:"metricNumerator"`
	MetricDenominator   string  `json:"metricDenominator"`
	TargetSuccessOLD    float64 `json:"targetSuccess"`
	TargetWarningOLD    float64 `json:"targetWarning"`
	Target              float64 `json:"target"`
	Warning             float64 `json:"warning"`
	EvaluationType      string  `json:"evaluationType"`
	TimeWindow          string  `json:"timeWindow"`
	Filter              string  `json:"filter"`
}

type DtEnvAPIv2Error struct {
	Error struct {
		Code                 int    `json:"code"`
		Message              string `json:"message"`
		ConstraintViolations []struct {
			Path              string `json:"path"`
			Message           string `json:"message"`
			ParameterLocation string `json:"parameterLocation"`
			Location          string `json:"location"`
		} `json:"constraintViolations"`
	} `json:"error"`
}

/**
{
    "totalCount": 8,
    "nextPageKey": null,
    "result": [
        {
            "metricId": "builtin:service.response.time:percentile(50):merge(0)",
            "data": [
                {
                    "dimensions": [],
                    "timestamps": [
                        1579097520000
                    ],
                    "values": [
                        65005.48481639812
                    ]
                }
            ]
        }
    ]
}
*/

// DynatraceMetricsQueryResult is struct for /metrics/query
type DynatraceMetricsQueryResult struct {
	TotalCount  int                       `json:"totalCount"`
	NextPageKey string                    `json:"nextPageKey"`
	Result      []MetricQueryResultValues `json:"result"`
}

// DynatraceProblem Problem Detail returned by /api/v2/problems
type DynatraceProblem struct {
	ProblemID        string `json:"problemId"`
	DisplayID        string `json:"displayId"`
	Title            string `json:"title"`
	ImpactLevel      string `json:"impactLevel"`
	SeverityLevel    string `json:"severityLevel"`
	Status           string `json:"status"`
	AffectedEntities []struct {
		EntityID struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"entityId"`
		Name string `json:"name"`
	} `json:"affectedEntities"`
	ImpactedEntities []struct {
		EntityID struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"entityId"`
		Name string `json:"name"`
	} `json:"impactedEntities"`
	RootCauseEntity struct {
		EntityID struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"entityId"`
		Name string `json:"name"`
	} `json:"rootCauseEntity"`
	ManagementZones []interface{} `json:"managementZones"`
	EntityTags      []struct {
		Context              string `json:"context"`
		Key                  string `json:"key"`
		Value                string `json:"value"`
		StringRepresentation string `json:"stringRepresentation"`
	} `json:"entityTags"`
	ProblemFilters []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"problemFilters"`
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

// DynatraceSecurityProblem Problem Detail returned by /api/v2/securityProblems
type DynatraceSecurityProblem struct {
	SecurityProblemID    string `json:"securityProblemId"`
	DisplayID            int    `json:"displayId"`
	State                string `json:"state"`
	VulnerabilityID      string `json:"vulnerabilityId"`
	VulnerabilityType    string `json:"vulnerabilityType"`
	FirstSeenTimestamp   int    `json:"firstSeenTimestamp"`
	LastUpdatedTimestamp int    `json:"lastUpdatedTimestamp"`
	RiskAssessment       struct {
		RiskCategory string `json:"riskCategory"`
		RiskScore    struct {
			Value int `json:"value"`
		} `json:"riskScore"`
		Exposed                bool `json:"exposed"`
		SensitiveDataAffected  bool `json:"sensitiveDataAffected"`
		PublicExploitAvailable bool `json:"publicExploitAvailable"`
	} `json:"riskAssessment"`
	ManagementZones      []string `json:"managementZones"`
	VulnerableComponents []struct {
		ID                          string   `json:"id"`
		DisplayName                 string   `json:"displayName"`
		FileName                    string   `json:"fileName"`
		NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
		VulnerableProcesses         []string `json:"vulnerableProcesses"`
	} `json:"vulnerableComponents"`
	VulnerableEntities  []string `json:"vulnerableEntities"`
	ExposedEntities     []string `json:"exposedEntities"`
	SensitiveDataAssets []string `json:"sensitiveDataAssets"`
	AffectedEntities    struct {
		Applications []struct {
			ID                          string   `json:"id"`
			NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
			VulnerableProcesses         []string `json:"vulnerableProcesses"`
		} `json:"applications"`
		Services []struct {
			ID                          string   `json:"id"`
			NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
			VulnerableProcesses         []string `json:"vulnerableProcesses"`
		} `json:"services"`
		Hosts []struct {
			ID                          string   `json:"id"`
			NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
			VulnerableProcesses         []string `json:"vulnerableProcesses"`
		} `json:"hosts"`
		Databases []string `json:"databases"`
	} `json:"affectedEntities"`
}

// DynatraceProblemQueryResult Result of /api/v1/problems
type DynatraceProblemQueryResult struct {
	TotalCount int                `json:"totalCount"`
	PageSize   int                `json:"pageSize"`
	Problems   []DynatraceProblem `json:"problems"`
}

// DynatraceSecurityProblemQueryResult Result of/api/v2/securityProblems
type DynatraceSecurityProblemQueryResult struct {
	TotalCount       int                        `json:"totalCount"`
	PageSize         int                        `json:"pageSize"`
	NextPageKey      string                     `json:"nextPageKey"`
	SecurityProblems []DynatraceSecurityProblem `json:"securityProblems"`
}

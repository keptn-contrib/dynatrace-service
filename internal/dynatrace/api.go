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

// MetricsQueryResult is struct for /metrics/query
type MetricsQueryResult struct {
	TotalCount  int                       `json:"totalCount"`
	NextPageKey string                    `json:"nextPageKey"`
	Result      []MetricQueryResultValues `json:"result"`
}

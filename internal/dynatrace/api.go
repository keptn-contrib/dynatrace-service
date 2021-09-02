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

// SLI struct for SLI.yaml
type SLI struct {
	SpecVersion string            `yaml:"spec_version"`
	Indicators  map[string]string `yaml:"indicators"`
}

// MetricsQueryResult is struct for /metrics/query
type MetricsQueryResult struct {
	TotalCount  int                       `json:"totalCount"`
	NextPageKey string                    `json:"nextPageKey"`
	Result      []MetricQueryResultValues `json:"result"`
}

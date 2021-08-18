package dynatrace

type MetricQueryComponents struct {
	metricID                      string
	metricUnit                    string
	metricQuery                   string
	fullMetricQuery               string
	entitySelectorSLIDefinition   string
	filterSLIDefinitionAggregator string
}

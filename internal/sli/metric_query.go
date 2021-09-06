package sli

type MetricQueryComponents struct {
	metricID                      string
	metricUnit                    string
	metricQuery                   string
	fullMetricQueryString         string
	entitySelectorSLIDefinition   string
	filterSLIDefinitionAggregator string
}

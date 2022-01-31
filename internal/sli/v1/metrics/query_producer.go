package metrics

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

// QueryProducer for metrics queries.
type QueryProducer struct {
	query metrics.Query
}

// NewQueryProducer creates a QueryProducer the specified metrics Query.
func NewQueryProducer(query metrics.Query) QueryProducer {
	return QueryProducer{query: query}
}

// Produce returns the unencoded metrics query string for a Query.
func (b QueryProducer) Produce() string {
	keyValues := make(map[string]string, 2)
	keyValues[metricSelectorKey] = b.query.GetMetricSelector()
	if b.query.GetEntitySelector() != "" {
		keyValues[entitySelectorKey] = b.query.GetEntitySelector()
	}
	return common.NewSLIProducer(common.NewKeyValuePairs(keyValues)).Produce()
}

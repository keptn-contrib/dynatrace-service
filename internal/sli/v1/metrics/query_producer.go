package metrics

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

type QueryProducer struct {
	query *metrics.Query
}

func NewQueryProducer(query *metrics.Query) *QueryProducer {
	return &QueryProducer{query: query}
}

// Produce builds a short metric query string from a Query with metricSelector and optional entitySelector
// It does not encode the parameters
// Returns:
//  #1: Dynatrace API metric query string
//  #2: error
func (b *QueryProducer) Produce() (string, error) {
	rawParameters := make(map[string]string, 2)
	rawParameters[metricSelectorKey] = b.query.GetMetricSelector()
	if b.query.GetEntitySelector() != "" {
		rawParameters[entitySelectorKey] = b.query.GetEntitySelector()
	}
	return common.NewSLIProducer(common.NewKeyValuePairs(rawParameters), b).Produce()
}

func (b *QueryProducer) GetKeyPosition(key string) (int, bool) {
	switch key {
	case metricSelectorKey:
		return 0, true
	case entitySelectorKey:
		return 1, true
	default:
		return 0, false
	}
}

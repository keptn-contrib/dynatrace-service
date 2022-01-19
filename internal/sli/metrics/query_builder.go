package metrics

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/parser"
)

type queryBuilder struct {
	query *Query
}

func newQueryBuilder(query *Query) *queryBuilder {
	return &queryBuilder{query: query}
}

// Build builds a short metric query string from a Query with metricSelector and optional entitySelector
// It does not encode the parameters
// Returns:
//  #1: Dynatrace API metric query string
//  #2: error
func (b *queryBuilder) build() (string, error) {
	rawParameters := make(map[string]string, 2)
	rawParameters[metricSelectorKey] = b.query.GetMetricSelector()
	if b.query.GetEntitySelector() != "" {
		rawParameters[entitySelectorKey] = b.query.GetEntitySelector()
	}
	return parser.NewQueryBuilder(parser.NewQueryParameters(rawParameters), b).Build()
}

func (b *queryBuilder) GetKeyPosition(key string) (int, bool) {
	switch key {
	case metricSelectorKey:
		return 0, true
	case entitySelectorKey:
		return 1, true
	default:
		return 0, false
	}
}

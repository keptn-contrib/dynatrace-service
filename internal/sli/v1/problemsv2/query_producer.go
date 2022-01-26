package problemsv2

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

// QueryProducer for problems v2 queries.
type QueryProducer struct {
	query problems.Query
}

// NewQueryProducer creates a QueryProducer the specified metrics Query.
func NewQueryProducer(query problems.Query) QueryProducer {
	return QueryProducer{query: query}
}

// Produce returns the unencoded metrics query string for a Query.
func (p QueryProducer) Produce() string {
	keyValues := make(map[string]string, 2)
	if p.query.GetProblemSelector() != "" {
		keyValues[problemSelectorKey] = p.query.GetProblemSelector()
	}

	if p.query.GetEntitySelector() != "" {
		keyValues[entitySelectorKey] = p.query.GetEntitySelector()
	}
	return ProblemsV2Prefix + common.NewSLIProducer(common.NewKeyValuePairs(keyValues)).Produce()
}

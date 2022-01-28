package mv2

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/metrics"
)

// QueryProducer for MV2 queries.
type QueryProducer struct {
	query Query
}

// NewQueryProducer creates a QueryProducer for the specified MV2 Query.
func NewQueryProducer(query Query) QueryProducer {
	return QueryProducer{query: query}
}

// Produce returns the MV2 query string for a Query.
func (p QueryProducer) Produce() string {
	return common.ProducePrefixedSLI(MV2Prefix, p.query.unit, metrics.NewQueryProducer(p.query.GetQuery()).Produce())
}

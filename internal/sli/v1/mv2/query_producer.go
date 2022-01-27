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
	pieces := make([]string, 0, 3)
	pieces = append(pieces, MV2Prefix)
	pieces = append(pieces, p.query.unit)
	pieces = append(pieces, metrics.NewQueryProducer(p.query.GetQuery()).Produce())

	return common.NewSLIPrefixProducer(common.NewSLIPieces(pieces)).Produce()
}

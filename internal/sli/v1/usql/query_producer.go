package usql

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

// QueryProducer for USQL queries.
type QueryProducer struct {
	query Query
}

// NewQueryProducer creates a QueryProducer for the specified USQL Query.
func NewQueryProducer(query Query) QueryProducer {
	return QueryProducer{query: query}
}

// Produce returns USQL query string for a Query.
func (p QueryProducer) Produce() string {
	return common.ProducePrefixedSLI(USQLPrefix, p.query.GetResultType(), p.query.GetDimension(), p.query.GetQuery().GetQuery())
}

package slo

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

// QueryProducer for SLO queries.
type QueryProducer struct {
	query Query
}

// NewQueryProducer creates a QueryProducer for the specified SLO Query.
func NewQueryProducer(query Query) QueryProducer {
	return QueryProducer{query: query}
}

// Produce returns SLO query string for a Query.
func (p QueryProducer) Produce() string {
	return common.ProducePrefixedSLI(SLOPrefix, p.query.GetSLOID())
}

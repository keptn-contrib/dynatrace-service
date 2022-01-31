package secpv2

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/secpv2"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

// QueryProducer for security problems v2 queries.
type QueryProducer struct {
	query secpv2.Query
}

// NewQueryProducer creates a QueryProducer for the specified security problems Query.
func NewQueryProducer(query secpv2.Query) QueryProducer {
	return QueryProducer{query: query}
}

// Produce returns security problems v2 query string for a Query.
func (p QueryProducer) Produce() string {
	keyValues := make(map[string]string, 2)
	if p.query.GetSecurityProblemSelector() != "" {
		keyValues[securityProblemSelectorKey] = p.query.GetSecurityProblemSelector()
	}

	return common.ProducePrefixedSLI(SecurityProblemsV2Prefix, common.NewSLIProducer(common.NewKeyValuePairs(keyValues)).Produce())
}

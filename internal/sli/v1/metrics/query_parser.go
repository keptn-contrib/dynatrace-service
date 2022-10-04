package metrics

import (
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

const (
	metricSelectorKey = "metricSelector"
	entitySelectorKey = "entitySelector"
	resolutionKey     = "resolution"
	mzSelectorKey     = "mzSelector"
)

// QueryParser will parse an un-encoded metrics query string (usually found in sli.yaml files) into a Query
type QueryParser struct {
	query string
}

// NewQueryParser creates a new QueryParser for the specified query string.
func NewQueryParser(query string) *QueryParser {
	return &QueryParser{
		query: strings.TrimSpace(query),
	}
}

// Parse parses an un-encoded metrics query string (usually found in sli.yaml files) into a Query or returns an error.
// It only supports the current Metrics API V2 format (without a '?' prefix)
func (p *QueryParser) Parse() (*metrics.Query, error) {
	keyValuePairs, err := common.NewSLIParser(p.query, &metricsQueryKeyValidator{}).Parse()
	if err != nil {
		return nil, err
	}
	return metrics.NewQueryWithResolutionAndMZSelector(keyValuePairs.GetValue(metricSelectorKey), keyValuePairs.GetValue(entitySelectorKey), keyValuePairs.GetValue(resolutionKey), keyValuePairs.GetValue(mzSelectorKey))
}

type metricsQueryKeyValidator struct{}

// ValidateKey returns true if the specified key is part of a metrics query.
func (p *metricsQueryKeyValidator) ValidateKey(key string) bool {
	switch key {
	case metricSelectorKey, entitySelectorKey, resolutionKey, mzSelectorKey:
		return true
	default:
		return false
	}
}

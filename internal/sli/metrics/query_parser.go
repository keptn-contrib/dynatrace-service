package metrics

import (
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/parser"
)

// queryParser will parse an un-encoded metric query string (usually found in sli.yaml files) into a Query
type queryParser struct {
	query string
}

// newQueryParser creates a new queryParser for the specified query string.
func newQueryParser(query string) *queryParser {
	return &queryParser{
		query: strings.TrimSpace(query),
	}
}

// parse parses an un-encoded metrics query string (usually found in sli.yaml files) into a Query or returns an error.
// It only supports the current Metrics API V2 format (without a '?' prefix)
func (p *queryParser) parse() (*Query, error) {
	queryParameters, err := parser.NewQueryParser(p.query, p).Parse()
	if err != nil {
		return nil, err
	}
	return NewQuery(queryParameters.Get(metricSelectorKey), queryParameters.Get(entitySelectorKey))
}

// ValidateKey returns true if the specified key is part of a metrics query.
func (p *queryParser) ValidateKey(key string) bool {
	switch key {
	case metricSelectorKey, entitySelectorKey:
		return true
	default:
		return false
	}
}

package problemsv2

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

const (
	ProblemsV2Prefix = "PV2;"
)

const (
	problemSelectorKey = "problemSelector"
	entitySelectorKey  = "entitySelector"
)

// QueryParser will parse a v1 Problems v2 query string (usually found in sli.yaml files) into a Query
type QueryParser struct {
	query string
}

// NewQueryParser creates a new QueryParser for the specified Problems v2 query string.
func NewQueryParser(query string) *QueryParser {
	return &QueryParser{
		query: strings.TrimSpace(query),
	}
}

// Parse parses the query string into a Query or returns an error.
func (p *QueryParser) Parse() (*problems.Query, error) {
	if !strings.HasPrefix(p.query, ProblemsV2Prefix) {
		return nil, fmt.Errorf("Problems V2 queries should start with %s", ProblemsV2Prefix)
	}

	pieces, err := common.NewSLIPrefixParser(p.query, 2).Parse()
	if err != nil {
		return nil, err
	}

	keyValuePairs, err := common.NewSLIParser(pieces.Get(1), &problemsQueryKeyValidator{}).Parse()
	if err != nil {
		return nil, err
	}

	query := problems.NewQuery(keyValuePairs.GetValue(problemSelectorKey), keyValuePairs.GetValue(entitySelectorKey))
	return &query, nil
}

type problemsQueryKeyValidator struct{}

// ValidateKey returns true if the specified key is part of a Problems v2 query.
func (p *problemsQueryKeyValidator) ValidateKey(key string) bool {
	switch key {
	case problemSelectorKey, entitySelectorKey:
		return true
	default:
		return false
	}
}

package secpv2

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/secpv2"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

const (
	SecurityProblemsV2Prefix = "SECPV2;"
)

const (
	securityProblemSelectorKey = "securityProblemSelector"
)

// QueryParser will parse a v1 Security Problems v2 query string (usually found in sli.yaml files) into a Query
type QueryParser struct {
	query string
}

// NewQueryParser creates a new QueryParser for the specified Security Problems v2 query string.
func NewQueryParser(query string) *QueryParser {
	return &QueryParser{
		query: strings.TrimSpace(query),
	}
}

// Parse parses the query string into a Query or returns an error.
func (p *QueryParser) Parse() (*secpv2.Query, error) {
	if !strings.HasPrefix(p.query, SecurityProblemsV2Prefix) {
		return nil, fmt.Errorf("Security Problems V2 queries should start with %s", SecurityProblemsV2Prefix)
	}

	pieces, err := common.NewSLIPrefixParser(p.query, 2).Parse()
	if err != nil {
		return nil, err
	}

	keyValuePairs, err := common.NewSLIParser(pieces.Get(1), &securityProblemsQueryKeyValidator{}).Parse()
	if err != nil {
		return nil, err
	}

	query := secpv2.NewQuery(keyValuePairs.GetValue(securityProblemSelectorKey))
	return &query, nil
}

type securityProblemsQueryKeyValidator struct{}

// ValidateKey returns true if the specified key is part of a Security Problems v2 query.
func (v *securityProblemsQueryKeyValidator) ValidateKey(key string) bool {
	switch key {
	case securityProblemSelectorKey:
		return true
	default:
		return false
	}
}
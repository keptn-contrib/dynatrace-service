package slo

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

// SLOPrefix is the prefix of SLO queries.
const SLOPrefix = "SLO"

// QueryParser will parse a v1 SLO query string (usually found in sli.yaml files) into a Query
type QueryParser struct {
	query string
}

// NewQueryParser creates a new QueryParser for the specified SLO query string.
func NewQueryParser(query string) *QueryParser {
	return &QueryParser{
		query: strings.TrimSpace(query),
	}
}

// Parse parses the SLO string into a Query or returns an error.
func (p *QueryParser) Parse() (*Query, error) {
	pieces, err := common.NewSLIPrefixParser(p.query, 2).Parse()
	if err != nil {
		return nil, err
	}

	prefix, err := pieces.Get(0)
	if err != nil {
		return nil, err
	}
	if prefix != SLOPrefix {
		return nil, fmt.Errorf("SLO queries should start with %s", SLOPrefix)
	}

	sloID, err := pieces.Get(1)
	if err != nil {
		return nil, err
	}

	return NewQuery(sloID)
}

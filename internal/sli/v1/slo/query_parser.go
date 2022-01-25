package slo

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

const (
	SLOPrefix = "SLO;"
)

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
	if !strings.HasPrefix(p.query, SLOPrefix) {
		return nil, fmt.Errorf("SLO queries should start with %s", SLOPrefix)
	}

	pieces, err := common.NewSLIPrefixParser(p.query, 2).Parse()
	if err != nil {
		return nil, err
	}

	return NewQuery(pieces.Get(1))
}

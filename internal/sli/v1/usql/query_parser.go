package usql

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
)

// USQLPrefix is the prefix of USQL queries.
const USQLPrefix = "USQL"

// QueryParser will parse a v1 USQL query string (usually found in sli.yaml files) into a Query
type QueryParser struct {
	query string
}

// NewQueryParser creates a new QueryParser for the specified USQL query string.
func NewQueryParser(query string) *QueryParser {
	return &QueryParser{
		query: strings.TrimSpace(query),
	}
}

// Parse parses the query string into a Query or returns an error.
func (p *QueryParser) Parse() (*Query, error) {
	pieces, err := common.NewSLIPrefixParser(p.query, 4).Parse()
	if err != nil {
		return nil, err
	}

	prefix, err := pieces.Get(0)
	if err != nil {
		return nil, err
	}
	if prefix != USQLPrefix {
		return nil, fmt.Errorf("USQL queries should start with %s", USQLPrefix)
	}

	resultType, err := pieces.Get(1)
	if err != nil {
		return nil, err
	}

	dimension, err := pieces.Get(2)
	if err != nil {
		return nil, err
	}

	usqlQueryString, err := pieces.Get(3)
	if err != nil {
		return nil, err
	}

	innerQuery, err := usql.NewQuery(usqlQueryString)
	if err != nil {
		return nil, err
	}

	return NewQuery(resultType, dimension, *innerQuery)
}

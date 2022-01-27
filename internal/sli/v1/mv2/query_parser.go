package mv2

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
	v1metrics "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/metrics"
)

const (
	MV2Prefix = "MV2;"
)

// QueryParser will parse a MV2 query string (usually found in sli.yaml files) into a Query
type QueryParser struct {
	query string
}

// NewQueryParser creates a new QueryParser for the specified MV2 query string.
func NewQueryParser(query string) QueryParser {
	return QueryParser{
		query: strings.TrimSpace(query),
	}
}

// Parse parses the query string into a Query or returns an error.
func (p QueryParser) Parse() (*Query, error) {
	if !strings.HasPrefix(p.query, MV2Prefix) {
		return nil, fmt.Errorf("MV2 queries should start with %s", MV2Prefix)
	}

	pieces, err := common.NewSLIPrefixParser(p.query, 3).Parse()
	if err != nil {
		return nil, err
	}

	unit := pieces.Get(1)

	query, err := v1metrics.NewQueryParser(pieces.Get(2)).Parse()
	if err != nil {
		return nil, err
	}

	return NewQuery(unit, *query)
}

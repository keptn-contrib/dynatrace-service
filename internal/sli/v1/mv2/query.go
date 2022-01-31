package mv2

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
)

var unitPattern = regexp.MustCompile(`^(([bB]yte)|([Mm]icro[sS]econd))$`)

// Query encapsulates a MV2 query-
type Query struct {
	unit  string
	query metrics.Query
}

// NewQuery creates a Query from the specified unit and metrics query or returns an error.
func NewQuery(unit string, query metrics.Query) (*Query, error) {
	if unit == "" {
		return nil, errors.New("unit should not be empty")
	}

	if !unitPattern.MatchString(unit) {
		return nil, fmt.Errorf("invalid unit: %s", unit)
	}

	return &Query{
		unit:  unit,
		query: query,
	}, nil
}

// GetUnit gets the unit.
func (q *Query) GetUnit() string {
	return q.unit
}

// GetQuery gets the query.
func (q *Query) GetQuery() metrics.Query {
	return q.query
}

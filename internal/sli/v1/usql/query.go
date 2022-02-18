package usql

import (
	"errors"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
)

const (
	// SingleValueResultType is the result type for queries based on a single value.
	SingleValueResultType = "SINGLE_VALUE"

	// TableResultType is the result type for queries based on a table.
	TableResultType = "TABLE"

	// ColumnChartResultType is the result type for queries based on a column chart.
	ColumnChartResultType = "COLUMN_CHART"

	// LineChartResultType is the result type for queries based on a line chart.
	LineChartResultType = "LINE_CHART"

	// PieChartResultType is the result type for queries based on a pie chart.
	PieChartResultType = "PIE_CHART"
)

// Query represents a v1 USQL query.
type Query struct {
	resultType string
	dimension  string
	query      usql.Query
}

// NewQuery creates a Query from the specified result type, dimension and USQL query or returns an error.
func NewQuery(resultType string, dimension string, query usql.Query) (*Query, error) {
	if resultType == "" {
		return nil, errors.New("result type should not be empty")
	}
	if !isValidResultType(resultType) {
		return nil, fmt.Errorf("unknown result type: %s", resultType)
	}

	if (resultType == SingleValueResultType) && (dimension != "") {
		return nil, errors.New("dimension should be empty")
	} else if (resultType != SingleValueResultType) && (dimension == "") {
		return nil, errors.New("dimension should not be empty")
	}

	return &Query{
		resultType: resultType,
		dimension:  dimension,
		query:      query,
	}, nil
}

// GetResultType returns the result type.
func (u *Query) GetResultType() string {
	return u.resultType
}

// GetDimension returns the dimension.
func (u *Query) GetDimension() string {
	return u.dimension
}

// GetQuery returns the USQL query.
func (u *Query) GetQuery() usql.Query {
	return u.query
}

func isValidResultType(resultType string) bool {
	switch resultType {
	case SingleValueResultType, TableResultType, ColumnChartResultType, PieChartResultType:
		return true
	}
	return false
}

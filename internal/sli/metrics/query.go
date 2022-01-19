package metrics

import "errors"

const (
	metricSelectorKey = "metricSelector"
	entitySelectorKey = "entitySelector"
)

// Query encapsulates a metrics query.
type Query struct {
	metricSelector string
	entitySelector string
}

// NewQuery creates a new Query based on the provided metric and entity selector or returns an error.
func NewQuery(metricSelector string, entitySelector string) (*Query, error) {
	if metricSelector == "" {
		return nil, errors.New("metrics query must include a metric selector")
	}
	return &Query{
		metricSelector: metricSelector,
		entitySelector: entitySelector,
	}, nil
}

// ParseQuery parses a query string and returns a Query or an error.
// It only supports the current Metrics API V2 format (without a '?' prefix)
func ParseQuery(queryString string) (*Query, error) {
	return newQueryParser(queryString).parse()
}

// GetMetricSelector returns the metric selector.
func (m *Query) GetMetricSelector() string {
	return m.metricSelector
}

// GetEntitySelector returns the entity selector.
func (m *Query) GetEntitySelector() string {
	return m.entitySelector
}

// Build builds the query back into a string or returns an error.
func (m *Query) Build() (string, error) {
	return newQueryBuilder(m).build()
}

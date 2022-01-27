package metrics

import "errors"

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

// GetMetricSelector returns the metric selector.
func (m Query) GetMetricSelector() string {
	return m.metricSelector
}

// GetEntitySelector returns the entity selector.
func (m Query) GetEntitySelector() string {
	return m.entitySelector
}

package metrics

import "errors"

const ResolutionInf = "Inf"

// Query encapsulates a metrics query.
type Query struct {
	metricSelector string
	entitySelector string
	resolution     string
	mzSelector     string
}

// NewQuery creates a new Query based on the provided metric and entity selector with infinite resolution and no management or returns an error.
func NewQuery(metricSelector string, entitySelector string) (*Query, error) {
	return NewQueryWithResolutionAndMZSelector(metricSelector, entitySelector, ResolutionInf, "")
}

// NewQueryWithResolutionAndMZSelector creates a new Query based on the provided metric and entity selector, resolution and management zone selector or returns an error.
func NewQueryWithResolutionAndMZSelector(metricSelector string, entitySelector string, resolution string, mzSelector string) (*Query, error) {
	if metricSelector == "" {
		return nil, errors.New("metrics query must include a metric selector")
	}
	return &Query{
		metricSelector: metricSelector,
		entitySelector: entitySelector,
		resolution:     resolution,
		mzSelector:     mzSelector,
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

// GetResolution returns the resolution.
func (m Query) GetResolution() string {
	return m.resolution
}

// GetMZSelector returns the management zone selector.
func (m Query) GetMZSelector() string {
	return m.mzSelector
}

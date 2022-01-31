package usql

import "errors"

// Query encapsulates a USQL query.
type Query struct {
	query string
}

// NewQuery creates a new Query based on the provided USQL query or returns an error.
func NewQuery(query string) (*Query, error) {
	if query == "" {
		return nil, errors.New("USQL query should not be empty")
	}
	return &Query{
		query: query,
	}, nil
}

// GetQuery returns the USQL query.
func (m Query) GetQuery() string {
	return m.query
}

package slo

import "errors"

// Query encapsulates an SLO query-
type Query struct {
	sloID string
}

// NewQuery creates a Query from the specified SLO ID or returns an error.
func NewQuery(sloID string) (*Query, error) {
	if sloID == "" {
		return nil, errors.New("SLO ID should not be empty")
	}

	return &Query{
		sloID: sloID,
	}, nil
}

// GetSLOID gets the SLO ID.
func (q *Query) GetSLOID() string {
	return q.sloID
}

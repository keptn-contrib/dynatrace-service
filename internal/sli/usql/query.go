package usql

// Query encapsulates a USQL query.
type Query struct {
	query string
}

// NewQuery creates a new Query based on the provided USQL query.
func NewQuery(query string) Query {
	return Query{
		query: query,
	}
}

// GetQuery returns the USQL query.
func (m *Query) GetQuery() string {
	return m.query
}

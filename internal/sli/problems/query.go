package problems

// Query encapsulates a Problems v2 query.
type Query struct {
	problemSelector string
	entitySelector  string
}

// NewQuery creates a new Query based on the provided problem and entity selector.
func NewQuery(problemSelector string, entitySelector string) Query {
	return Query{
		problemSelector: problemSelector,
		entitySelector:  entitySelector,
	}
}

// GetProblemSelector returns the problem selector.
func (m *Query) GetProblemSelector() string {
	return m.problemSelector
}

// GetEntitySelector returns the entity selector.
func (m *Query) GetEntitySelector() string {
	return m.entitySelector
}

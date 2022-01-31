package secpv2

// Query encapsulates a Security Problems v2 query.
type Query struct {
	securityProblemSelector string
}

// NewQuery creates a new Query based on the provided security problem selector.
func NewQuery(securityProblemSelector string) Query {
	return Query{
		securityProblemSelector: securityProblemSelector,
	}
}

// GetSecurityProblemSelector returns the security problem selector.
func (m *Query) GetSecurityProblemSelector() string {
	return m.securityProblemSelector
}

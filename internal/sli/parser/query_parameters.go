package parser

// QueryParameters encapsulates a map of parameters.
type QueryParameters struct {
	parameters map[string]string
}

// NewQueryParameters creates a QueryParameters from the specified map of parameters.
func NewQueryParameters(parameters map[string]string) *QueryParameters {
	return &QueryParameters{
		parameters: parameters,
	}
}

// Get gets the value of a specified key or the zero value if it doesn't exist.
func (q *QueryParameters) Get(key string) string {
	return q.parameters[key]
}

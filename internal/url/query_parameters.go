package url

import "net/url"

type QueryParameters struct {
	values url.Values
}

// NewQueryParameters creates a new QueryParameters
func NewQueryParameters() *QueryParameters {
	return &QueryParameters{
		values: make(url.Values),
	}
}

// Add adds the value to the key
func (q *QueryParameters) Add(key string, value string) *QueryParameters {
	q.values.Add(key, value)
	return q
}

// Encode URL encodes the values
func (q *QueryParameters) Encode() string {
	return q.values.Encode()
}

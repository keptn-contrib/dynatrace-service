package dynatrace

import "net/url"

type queryParameters struct {
	values url.Values
}

// newQueryParameters creates a new QueryParameters
func newQueryParameters() *queryParameters {
	return &queryParameters{
		values: make(url.Values),
	}
}

// add adds the value to the key
func (q *queryParameters) add(key string, value string) {
	q.values.Add(key, value)
}

// encode URL encodes the values
func (q *queryParameters) encode() string {
	return q.values.Encode()
}

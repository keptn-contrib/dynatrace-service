package metrics

import (
	"fmt"
	"net/url"
	"strings"
)

// QueryParameters store not URL encoded key/value pairs
type QueryParameters struct {
	values         map[string]string
	iterationOrder []string
}

func NewQueryParameters() *QueryParameters {
	return &QueryParameters{
		values:         map[string]string{},
		iterationOrder: []string{},
	}
}

// Add will add a key/value pair to the QueryParameters if there was no entry for the key
// It will return an error if the key already exists
func (p *QueryParameters) Add(key string, value string) error {
	v, exists := p.values[key]
	if exists {
		return fmt.Errorf("duplicate query parameter '%s'", v)
	}

	p.values[key] = value
	p.iterationOrder = append(p.iterationOrder, key)
	return nil
}

func (p *QueryParameters) Get(key string) (string, bool) {
	value, exists := p.values[key]
	return value, exists
}

// ForEach will iterate the QueryParameters in insertion order
func (p *QueryParameters) ForEach(consumerFunc func(key string, value string)) {
	for _, key := range p.iterationOrder {
		consumerFunc(key, p.values[key])
	}
}

// Encode will encode the query parameters and return a correctly encoded URL query string
func (p *QueryParameters) Encode() string {
	queryString := ""
	for _, key := range p.iterationOrder {
		if queryString != "" {
			queryString += "&"
		}
		queryString += key + "=" + url.QueryEscape(p.values[key])
	}

	return queryString
}

const (
	fromKey           = "from"
	toKey             = "to"
	metricSelectorKey = "metricSelector"
	resolutionKey     = "resolution"
	entitySelectorKey = "entitySelector"
	delimiter         = "&"
	keyValueDelimiter = "="
)

// QueryParsing will parse a un-encoded metric definition query string (usually found in sli.yaml files) into QueryParameters
type QueryParsing struct {
	query string
}

func NewQueryParsing(query string) *QueryParsing {
	return &QueryParsing{
		query: strings.TrimSpace(query),
	}
}

// Parse will try parse a un-encoded metric definition query string (usually found in sli.yaml files) into QueryParameters
// or return an error in case it could not successfully do that.
// It will only support the current Metrics API V2 format (without a '?' prefix)
func (p *QueryParsing) Parse() (*QueryParameters, error) {
	if p.query == "" {
		return nil, fmt.Errorf("empty metric definition")
	}

	chunks := strings.Split(p.query, delimiter)
	// if we have more than 5 chunks, then we at least know that either there are duplicate keys, or there are some extra keys that we do not support
	if len(chunks) > 5 {
		return nil, fmt.Errorf("could not parse metric definition: %s", p.query)
	}

	queryParameters := NewQueryParameters()
	for _, chunk := range chunks {
		key, value, err := splitKeyValuePair(chunk)
		if err != nil {
			return nil, err
		}

		err = queryParameters.Add(key, value)
		if err != nil {
			return nil, err
		}
	}

	return queryParameters, nil
}

// splitKeyValuePair returns the split key-value pair or an error.
// we do not allow empty values like 'key=' or just 'key'
func splitKeyValuePair(keyValue string) (string, string, error) {
	keyValue = strings.TrimSpace(keyValue)
	if keyValue == "" {
		return "", "", fmt.Errorf("empty 'key=value' pair")
	}

	chunks := strings.Split(keyValue, keyValueDelimiter)
	if len(chunks) != 2 || chunks[0] == "" || chunks[1] == "" {
		return "", "", fmt.Errorf("could not parse 'key=value' pair correctly: %s", keyValue)
	}

	switch chunks[0] {
	case metricSelectorKey, entitySelectorKey, resolutionKey, fromKey, toKey:
		return chunks[0], chunks[1], nil
	default:
		return "", "", fmt.Errorf("unknown key in 'key=value' pair: %s", keyValue)
	}
}

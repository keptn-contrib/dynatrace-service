package metrics

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// QueryParameters store not URL encoded key/value pairs
type QueryParameters struct {
	values map[string]string
}

func NewQueryParameters() *QueryParameters {
	return &QueryParameters{
		values: make(map[string]string, 10),
	}
}

// Add will add a key/value pair to the QueryParameters if there was no entry for the key
// It will return an error if the key already exists
func (p *QueryParameters) Add(key string, value string) error {
	_, exists := p.values[key]
	if exists {
		return fmt.Errorf("duplicate key '%s'", key)
	}

	p.values[key] = value
	return nil
}

func (p *QueryParameters) Get(key string) string {
	return p.values[key]
}

func (p *QueryParameters) get(key string) (string, bool) {
	value, exists := p.values[key]
	return value, exists
}

func (p *QueryParameters) GetMetricSelector() string {
	value, _ := p.get(metricSelectorKey)
	return value
}

// ForEach will iterate the QueryParameters in insertion order
func (p *QueryParameters) ForEach(consumerFunc func(key string, value string)) {
	for _, key := range p.getSortedKeys() {
		consumerFunc(key, p.values[key])
	}
}

func (p *QueryParameters) getSortedKeys() []string {
	keys := make([]string, 0, len(p.values))
	for key := range p.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Encode will encode the query parameters and return a correctly encoded URL query string
func (p *QueryParameters) Encode() string {
	var buffer strings.Builder
	for _, key := range p.getSortedKeys() {
		if buffer.Len() > 0 {
			buffer.WriteByte('&')
		}
		buffer.WriteString(key)
		buffer.WriteByte('=')
		buffer.WriteString(url.QueryEscape(p.values[key]))
	}

	return buffer.String()
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

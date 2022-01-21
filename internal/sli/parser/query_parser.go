package parser

import (
	"fmt"
	"strings"
)

const (
	delimiter         = "&"
	keyValueDelimiter = "="
)

// KeyValidator interface validates keys within the query.
type KeyValidator interface {
	// ValidateKey returns true iff the specified key is allowed within the query.
	ValidateKey(key string) bool
}

// QueryParser parses an un-encoded query string (usually found in sli.yaml files) into QueryParameters.
type QueryParser struct {
	query     string
	validator KeyValidator
}

// NewQueryParser creates a QueryParser for the specified query and validator.
func NewQueryParser(query string, validator KeyValidator) *QueryParser {
	return &QueryParser{
		query:     strings.TrimSpace(query),
		validator: validator,
	}
}

// Parse parses an un-encoded query string (usually found in sli.yaml files) into QueryParameters or returns an error.
func (p *QueryParser) Parse() (*KeyValuePairs, error) {
	if p.query == "" {
		return nil, fmt.Errorf("query should not be empty")
	}

	if p.validator == nil {
		return nil, fmt.Errorf("key validator should not be nil")
	}

	chunks := strings.Split(p.query, delimiter)

	parameters := make(map[string]string)
	for _, chunk := range chunks {
		key, value, err := splitKeyValuePair(chunk, p.validator)
		if err != nil {
			return nil, err
		}

		err = add(parameters, key, value)
		if err != nil {
			return nil, err
		}
	}

	return NewKeyValuePairs(parameters), nil
}

// add adds the specified key and value to the map or returns an error if the map already contains the key.
func add(parameters map[string]string, key string, value string) error {
	_, exists := parameters[key]
	if exists {
		return fmt.Errorf("duplicate key '%s'", key)
	}

	parameters[key] = value
	return nil
}

// splitKeyValuePair returns the split key-value pair or an error.
// The pair must have both a non-empty key and value, i.e. 'key=' or just 'key' are not allowed.
func splitKeyValuePair(keyValue string, validator KeyValidator) (string, string, error) {
	keyValue = strings.TrimSpace(keyValue)
	if keyValue == "" {
		return "", "", fmt.Errorf("empty 'key=value' pair")
	}

	chunks := strings.Split(keyValue, keyValueDelimiter)
	if len(chunks) != 2 || chunks[0] == "" || chunks[1] == "" {
		return "", "", fmt.Errorf("could not parse 'key=value' pair correctly: %s", keyValue)
	}

	if !validator.ValidateKey(chunks[0]) {
		return "", "", fmt.Errorf("unknown key: %s", chunks[0])
	}

	return chunks[0], chunks[1], nil
}

package common

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

// SLIParser parses an un-encoded query string (usually found in sli.yaml files) into key-value pairs.
type SLIParser struct {
	query     string
	validator KeyValidator
}

// NewSLIParser creates a SLIParser for the specified query and validator.
func NewSLIParser(query string, validator KeyValidator) *SLIParser {
	return &SLIParser{
		query:     strings.TrimSpace(query),
		validator: validator,
	}
}

// Parse parses an un-encoded query string (usually found in sli.yaml files) into KeyValuePairs or returns an error.
func (p *SLIParser) Parse() (*KeyValuePairs, error) {
	if p.validator == nil {
		return nil, fmt.Errorf("key validator should not be nil")
	}

	chunks := strings.Split(p.query, delimiter)

	keyValues := make(map[string]string)
	for _, chunk := range chunks {
		if chunk == "" {
			continue
		}
		key, value, err := splitKeyValuePair(chunk)
		if err != nil {
			return nil, err
		}

		if !p.validator.ValidateKey(key) {
			return nil, fmt.Errorf("unknown key: %s", key)
		}

		err = add(keyValues, key, value)
		if err != nil {
			return nil, err
		}
	}

	kvp := NewKeyValuePairs(keyValues)
	return &kvp, nil
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
func splitKeyValuePair(keyValue string) (string, string, error) {
	keyValue = strings.TrimSpace(keyValue)
	if keyValue == "" {
		return "", "", fmt.Errorf("empty 'key=value' pair")
	}

	chunks := strings.SplitN(keyValue, keyValueDelimiter, 2)
	if len(chunks) != 2 || chunks[0] == "" || chunks[1] == "" {
		return "", "", fmt.Errorf("could not parse 'key=value' pair correctly: %s", keyValue)
	}

	return chunks[0], chunks[1], nil
}

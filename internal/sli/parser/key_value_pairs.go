package parser

// KeyValuePairs encapsulates a map of keys to values.
type KeyValuePairs struct {
	keyValues map[string]string
}

// NewKeyValuePairs creates a KeyValuePairs instance from the specified map of keys to values.
func NewKeyValuePairs(keyValues map[string]string) *KeyValuePairs {
	return &KeyValuePairs{
		keyValues: keyValues,
	}
}

// GetValue gets the value of a specified key or the zero value if it doesn't exist.
func (q *KeyValuePairs) GetValue(key string) string {
	return q.keyValues[key]
}

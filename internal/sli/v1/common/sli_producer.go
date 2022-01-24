package common

import (
	"sort"
	"strings"
)

// SLIProducer build a SLI string from KeyValuePairs.
type SLIProducer struct {
	keyValues *KeyValuePairs
}

// NewSLIProducer creates a new SLIProducer based on the specified KeyValuePairs and key orderer.
func NewSLIProducer(keyValues *KeyValuePairs) *SLIProducer {
	return &SLIProducer{
		keyValues: keyValues,
	}
}

// Produce produces a SLI string based on the KeyValuePairs ordered by key.
func (b *SLIProducer) Produce() string {
	sortedPairs := make([]string, 0, len(b.keyValues.keyValues))
	for _, key := range b.getSortedKeys() {
		sortedPairs = append(sortedPairs, makeKeyValuePair(key, b.keyValues.keyValues[key]))
	}
	return strings.Join(sortedPairs, delimiter)
}

func (b *SLIProducer) getSortedKeys() []string {
	keys := make([]string, 0, len(b.keyValues.keyValues))
	for key, _ := range b.keyValues.keyValues {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func makeKeyValuePair(key string, value string) string {
	keyValueBuilder := strings.Builder{}
	keyValueBuilder.WriteString(key)
	keyValueBuilder.WriteString(keyValueDelimiter)
	keyValueBuilder.WriteString(value)
	return keyValueBuilder.String()
}

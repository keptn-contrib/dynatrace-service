package common

import (
	"strings"
)

// SLIProducer build a SLI string from KeyValuePairs.
type SLIProducer struct {
	pairs *KeyValuePairs
}

// NewSLIProducer creates a new SLIProducer based on the specified KeyValuePairs and key orderer.
func NewSLIProducer(keyValues *KeyValuePairs) *SLIProducer {
	return &SLIProducer{
		pairs: keyValues,
	}
}

// Produce produces a SLI string based on the KeyValuePairs ordered by key.
func (b *SLIProducer) Produce() string {
	sortedPairs := make([]string, 0, b.pairs.count())
	for _, key := range b.pairs.getSortedKeys() {
		sortedPairs = append(sortedPairs, makeKeyValuePair(key, b.pairs.GetValue(key)))
	}
	return strings.Join(sortedPairs, delimiter)
}

func makeKeyValuePair(key string, value string) string {
	keyValueBuilder := strings.Builder{}
	keyValueBuilder.WriteString(key)
	keyValueBuilder.WriteString(keyValueDelimiter)
	keyValueBuilder.WriteString(value)
	return keyValueBuilder.String()
}

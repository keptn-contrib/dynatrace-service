package common

import (
	"fmt"
	"sort"
	"strings"
)

// KeyOrderer orders the keys within a query.
type KeyOrderer interface {
	// GetKeyPosition returns the position of a key in a query string and true, or false if it should not be present.
	GetKeyPosition(key string) (int, bool)
}

// SLIProducer build a SLI string from KeyValuePairs.
type SLIProducer struct {
	keyValues  *KeyValuePairs
	keyOrderer KeyOrderer
}

// NewSLIProducer creates a new SLIProducer based on the specified KeyValuePairs and key orderer.
func NewSLIProducer(keyValues *KeyValuePairs, keyOrderer KeyOrderer) *SLIProducer {
	return &SLIProducer{
		keyValues:  keyValues,
		keyOrderer: keyOrderer,
	}
}

// Produce produces a SLI string based on the KeyValuePairs or returns an error if the key is unexpected or cannot be ordered.
func (b *SLIProducer) Produce() (string, error) {
	if b.keyOrderer == nil {
		return "", fmt.Errorf("key orderer should not be nil")
	}

	pairs, err := b.makePairs()
	if err != nil {
		return "", err
	}

	return strings.Join(sortPairs(pairs), delimiter), nil
}

// makePairs combines the key-value pairs indexed by their order or returns an error.
func (b *SLIProducer) makePairs() (map[int]string, error) {
	pairs := make(map[int]string)
	for key, value := range b.keyValues.keyValues {

		order, shouldAppear := b.keyOrderer.GetKeyPosition(key)
		if !shouldAppear {
			return nil, fmt.Errorf("unexpected key: %s", key)
		}

		_, alreadyExists := pairs[order]
		if alreadyExists {
			return nil, fmt.Errorf("ambiguous ordering: %d", order)
		}

		keyValueBuilder := strings.Builder{}
		keyValueBuilder.WriteString(key)
		keyValueBuilder.WriteString(keyValueDelimiter)
		keyValueBuilder.WriteString(value)
		pairs[order] = keyValueBuilder.String()
	}
	return pairs, nil
}

// sortPairs sorts key-value pairs based on the index and thus returns them in sorted order.
func sortPairs(pairs map[int]string) []string {
	orderedKeys := make([]int, 0, len(pairs))
	for orderKey := range pairs {
		orderedKeys = append(orderedKeys, orderKey)
	}
	sort.Ints(orderedKeys)

	sortedPairs := make([]string, 0, len(pairs))
	for _, orderKey := range orderedKeys {
		sortedPairs = append(sortedPairs, pairs[orderKey])
	}
	return sortedPairs
}

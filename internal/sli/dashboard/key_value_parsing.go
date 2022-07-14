package dashboard

import (
	"fmt"
	"strings"
)

type keyValue struct {
	key   string
	value string
	split bool
}

type keyValueParsing struct {
	value string
}

func newKeyValueParsing(value string) *keyValueParsing {
	return &keyValueParsing{value: value}
}

func (p *keyValueParsing) parse() []keyValue {
	pairs := []keyValue{}

	pieces := strings.Split(p.value, ";")
	for _, pair := range pieces {
		pair = strings.TrimSpace(pair)
		key, value, found := strings.Cut(pair, "=")

		pairs = append(pairs, keyValue{
			key:   strings.TrimSpace(key),
			value: strings.TrimSpace(value),
			split: found,
		})
	}

	return pairs
}

type duplicateKeyError struct {
	key string
}

func (err *duplicateKeyError) Error() string {
	return fmt.Sprintf("duplicate key '%s' in SLO definition", err.key)
}

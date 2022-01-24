package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSLIProducer tests SLIProducer.
func TestSLIProducer(t *testing.T) {
	tests := []struct {
		name                 string
		inputQueryParameters *KeyValuePairs
		expectedOutput       string
	}{
		{
			name:                 "Empty query parameters",
			inputQueryParameters: NewKeyValuePairs(map[string]string{}),
		},
		{
			name:                 "One key",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key1": "value1"}),
			expectedOutput:       "key1=value1",
		},
		{
			name:                 "Two keys",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key1": "value1", "key2": "value2"}),
			expectedOutput:       "key1=value1&key2=value2",
		},
		{
			name:                 "Two keys - potentially different order",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key2": "value2", "key1": "value1"}),
			expectedOutput:       "key1=value1&key2=value2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := NewSLIProducer(tc.inputQueryParameters).Produce()
			assert.EqualValues(t, tc.expectedOutput, output)
		})
	}
}

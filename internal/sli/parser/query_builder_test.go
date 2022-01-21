package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryBuilder tests QueryBuilder.
func TestQueryBuilder(t *testing.T) {
	tests := []struct {
		name                 string
		inputQueryParameters *KeyValuePairs
		keyOrderer           KeyOrderer
		expectedOutput       string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:                 "Empty query parameters",
			inputQueryParameters: NewKeyValuePairs(map[string]string{}),
			keyOrderer:           &testKeyOrderer{},
		},
		{
			name:                 "One key",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key1": "value1"}),
			keyOrderer:           &testKeyOrderer{},
			expectedOutput:       "key1=value1",
		},
		{
			name:                 "Two keys",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key1": "value1", "key2": "value2"}),
			keyOrderer:           &testKeyOrderer{},
			expectedOutput:       "key1=value1&key2=value2",
		},
		{
			name:                 "Two keys - potentially different order",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key2": "value2", "key1": "value1"}),
			keyOrderer:           &testKeyOrderer{},
			expectedOutput:       "key1=value1&key2=value2",
		},

		// The following cases expect errors
		{
			name:                 "One unknown key",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key3": "value3"}),
			keyOrderer:           &testKeyOrderer{},
			expectError:          true,
			expectedErrorMessage: "unexpected key",
		},
		{
			name:                 "Two keys - ambiguous ordering",
			inputQueryParameters: NewKeyValuePairs(map[string]string{"key1": "value1", "key2": "value2"}),
			keyOrderer:           &testAmbiguousKeyOrderer{},
			expectError:          true,
			expectedErrorMessage: "ambiguous ordering",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := NewQueryBuilder(tc.inputQueryParameters, tc.keyOrderer).Build()
			if tc.expectError {
				assert.Error(t, err)
				assert.EqualValues(t, "", output)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.expectedOutput, output)
				assert.Empty(t, tc.expectedErrorMessage, "fix test setup")
			}
		})
	}
}

type testKeyOrderer struct{}

func (o *testKeyOrderer) GetKeyPosition(key string) (int, bool) {
	switch key {
	case "key1":
		return 0, true
	case "key2":
		return 1, true
	default:
		return 0, false
	}
}

type testAmbiguousKeyOrderer struct{}

func (o *testAmbiguousKeyOrderer) GetKeyPosition(key string) (int, bool) {
	switch key {
	case "key1":
		return 0, true
	case "key2":
		return 0, true
	default:
		return 0, false
	}
}

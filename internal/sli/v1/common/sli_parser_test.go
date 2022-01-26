package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSLIParser tests parsing of SLIs.
func TestSLIParser(t *testing.T) {
	tests := []struct {
		name                               string
		input                              string
		keyValidator                       KeyValidator
		expectedKeyValuePairsAssertionFunc func(t assert.TestingT, p *KeyValuePairs)
		expectError                        bool
		expectedErrorMessage               string
	}{
		{
			name:         "one key-value, expecting one",
			input:        "key1=value1",
			keyValidator: &validatorWithOneKey{},
			expectedKeyValuePairsAssertionFunc: func(t assert.TestingT, p *KeyValuePairs) {
				assert.EqualValues(t, "value1", p.GetValue("key1"))
			},
		},
		{
			name:         "two key-values, expecting two",
			input:        "key1=value1&key2=value2",
			keyValidator: &validatorWithTwoKeys{},
			expectedKeyValuePairsAssertionFunc: func(t assert.TestingT, p *KeyValuePairs) {
				assert.EqualValues(t, "value1", p.GetValue("key1"))
				assert.EqualValues(t, "value2", p.GetValue("key2"))
			},
		},

		// Expect error in these cases
		{
			name:         "empty input",
			keyValidator: &validatorWithOneKey{},
			expectedKeyValuePairsAssertionFunc: func(t assert.TestingT, p *KeyValuePairs) {
				assert.EqualValues(t, 0, p.count())
			},
		},
		{
			name:                 "just key",
			input:                "key1",
			keyValidator:         &validatorWithOneKey{},
			expectError:          true,
			expectedErrorMessage: "could not parse 'key=value' pair correctly",
		},
		{
			name:                 "empty value",
			input:                "key1=",
			keyValidator:         &validatorWithOneKey{},
			expectError:          true,
			expectedErrorMessage: "could not parse 'key=value' pair correctly",
		},
		{
			name:                 "empty key",
			input:                "=value1",
			keyValidator:         &validatorWithOneKey{},
			expectError:          true,
			expectedErrorMessage: "could not parse 'key=value' pair correctly",
		},
		{
			name:         "empty key-value pair",
			input:        "key1=value1&",
			keyValidator: &validatorWithOneKey{},
			expectedKeyValuePairsAssertionFunc: func(t assert.TestingT, p *KeyValuePairs) {
				assert.EqualValues(t, "value1", p.GetValue("key1"))
			},
		},
		{
			name:                 "unexpected key",
			input:                "key1=value1&key2=value2",
			keyValidator:         &validatorWithOneKey{},
			expectError:          true,
			expectedErrorMessage: "unknown key",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyValuePairs, err := NewSLIParser(tc.input, tc.keyValidator).Parse()
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, keyValuePairs)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, keyValuePairs) {
					tc.expectedKeyValuePairsAssertionFunc(t, keyValuePairs)
				}
				assert.Empty(t, tc.expectedErrorMessage, "fix test setup")
			}
		})
	}
}

type validatorWithOneKey struct{}

func (t *validatorWithOneKey) ValidateKey(key string) bool {
	switch key {
	case "key1":
		return true
	default:
		return false
	}
}

type validatorWithTwoKeys struct{}

func (t *validatorWithTwoKeys) ValidateKey(key string) bool {
	switch key {
	case "key1":
		return true
	case "key2":
		return true
	default:
		return false
	}
}

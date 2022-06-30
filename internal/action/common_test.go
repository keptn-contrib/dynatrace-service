package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomProperties_Add(t *testing.T) {
	testConfigs := []struct {
		name     string
		key      string
		value    string
		expected customProperties
	}{
		{
			name:     "both empty",
			key:      "",
			value:    "",
			expected: customProperties{"": ""},
		},
		{
			name:     "key empty",
			key:      "",
			value:    "value",
			expected: customProperties{"": "value"},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "",
			expected: customProperties{"key": ""},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "value",
			expected: customProperties{"key": "value"},
		},
	}
	for _, testCfg := range testConfigs {
		t.Run(testCfg.name, func(t *testing.T) {
			cp := customProperties{}
			cp.add(testCfg.key, testCfg.value)

			assert.Equal(t, testCfg.expected, cp)
		})
	}
}

func TestCustomProperties_AddIfNonEmpty(t *testing.T) {

	testConfigs := []struct {
		name     string
		key      string
		value    string
		expected customProperties
	}{
		{
			name:     "both empty",
			key:      "",
			value:    "",
			expected: customProperties{},
		},
		{
			name:     "key empty",
			key:      "",
			value:    "value",
			expected: customProperties{},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "",
			expected: customProperties{},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "value",
			expected: customProperties{"key": "value"},
		},
	}
	for _, testCfg := range testConfigs {
		t.Run(testCfg.name, func(t *testing.T) {
			cp := customProperties{}
			cp.addIfNonEmpty(testCfg.key, testCfg.value)

			assert.Equal(t, testCfg.expected, cp)
		})
	}
}

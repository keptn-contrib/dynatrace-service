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
		expected CustomProperties
	}{
		{
			name:     "both empty",
			key:      "",
			value:    "",
			expected: CustomProperties{"": ""},
		},
		{
			name:     "key empty",
			key:      "",
			value:    "value",
			expected: CustomProperties{"": "value"},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "",
			expected: CustomProperties{"key": ""},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "value",
			expected: CustomProperties{"key": "value"},
		},
	}
	for _, testCfg := range testConfigs {
		t.Run(testCfg.name, func(t *testing.T) {
			cp := CustomProperties{}
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
		expected CustomProperties
	}{
		{
			name:     "both empty",
			key:      "",
			value:    "",
			expected: CustomProperties{},
		},
		{
			name:     "key empty",
			key:      "",
			value:    "value",
			expected: CustomProperties{},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "",
			expected: CustomProperties{},
		},
		{
			name:     "value empty",
			key:      "key",
			value:    "value",
			expected: CustomProperties{"key": "value"},
		},
	}
	for _, testCfg := range testConfigs {
		t.Run(testCfg.name, func(t *testing.T) {
			cp := CustomProperties{}
			cp.addIfNonEmpty(testCfg.key, testCfg.value)

			assert.Equal(t, testCfg.expected, cp)
		})
	}
}

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryParseImageAndTag(t *testing.T) {

	testConfigs := []struct {
		name     string
		input    interface{}
		expected ImageAndTag
	}{
		{
			name:     "valid image and tag",
			input:    "my-image:my-tag",
			expected: NewImageAndTag("my-image", "my-tag"),
		},
		{
			name:     "valid image",
			input:    "my-image",
			expected: NewImageAndTag("my-image", NotAvailable),
		},
		{
			name:     "invalid input string - missing image",
			input:    ":my-tag",
			expected: NewNotAvailableImageAndTag(),
		},
		{
			name:     "invalid input string - missing tag",
			input:    "my-image:",
			expected: NewNotAvailableImageAndTag(),
		},
		{
			name:     "invalid input string - image and tag missing",
			input:    ":",
			expected: NewNotAvailableImageAndTag(),
		},
		{
			name:     "invalid input string - empty",
			input:    "",
			expected: NewNotAvailableImageAndTag(),
		},
		{
			name:     "valid input string - n/a ",
			input:    "n/a",
			expected: NewNotAvailableImageAndTag(),
		},
		{
			name:     "valid input - number",
			input:    1244,
			expected: NewNotAvailableImageAndTag(),
		},
		{
			name:     "valid input - bool",
			input:    false,
			expected: NewNotAvailableImageAndTag(),
		},
		{
			name: "valid input - struct",
			input: struct {
				image string
				tag   string
			}{
				image: "my-image",
				tag:   "my-tag",
			},
			expected: NewNotAvailableImageAndTag(),
		},
	}
	for _, testConfig := range testConfigs {
		t.Run(testConfig.name, func(t *testing.T) {
			assert.EqualValues(t, testConfig.expected, TryParseImageAndTag(testConfig.input))
		})
	}
}

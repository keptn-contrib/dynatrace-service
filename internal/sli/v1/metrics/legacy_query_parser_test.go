package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLegacyQueryParser(t *testing.T) {
	testConfigs := []struct {
		name                   string
		input                  string
		expectedMetricSelector string
		expectedEntitySelector string
		expectError            bool
		expectedErrorMessage   string
	}{
		{
			name:                   "old standard format transformed to new one",
			input:                  "builtin:service.requestCount.total:merge(0):sum?scope=tag(keptn_project:my-proj),tag(keptn_stage:dev),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedMetricSelector: "builtin:service.requestCount.total:merge(0):sum",
			expectedEntitySelector: "tag(keptn_project:my-proj),tag(keptn_stage:dev),tag(keptn_service:carts),tag(keptn_deployment:direct),type(SERVICE)",
		},
		{
			name:                   "old standard format without scope - no changes to input",
			input:                  "builtin:service.requestCount.total:merge(0):sum",
			expectedMetricSelector: "builtin:service.requestCount.total:merge(0):sum",
		},
		{
			name:                   "old standard format, missing scope key",
			input:                  "builtin:service.requestCount.total:merge(0):sum?",
			expectedMetricSelector: "builtin:service.requestCount.total:merge(0):sum",
		},
		{
			name:                 "old standard format, missing scope value",
			input:                "builtin:service.requestCount.total:merge(0):sum?scope",
			expectError:          true,
			expectedErrorMessage: "could not parse 'key=value' pair correctly",
		},
		{
			name:                 "old standard format, missing scope value",
			input:                "builtin:service.requestCount.total:merge(0):sum?scope=",
			expectError:          true,
			expectedErrorMessage: "could not parse 'key=value' pair correctly",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewLegacyQueryParser(tc.input).Parse()
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
				assert.Nil(t, query)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, query) {
					assert.Equal(t, tc.expectedMetricSelector, query.GetMetricSelector())
					assert.Equal(t, tc.expectedEntitySelector, query.GetEntitySelector())
				}
				assert.Empty(t, tc.expectedErrorMessage, "fix test setup")
			}
		})
	}
}

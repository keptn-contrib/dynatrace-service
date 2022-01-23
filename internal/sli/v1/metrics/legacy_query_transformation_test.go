package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformationFromOldFormatToNewFormatWorks(t *testing.T) {
	testConfigs := []struct {
		name           string
		input          string
		expectedResult string
		shouldFail     bool
		errMessage     string
	}{
		{
			name:           "old standard format transformed to new one",
			input:          "builtin:service.requestCount.total:merge(0):sum?scope=tag(keptn_project:my-proj),tag(keptn_stage:dev),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedResult: "metricSelector=builtin:service.requestCount.total:merge(0):sum&entitySelector=tag(keptn_project:my-proj),tag(keptn_stage:dev),tag(keptn_service:carts),tag(keptn_deployment:direct),type(SERVICE)",
		},
		{
			name:           "old standard format without scope - no changes to input",
			input:          "builtin:service.requestCount.total:merge(0):sum",
			expectedResult: "builtin:service.requestCount.total:merge(0):sum",
		},
		{
			name:       "old standard format, missing scope key",
			input:      "builtin:service.requestCount.total:merge(0):sum?",
			shouldFail: true,
			errMessage: "missing 'scope=<scope>'",
		},
		{
			name:       "old standard format, missing scope value",
			input:      "builtin:service.requestCount.total:merge(0):sum?scope",
			shouldFail: true,
			errMessage: "missing 'scope=<scope>'",
		},
		{
			name:       "old standard format, missing scope value",
			input:      "builtin:service.requestCount.total:merge(0):sum?scope=",
			shouldFail: true,
			errMessage: "missing value",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewLegacyQueryTransformation(tc.input).Transform()
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMessage)
				assert.Empty(t, actual)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, actual)
				assert.Empty(t, tc.errMessage, "fix test setup")
			}
		})
	}
}

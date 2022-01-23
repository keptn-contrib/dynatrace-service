package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuery(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name                   string
		metricSelector         string
		entitySelector         string
		expectedMetricSelector string
		expectedEntitySelector string
		expectError            bool
		expectedErrorMessage   string
	}{
		{
			name:                   "with metric and entity selector",
			metricSelector:         "builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg",
			entitySelector:         "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectedMetricSelector: "builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg",
			expectedEntitySelector: "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
		},
		{
			name:                   "with just metric selector",
			metricSelector:         "builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg",
			expectedMetricSelector: "builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg",
		},
		// Error cases below:
		{
			name:                 "with no metric or entity selector",
			entitySelector:       "",
			expectError:          true,
			expectedErrorMessage: "metrics query must include a metric selector",
		},
		{
			name:                 "with just entity selector",
			entitySelector:       "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectError:          true,
			expectedErrorMessage: "metrics query must include a metric selector",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewQuery(tc.metricSelector, tc.entitySelector)
			if tc.expectError {
				assert.Nil(t, query)
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage)
				}
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, query) {
					assert.EqualValues(t, tc.expectedMetricSelector, query.GetMetricSelector())
					assert.EqualValues(t, tc.expectedEntitySelector, query.GetEntitySelector())
				}
			}
		})
	}
}

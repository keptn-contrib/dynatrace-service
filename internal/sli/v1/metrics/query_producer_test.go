package metrics

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"

	"github.com/stretchr/testify/assert"
)

func TestQueryProducer_Produce(t *testing.T) {
	testConfigs := []struct {
		name                      string
		inputMetricQuery          metrics.Query
		expectedMetricQueryString string
		expectError               bool
		expectedErrorMessage      string
	}{
		{
			name:                      "valid with both metric and entity selectors",
			inputMetricQuery:          newQuery(t, "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)", "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)"),
			expectedMetricQueryString: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectError:               false,
		},
		{
			name:                      "valid with just metric selector",
			inputMetricQuery:          newQuery(t, "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)", ""),
			expectedMetricQueryString: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectError:               false,
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			metricQueryString, err := NewQueryProducer(&tc.inputMetricQuery).Produce()
			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, metricQueryString)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMetricQueryString, metricQueryString)
				assert.Empty(t, tc.expectedErrorMessage, "fix test setup")
			}
		})
	}
}

func newQuery(t *testing.T, meticSelector string, entitySelector string) metrics.Query {
	query, err := metrics.NewQuery(meticSelector, entitySelector)
	assert.NoError(t, err)
	assert.NotNil(t, query)
	return *query
}

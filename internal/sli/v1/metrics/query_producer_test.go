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
	}{
		{
			name:                      "valid with both metric and entity selectors",
			inputMetricQuery:          newQuery(t, "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)", "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)"),
			expectedMetricQueryString: "entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
		},
		{
			name:                      "valid with just metric selector",
			inputMetricQuery:          newQuery(t, "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)", ""),
			expectedMetricQueryString: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			metricQueryString := NewQueryProducer(tc.inputMetricQuery).Produce()
			assert.Equal(t, tc.expectedMetricQueryString, metricQueryString)
		})
	}
}

func newQuery(t *testing.T, meticSelector string, entitySelector string) metrics.Query {
	query, err := metrics.NewQuery(meticSelector, entitySelector)
	assert.NoError(t, err)
	assert.NotNil(t, query)
	return *query
}

package mv2

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"

	"github.com/stretchr/testify/assert"
)

func TestQueryProducer_Produce(t *testing.T) {
	testConfigs := []struct {
		name                   string
		inputMV2Query          Query
		expectedMV2QueryString string
	}{
		{
			name:                   "Byte works",
			inputMV2Query:          newQuery(t, "Byte", "builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names", "type(DOCKER_CONTAINER_GROUP_INSTANCE)"),
			expectedMV2QueryString: "MV2;Byte;entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)&metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names",
		},
		{
			name:                   "MicroSecond works",
			inputMV2Query:          newQuery(t, "MicroSecond", "builtin:service.response.client:merge(\"dt.entity.service\"):avg:names", "type(SERVICE)"),
			expectedMV2QueryString: "MV2;MicroSecond;entitySelector=type(SERVICE)&metricSelector=builtin:service.response.client:merge(\"dt.entity.service\"):avg:names",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			pv2QueryString := NewQueryProducer(tc.inputMV2Query).Produce()
			assert.Equal(t, tc.expectedMV2QueryString, pv2QueryString)
		})
	}
}

func newQuery(t *testing.T, unit string, meticSelector string, entitySelector string) Query {

	metricsQuery, err := metrics.NewQuery(meticSelector, entitySelector, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, metricsQuery)

	query, err := NewQuery(unit, *metricsQuery)
	assert.NoError(t, err)
	assert.NotNil(t, query)
	return *query
}

package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMV2Query(t *testing.T) {

	tests := []struct {
		name          string
		mv2Query      string
		expectedQuery string
		expectedUnit  string
		shouldFail    bool
	}{
		// these should fail
		{
			name:       "percent unit does not work",
			mv2Query:   "MV2;Percent;metricSelector=builtin:host.cpu.usage:merge(\"dt.entity.host\"):avg:names&entitySelector=type(HOST)",
			shouldFail: true,
		},
		{
			name:       "missing microsecond metric unit",
			mv2Query:   "MV2;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			shouldFail: true,
		},
		{
			name:       "missing mv2 prefix",
			mv2Query:   "MicroSecond;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			shouldFail: true,
		},
		{
			name:       "missing mv2 prefix",
			mv2Query:   "MV2;MicroSeconds;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			shouldFail: true,
		},
		// these should not fail
		{
			name:          "microsecond metric works",
			mv2Query:      "MV2;MicroSecond;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectedQuery: "metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works 2",
			mv2Query:      "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.response.server:filter(and(in(\"dt.entity.service_method\",entitySelector(\"type(service_method),entityName(~\"/api/ui/v2/bootstrap~\")\")))):splitBy(\"dt.entity.service_method\"):percentile(90)",
			expectedQuery: "metricSelector=builtin:service.keyRequest.response.server:filter(and(in(\"dt.entity.service_method\",entitySelector(\"type(service_method),entityName(~\"/api/ui/v2/bootstrap~\")\")))):splitBy(\"dt.entity.service_method\"):percentile(90)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works - metric selector first",
			mv2Query:      "MV2;MicroSecond;metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedQuery: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works - entity selector first - MicroSecond unit",
			mv2Query:      "MV2;MicroSecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works - entity selector first - Microsecond unit",
			mv2Query:      "MV2;Microsecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "Microsecond",
		},
		{
			name:          "microsecond metric works - entity selector first - microsecond unit",
			mv2Query:      "MV2;microsecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "microsecond",
		},
		{
			name:          "microsecond metric works - entity selector first - microSecond unit",
			mv2Query:      "MV2;microSecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "microSecond",
		},
		{
			name:          "byte metric works - Byte unit",
			mv2Query:      "MV2;Byte;metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedQuery: "metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedUnit:  "Byte",
		},
		{
			name:          "byte metric works - byte unit",
			mv2Query:      "MV2;byte;metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedQuery: "metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedUnit:  "byte",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			metricsQuery, metricsUnit, err := ParseMV2Query(tt.mv2Query)
			if tt.shouldFail {
				assert.Error(t, err)
				assert.Empty(t, metricsQuery)
				assert.Empty(t, metricsUnit)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.expectedQuery, metricsQuery)
				assert.EqualValues(t, tt.expectedUnit, metricsUnit)
			}
		})
	}
}

func TestConvertToMV2Query(t *testing.T) {

	tests := []struct {
		name             string
		metricsQuery     string
		metricUnit       string
		expectedMV2Query string
		shouldFail       bool
	}{
		// this should fail
		{
			name:         "Count does not work",
			metricsQuery: "metricSelector=builtin:host.dns.queryCount:merge(\"dnsServerIp\"):merge(\"dt.entity.host\"):avg:names&entitySelector=type(HOST)",
			metricUnit:   "Count",
			shouldFail:   true,
		},

		// these should work
		{
			name:             "Byte works",
			metricsQuery:     "metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names&entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)",
			metricUnit:       "Byte",
			expectedMV2Query: "MV2;Byte;metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names&entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)",
		},
		{
			name:             "MicroSecond works",
			metricsQuery:     "metricSelector=builtin:service.response.client:merge(\"dt.entity.service\"):avg:names&entitySelector=type(SERVICE)",
			metricUnit:       "MicroSecond",
			expectedMV2Query: "MV2;MicroSecond;metricSelector=builtin:service.response.client:merge(\"dt.entity.service\"):avg:names&entitySelector=type(SERVICE)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mv2Query, err := ConvertToMV2Query(tt.metricsQuery, tt.metricUnit)
			if tt.shouldFail {
				assert.Error(t, err)
				assert.Empty(t, mv2Query)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.expectedMV2Query, mv2Query)
			}
		})
	}
}

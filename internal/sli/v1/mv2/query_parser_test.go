package mv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryParser tests the QueryParser
func TestQueryParser(t *testing.T) {
	tests := []struct {
		name                   string
		inputMV2Query          string
		expectedMetricSelector string
		expectedEntitySelector string
		expectedUnit           string
		expectError            bool
	}{
		// these should fail
		{
			name:          "percent unit does not work",
			inputMV2Query: "MV2;Percent;metricSelector=builtin:host.cpu.usage:merge(\"dt.entity.host\"):avg:names&entitySelector=type(HOST)",
			expectError:   true,
		},
		{
			name:          "missing microsecond metric unit",
			inputMV2Query: "MV2;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectError:   true,
		},
		{
			name:          "missing mv2 prefix",
			inputMV2Query: "MicroSecond;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectError:   true,
		},
		{
			name:          "missing mv2 prefix",
			inputMV2Query: "MV2;MicroSeconds;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectError:   true,
		},
		// these should not fail
		{
			name:                   "microsecond metric works",
			inputMV2Query:          "MV2;MicroSecond;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectedMetricSelector: "builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectedUnit:           "MicroSecond",
		},
		{
			name:                   "microsecond metric works 2",
			inputMV2Query:          "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.response.server:filter(and(in(\"dt.entity.service_method\",entitySelector(\"type(service_method),entityName(~\"/api/ui/v2/bootstrap~\")\")))):splitBy(\"dt.entity.service_method\"):percentile(90)",
			expectedMetricSelector: "builtin:service.keyRequest.response.server:filter(and(in(\"dt.entity.service_method\",entitySelector(\"type(service_method),entityName(~\"/api/ui/v2/bootstrap~\")\")))):splitBy(\"dt.entity.service_method\"):percentile(90)",
			expectedUnit:           "MicroSecond",
		},
		{
			name:                   "microsecond metric works - metric selector first",
			inputMV2Query:          "MV2;MicroSecond;metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedUnit:           "MicroSecond",
		},
		{
			name:                   "microsecond metric works - entity selector first - MicroSecond unit",
			inputMV2Query:          "MV2;MicroSecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedUnit:           "MicroSecond",
		},
		{
			name:                   "microsecond metric works - entity selector first - Microsecond unit",
			inputMV2Query:          "MV2;Microsecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedUnit:           "Microsecond",
		},
		{
			name:                   "microsecond metric works - entity selector first - microsecond unit",
			inputMV2Query:          "MV2;microsecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedUnit:           "microsecond",
		},
		{
			name:                   "microsecond metric works - entity selector first - microSecond unit",
			inputMV2Query:          "MV2;microSecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedUnit:           "microSecond",
		},
		{
			name:                   "byte metric works - Byte unit",
			inputMV2Query:          "MV2;Byte;metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedMetricSelector: "builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedUnit:           "Byte",
		},
		{
			name:                   "byte metric works - byte unit",
			inputMV2Query:          "MV2;byte;metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedMetricSelector: "builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedUnit:           "byte",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewQueryParser(tc.inputMV2Query).Parse()
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, query)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, query) {
					assert.EqualValues(t, tc.expectedUnit, query.GetUnit())
					assert.EqualValues(t, tc.expectedMetricSelector, query.GetQuery().GetMetricSelector())
					assert.EqualValues(t, tc.expectedEntitySelector, query.GetQuery().GetEntitySelector())
				}
			}
		})
	}
}

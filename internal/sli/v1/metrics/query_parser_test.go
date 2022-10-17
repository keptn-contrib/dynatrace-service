package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryParser_Parse(t *testing.T) {
	testConfigs := []struct {
		name                   string
		input                  string
		expectedMetricSelector string
		expectedEntitySelector string
		expectedResolution     string
		expectedMZSelector     string
		expectError            bool
		expectedErrorMessage   string
	}{
		{
			name:                   "just metricSelector",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
		},
		{
			name:                   "metricSelector and resolution",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&resolution=30m",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedResolution:     "30m",
		},
		{
			name:                   "metricSelector, resolution and mzSelector",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&resolution=30m&mzSelector=mzId(123)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedResolution:     "30m",
			expectedMZSelector:     "mzId(123)",
		},
		{
			name:                   "metricSelector, entitySelector, resolution and mzSelector",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed)&resolution=30m&mzSelector=mzId(123)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedResolution:     "30m",
			expectedMZSelector:     "mzId(123)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_managed)",
		},

		{
			name:                   "standard service response time",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
		},
		{
			name:                   "standard total error rate",
			input:                  "metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectedMetricSelector: "builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg",
			expectedEntitySelector: "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
		},
		{
			name:                   "uri reserved character '/' and spaces in entity selector",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag2:tom / jerry)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag2:tom / jerry)",
		},
		{
			name:                   "uri reserved character '+' and spaces in metric selector",
			input:                  "metricSelector=(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Act ions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
			expectedMetricSelector: "(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Act ions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
		},
		{
			name:                   "uri reserved character '/' and spaces in entity selector",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag2:tom / jerry)",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedEntitySelector: "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag2:tom / jerry)",
		},
		{
			name:                   "uri reserved character '+' and spaces in metric selector",
			input:                  "metricSelector=(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
			expectedMetricSelector: "(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
		},
		// Error cases below:
		{
			// actually a tag 'my_tag:tom & jerry' would be totally fine from a Dynatrace API perspective
			name:                 "uri reserved character '&' in entity selector fails",
			input:                "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag:tom & jerry)",
			expectError:          true,
			expectedErrorMessage: "could not parse 'key=value' pair correctly",
		},
		{
			name:                 "standard service response time with additional unknown key fails",
			input:                "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&myKey=myValue",
			expectError:          true,
			expectedErrorMessage: "unknown key",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			metricsQuery, err := NewQueryParser(tc.input).Parse()
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, metricsQuery)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metricsQuery)
				assert.EqualValues(t, tc.expectedMetricSelector, metricsQuery.GetMetricSelector())
				assert.EqualValues(t, tc.expectedEntitySelector, metricsQuery.GetEntitySelector())
				assert.EqualValues(t, tc.expectedResolution, metricsQuery.GetResolution())
				assert.EqualValues(t, tc.expectedMZSelector, metricsQuery.GetMZSelector())
				assert.Empty(t, tc.expectedErrorMessage, "fix test setup")
			}
		})
	}
}

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

func TestQuery_Build(t *testing.T) {
	testConfigs := []struct {
		name                      string
		inputMetricQuery          Query
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
			metricQueryString, err := tc.inputMetricQuery.Build()
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

func newQuery(t *testing.T, meticSelector string, entitySelector string) Query {
	query, err := NewQuery(meticSelector, entitySelector)
	assert.NoError(t, err)
	assert.NotNil(t, query)
	return *query
}

func TestQuery_ParseQuery(t *testing.T) {
	testConfigs := []struct {
		name                   string
		input                  string
		expectedMetricSelector string
		expectedEntitySelector string
		expectError            bool
		expectedErrorMessage   string
	}{
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
			metricsQuery, err := ParseQuery(tc.input)
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, metricsQuery)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metricsQuery)
				assert.EqualValues(t, tc.expectedMetricSelector, metricsQuery.GetMetricSelector())
				assert.EqualValues(t, tc.expectedEntitySelector, metricsQuery.GetEntitySelector())
				assert.Empty(t, tc.expectedErrorMessage, "fix test setup")
			}
		})
	}
}

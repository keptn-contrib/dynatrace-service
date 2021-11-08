package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type keyValuePair struct {
	key   string
	value string
}

func createKeyValuePair(key string, value string) keyValuePair {
	return keyValuePair{
		key:   key,
		value: value,
	}
}

func TestParsingMetricQueryStringWorksAsExpected(t *testing.T) {
	testConfigs := []struct {
		name           string
		input          string
		expectedResult *QueryParameters
		shouldFail     bool
		errMessage     string
	}{
		{
			name:  "standard service response time",
			input: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectedResult: createExpectedResultFrom(
				t,
				createKeyValuePair("metricSelector", "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)"),
				createKeyValuePair("entitySelector", "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)"),
			),
		},
		{
			name:  "standard total error rate",
			input: "metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectedResult: createExpectedResultFrom(
				t,
				createKeyValuePair("metricSelector", "builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg"),
				createKeyValuePair("entitySelector", "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)"),
			),
		},
		{
			name:  "uri reserved character '/' and spaces in entity selector",
			input: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag2:tom / jerry)",
			expectedResult: createExpectedResultFrom(
				t,
				createKeyValuePair("metricSelector", "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)"),
				createKeyValuePair("entitySelector", "type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag2:tom / jerry)"),
			),
		},
		{
			name:  "uri reserved character '+' and spaces in metric selector",
			input: "metricSelector=(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Act ions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
			expectedResult: createExpectedResultFrom(
				t,
				createKeyValuePair(
					"metricSelector",
					"(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Act ions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))"),
			),
		},
		// Error cases below
		{
			// actually a tag 'my_tag:tom & jerry' would be totally fine from a Dynatrace API perspective
			name:       "uri reserved character '&' in entity selector fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag:tom & jerry)",
			shouldFail: true,
			errMessage: "jerry",
		},
		{
			name:       "no value for key entitySelector fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector",
			shouldFail: true,
			errMessage: "entitySelector",
		},
		{
			name:       "empty value for key entitySelector fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=",
			shouldFail: true,
			errMessage: "entitySelector=",
		},
		{
			name:       "empty key=value pair fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&",
			shouldFail: true,
			errMessage: "empty",
		},
		{
			name:       "empty input with spaces fails",
			input:      "   ",
			shouldFail: true,
			errMessage: "empty",
		},
		{
			name:       "standard service response time with additional unknown key fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)&myKey=myValue",
			shouldFail: true,
			errMessage: "unknown key",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewQueryParsing(tc.input).Parse()
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, actual)
				assert.Contains(t, err.Error(), tc.errMessage)
			} else {
				assert.NoError(t, err)
				assertEqual(t, tc.expectedResult, actual)
				assert.Empty(t, tc.errMessage, "fix test setup")
			}
		})
	}
}

func TestParsingMetricQueryStringAndEncodingItAgainWorksAsExpected(t *testing.T) {
	testConfigs := []struct {
		name           string
		input          string
		expectedResult string
		shouldFail     bool
		errMessage     string
	}{
		{
			name:           "standard service response time",
			input:          "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectedResult: "metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2850%29&entitySelector=type%28SERVICE%29%2Ctag%28keptn_managed%29%2Ctag%28keptn_service%3Amy-service%29",
		},
		{
			name:           "standard total error rate",
			input:          "metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)",
			expectedResult: "metricSelector=builtin%3Aservice.errors.total.rate%3Amerge%28%22dt.entity.service%22%29%3Aavg&entitySelector=type%28SERVICE%29%2Ctag%28keptn_managed%29%2Ctag%28keptn_service%3Amy-service%29",
		},
		{
			name:           "uri reserved character '/' and spaces in entity selector",
			input:          "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag2:tom / jerry)",
			expectedResult: "metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2850%29&entitySelector=type%28SERVICE%29%2Ctag%28keptn_managed%29%2Ctag%28keptn_service%3Amy-service%29%2Ctag%28my_tag2%3Atom+%2F+jerry%29",
		},
		{
			name:           "uri reserved character '+' and spaces in metric selector",
			input:          "metricSelector=(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
			expectedResult: "metricSelector=%28calc%3Aservice.%24rt_csm%3Afilter%28and%28eq%28Dimension%2C%22request+Actions.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave_Rest_customer-profile.%22%29%2Cin%28%22dt.entity.service%22%2CentitySelector%28%22type%28service%29%2CrequestAttribute%28~%22%24svc_id~%22%29%22%29%29%29%29%3AsplitBy%28%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%29%2F%28%28calc%3Aservice.%24reqcnt_csm%3Afilter%28and%28in%28%22dt.entity.service%22%2CentitySelector%28%22type%28service%29%2CrequestAttribute%28~%22%24svc_id~%22%29%22%29%29%2Ceq%28Dimension%2C%22request+Actions.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave_Rest_customer-profile.%22%29%29%29%29%3AsplitBy%28%29%3Asum%3Aauto%3Asort%28value%28sum%2Cdescending%29%29%2F%281%29%29",
		},
		// Error cases below:
		{
			// actually a tag 'my_tag:tom & jerry' would be totally fine from a Dynatrace API perspective
			name:       "uri reserved character '&' in entity selector fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service),tag(my_tag:tom & jerry)",
			shouldFail: true,
			errMessage: "jerry",
		},
		{
			name:       "no value for key entitySelector fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector",
			shouldFail: true,
			errMessage: "entitySelector",
		},
		{
			name:       "empty value for key entitySelector fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=",
			shouldFail: true,
			errMessage: "entitySelector=",
		},
		{
			name:       "empty key=value pair fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&",
			shouldFail: true,
			errMessage: "empty",
		},
		{
			name:       "empty input with spaces fails",
			input:      "   ",
			shouldFail: true,
			errMessage: "empty",
		},
		{
			name:       "standard service response time with additional unknown key fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:my-service)&myKey=myValue",
			shouldFail: true,
			errMessage: "unknown key",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewQueryParsing(tc.input).Parse()
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, actual)
				assert.Contains(t, err.Error(), tc.errMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, actual.Encode())
				assert.Empty(t, tc.errMessage, "fix test setup")
			}
		})
	}
}

func assertEqual(t *testing.T, expected *QueryParameters, actual *QueryParameters) {
	assertAllContained(t, actual, expected)
	assertAllContained(t, expected, actual)
}

func assertAllContained(t *testing.T, subset *QueryParameters, superset *QueryParameters) {
	subset.ForEach(func(keyFromSubset string, valueFromSubset string) {
		valueFromSuperSet, exists := superset.Get(keyFromSubset)
		assert.True(t, exists, "key: %s does not exist in super set", keyFromSubset)
		assert.Equal(t, valueFromSubset, valueFromSuperSet)
	})
}

func createExpectedResultFrom(t *testing.T, mapEntries ...keyValuePair) *QueryParameters {
	parameters := NewQueryParameters()
	for _, entry := range mapEntries {
		err := parameters.Add(entry.key, entry.value)
		if err != nil {
			assert.Fail(t, "incorrect test setup: %v", err)
		}
	}

	return parameters
}

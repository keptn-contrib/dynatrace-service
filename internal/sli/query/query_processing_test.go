package query

import (
	"net/http"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/stretchr/testify/assert"
)

const testDynatraceAPIToken = "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"

func TestGetSLIValueMetricsQueryErrorHandling(t *testing.T) {

	// TODO 2021-10-13: add rich error types as described in #358, including warnings
	tests := []struct {
		name                         string
		metricsQueryResponseFilename string
		expectedValue                float64
		shouldFail                   bool
		expectedErrorSubString       string
	}{
		{
			name:                         "One result, one data - want success",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_1result_1data_1value.json",
			expectedValue:                287.10692602352884 / 1000,
		},

		{
			name:                         "Request fails - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_constraints_violated.json",
			shouldFail:                   true,
			expectedErrorSubString:       "Dynatrace Metrics API returned an error",
		},

		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "Zero results 1 - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_0results_fake3.json",
			shouldFail:                   true,
			expectedErrorSubString:       "Dynatrace Metrics API returned an error",
		},

		{
			name:                         "One result, no data - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_1result_0data.json",
			shouldFail:                   true,
			expectedErrorSubString:       "Dynatrace Metrics API returned zero data points",
		},

		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "One result, one data, no values - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_1result_1data_0values_fake1.json",
			shouldFail:                   true,
			expectedErrorSubString:       "Dynatrace Metrics API returned zero data point values",
		},

		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "One result, one data, no values - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_1result_1data_0values_fake2.json",
			shouldFail:                   true,
			expectedErrorSubString:       "Dynatrace Metrics API returned zero data point values",
		},

		{
			name:                         "One result, one data, two values - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_1result_1data_2values.json",
			shouldFail:                   true,
			expectedErrorSubString:       "expected only a single data point value from Dynatrace Metrics API",
		},

		{
			name:                         "One result, two data - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_1result_2data.json",
			shouldFail:                   true,
			expectedErrorSubString:       "expected only a single data point from Dynatrace Metrics API",
		},

		{
			name:                         "Two results, one data - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_2results_1data.json",
			shouldFail:                   true,
			expectedErrorSubString:       "expected only a single result from Dynatrace Metrics API",
		},

		{
			name:                         "Two results, two data - want failure",
			metricsQueryResponseFilename: "./testdata/metrics_query_error_handling_test/metrics_query_2results_2data.json",
			shouldFail:                   true,
			expectedErrorSubString:       "expected only a single result from Dynatrace Metrics API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddStartsWith(dynatrace.MetricsQueryPath, tt.metricsQueryResponseFilename)

			value, err := runGetSLIValueTest(t, handler)

			assert.EqualValues(t, tt.expectedValue, value)
			if tt.shouldFail {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.expectedErrorSubString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// tests the GETSliValue function to return the proper datapoint
func TestGetSLIValue(t *testing.T) {

	okResponse := `{
		"totalCount": 8,
		"nextPageKey": null,
		"result": [
			{
				"metricId": "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
				"data": [
					{
						"dimensions": [],
						"timestamps": [
							1579097520000
						],
						"values": [
							8433.40
						]
					}
				]
			}
		]
	}`

	handler := test.NewPayloadBasedURLHandler(t)
	handler.AddStartsWith(dynatrace.MetricsQueryPath, []byte(okResponse))

	value, err := runGetSLIValueTest(t, handler)

	assert.NoError(t, err)
	assert.InDelta(t, 8.43340, value, 0.001)
}

// tests the GETSliValue function to return the proper datapoint with the old custom query format
func TestGetSLIValueWithOldAndNewCustomQueryFormat(t *testing.T) {

	okResponse := `{
		"totalCount": 8,
		"nextPageKey": null,
		"result": [
			{
				"metricId": "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
				"data": [
					{
						"dimensions": [],
						"timestamps": [
							1579097520000
						],
						"values": [
							8433.40
						]
					}
				]
			}
		]
	}`

	handler := test.NewPayloadBasedURLHandler(t)
	handler.AddStartsWith(dynatrace.MetricsQueryPath, []byte(okResponse))

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()

	testQueries := []string{
		"metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)",
		"?metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)",
		"builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)",
	}

	for _, testQuery := range testQueries {

		customQueries := make(map[string]string)
		customQueries[keptn.ResponseTimeP50] = testQuery

		p := createCustomQueryProcessing(t, keptnEvent, httpClient, keptn.NewCustomQueries(customQueries), start, end)
		value, err := p.GetSLIValue(keptn.ResponseTimeP50)

		assert.EqualValues(t, nil, err)
		assert.InDelta(t, 8.43340, value, 0.001)
	}
}

// Tests GetSLIValue with an empty result (no datapoints)
func TestGetSLIValueWithEmptyResult(t *testing.T) {

	okResponse := `{
		"totalCount": 4,
		"nextPageKey": null,
		"result": [
			{
				"metricId": "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
				"data": [
				]
			}
		]
	}`

	handler := test.NewPayloadBasedURLHandler(t)
	handler.AddStartsWith(dynatrace.MetricsQueryPath, []byte(okResponse))

	value, err := runGetSLIValueTest(t, handler)

	assert.Error(t, err)

	assert.EqualValues(t, 0.0, value)
}

/*
 * Helper function to test GetSLIValue
 */
func runGetSLIValueTest(t *testing.T, handler http.Handler) (float64, error) {
	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()

	dh := createQueryProcessing(t, keptnEvent, httpClient, start, end)

	return dh.GetSLIValue(keptn.ResponseTimeP50)
}

// Tests what happens when end time is too close to now
func TestGetSLISleep(t *testing.T) {
	okResponse := `{
		"totalCount": 3,
		"nextPageKey": null,
		"result": [
			{
				"metricId": "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
				"data": [
					{
						"dimensions": [],
						"timestamps": [
							1579097520000
						],
						"values": [
							8433.40
						]
					}
				]
			}
		]
	}`

	handler := test.NewPayloadBasedURLHandler(t)
	handler.AddStartsWith(dynatrace.MetricsQueryPath, []byte(okResponse))

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()

	start := time.Now().Add(-5 * time.Minute)
	// artificially increase end time to be in the future
	end := time.Now().Add(-80 * time.Second)
	dh := createQueryProcessing(t, keptnEvent, httpClient, start, end)

	value, err := dh.GetSLIValue(keptn.ResponseTimeP50)

	assert.NoError(t, err)
	assert.InDelta(t, 8.43340, value, 0.001)
}

// Tests the behaviour of the GetSLIValue function in case of a HTTP 400 return code
func TestGetSLIValueWithErrorResponse(t *testing.T) {
	handler := test.NewPayloadBasedURLHandler(t)
	handler.AddStartsWithError(dynatrace.MetricsQueryPath, http.StatusBadRequest, []byte{})

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()
	dh := createQueryProcessing(t, keptnEvent, httpClient, start, end)

	value, err := dh.GetSLIValue(keptn.Throughput)

	assert.Error(t, err)
	assert.EqualValues(t, 0.0, value)
}

func TestGetSLIValueForIndicator(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddStartsWith("/api/v2/slo", "./testdata/test_get_slo_id.json")
	handler.AddStartsWith("/api/v2/problems", "./testdata/test_get_problems.json")
	handler.AddStartsWith("/api/v2/securityProblems", "./testdata/test_get_securityproblems.json")

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()
	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()

	testConfigs := []struct {
		indicator string
		query     string
	}{
		{
			indicator: "problems",
			query:     "PV2;problemSelector=status(open)",
		},
		{
			indicator: "security_problems",
			query:     "SECPV2;problemEntity=status(open)",
		},
		{
			indicator: "RT_faster_500ms",
			query:     "SLO;524ca177-849b-3e8c-8175-42b93fbc33c5",
		},
	}

	for _, testConfig := range testConfigs {
		customQueries := make(map[string]string)
		customQueries[testConfig.indicator] = testConfig.query

		ret := createCustomQueryProcessing(t, keptnEvent, httpClient, keptn.NewCustomQueries(customQueries), startTime, endTime)

		res, err := ret.GetSLIValue(testConfig.indicator)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	}
}

func createQueryProcessing(t *testing.T, keptnEvent adapter.EventContentAdapter, httpClient *http.Client, start time.Time, end time.Time) *Processing {
	return createCustomQueryProcessing(
		t,
		keptnEvent,
		httpClient,
		keptn.NewEmptyCustomQueries(),
		start,
		end)
}

func createCustomQueryProcessing(t *testing.T, keptnEvent adapter.EventContentAdapter, httpClient *http.Client, queries *keptn.CustomQueries, start time.Time, end time.Time) *Processing {
	credentials, err := credentials.NewDynatraceCredentials("http://dynatrace", testDynatraceAPIToken)
	assert.NoError(t, err)

	return NewProcessing(
		dynatrace.NewClientWithHTTP(
			credentials,
			httpClient),
		keptnEvent,
		[]*keptnv2.SLIFilter{},
		queries,
		start,
		end)
}

func createDefaultTestEventData() adapter.EventContentAdapter {
	return &test.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}
}

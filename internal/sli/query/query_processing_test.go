package query

import (
	"errors"
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

const metricAPIURL = "/api/v2/metrics/query"

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
	handler.AddStartsWith(metricAPIURL, []byte(okResponse))

	value, err := runGetSLIValueTest(handler)

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
	handler.AddStartsWith(metricAPIURL, []byte(okResponse))

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

		p := createCustomQueryProcessing(keptnEvent, httpClient, keptn.NewCustomQueries(customQueries), start, end)
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
	handler.AddStartsWith(metricAPIURL, []byte(okResponse))

	value, err := runGetSLIValueTest(handler)

	assert.Error(t, err)

	assert.EqualValues(t, 0.0, value)
}

// Tests GetSLIValue without the expected metric in it
func TestGetSLIValueWithoutExpectedMetric(t *testing.T) {

	okResponse := `{
		"totalCount": 4,
		"nextPageKey": null,
		"result": [
			{
				"metricId": "something_else",
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
	handler.AddStartsWith(metricAPIURL, []byte(okResponse))

	value, err := runGetSLIValueTest(handler)

	assert.EqualValues(t, errors.New("No result matched the query's metric selector"), err)

	assert.EqualValues(t, 0.0, value)
}

/*
 * Helper function to test GetSLIValue
 */
func runGetSLIValueTest(handler http.Handler) (float64, error) {
	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()

	dh := createQueryProcessing(keptnEvent, httpClient, start, end)

	return dh.GetSLIValue(keptn.ResponseTimeP50)
}

func TestExtractMetricQueryFromMV2Query(t *testing.T) {

	testConfigs := []struct {
		name          string
		input         string
		expectedQuery string
		expectedUnit  string
		shouldFail    bool
	}{
		// these should fail
		{
			name:       "percent unit does not work",
			input:      "MV2;Percent;metricSelector=builtin:host.cpu.usage:merge(\"dt.entity.host\"):avg:names&entitySelector=type(HOST)",
			shouldFail: true,
		},
		{
			name:       "missing microsecond metric unit",
			input:      "MV2;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			shouldFail: true,
		},
		{
			name:       "missing mv2 prefix",
			input:      "MicroSecond;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			shouldFail: true,
		},
		{
			name:       "missing mv2 prefix",
			input:      "MV2;MicroSeconds;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			shouldFail: true,
		},
		// these should not fail
		{
			name:          "microsecond metric works",
			input:         "MV2;MicroSecond;metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectedQuery: "metricSelector=builtin:service.response.server:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"KeptnQualityGate~\")\")))):splitBy():percentile(90)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works 2",
			input:         "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.response.server:filter(and(in(\"dt.entity.service_method\",entitySelector(\"type(service_method),entityName(~\"/api/ui/v2/bootstrap~\")\")))):splitBy(\"dt.entity.service_method\"):percentile(90)",
			expectedQuery: "metricSelector=builtin:service.keyRequest.response.server:filter(and(in(\"dt.entity.service_method\",entitySelector(\"type(service_method),entityName(~\"/api/ui/v2/bootstrap~\")\")))):splitBy(\"dt.entity.service_method\"):percentile(90)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works - metric selector first",
			input:         "MV2;MicroSecond;metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedQuery: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works - entity selector first - MicroSecond unit",
			input:         "MV2;MicroSecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "MicroSecond",
		},
		{
			name:          "microsecond metric works - entity selector first - Microsecond unit",
			input:         "MV2;Microsecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "Microsecond",
		},
		{
			name:          "microsecond metric works - entity selector first - microsecond unit",
			input:         "MV2;microsecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "microsecond",
		},
		{
			name:          "microsecond metric works - entity selector first - microSecond unit",
			input:         "MV2;microSecond;entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedQuery: "entitySelector=type(SERVICE),tag(keptn_project:project1),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:direct)&metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)",
			expectedUnit:  "microSecond",
		},
		{
			name:          "byte metric works - Byte unit",
			input:         "MV2;Byte;metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedQuery: "metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedUnit:  "Byte",
		},
		{
			name:          "byte metric works - byte unit",
			input:         "MV2;byte;metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedQuery: "metricSelector=builtin:host.disk.avail:merge(\"dt.entity.host\"):merge(\"dt.entity.disk\")",
			expectedUnit:  "byte",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			adaptedQuery, unit, err := extractMetricQueryFromMV2Query(tc.input)
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Empty(t, adaptedQuery)
				assert.Empty(t, unit)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.expectedQuery, adaptedQuery)
				assert.EqualValues(t, tc.expectedUnit, unit)
			}
		})
	}
}

/*
// Tests what happens if the end-time is in the future
func TestGetSLIEndTimeFuture(t *testing.T) {
	keptnEvent := &GetSLITriggeredEvent{}
	keptnEvent.Project = "sockshop"
	keptnEvent.Stage = "dev"
	keptnEvent.Service = "carts"
	keptnEvent.DeploymentStrategy = ""

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
		}),
	)
	defer ts.Close()

	dh := NewRetrieval(ts.URL, keptnEvent, nil, nil)

	start := time.Now()
	// artificially increase end time to be in the future
	end := time.Now().Add(3 * time.Minute)
	value, err := dh.GetSLIValue(Throughput, start, end, []*events.SLIFilter{})

	assert.EqualValues(t, 0.0, value)
	assert.NotNil(t, err, nil)
	assert.EqualValues(t, "end time must not be in the future", err.Error())
}

// Tests what happens if start-time is after end-time
func TestGetSLIStartTimeAfterEndTime(t *testing.T) {
	keptnEvent := &GetSLITriggeredEvent{}
	keptnEvent.Project = "sockshop"
	keptnEvent.Stage = "dev"
	keptnEvent.Service = "carts"
	keptnEvent.DeploymentStrategy = ""

	dh := NewRetrieval("http://dynatrace", keptnEvent, nil, nil)

	start := time.Now()
	// artificially increase end time to be in the future
	end := time.Now().Add(-1 * time.Minute)
	value, err := dh.GetSLIValue(Throughput, start, end, []*events.SLIFilter{})

	assert.EqualValues(t, 0.0, value)
	assert.NotNil(t, err, nil)
	assert.EqualValues(t, "start time needs to be before end time", err.Error())
}
*/

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
	handler.AddStartsWith(metricAPIURL, []byte(okResponse))

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()

	start := time.Now().Add(-5 * time.Minute)
	// artificially increase end time to be in the future
	end := time.Now().Add(-80 * time.Second)
	dh := createQueryProcessing(keptnEvent, httpClient, start, end)

	value, err := dh.GetSLIValue(keptn.ResponseTimeP50)

	assert.NoError(t, err)
	assert.InDelta(t, 8.43340, value, 0.001)
}

// Tests the behaviour of the GetSLIValue function in case of a HTTP 400 return code
func TestGetSLIValueWithErrorResponse(t *testing.T) {
	handler := test.NewPayloadBasedURLHandler(t)
	handler.AddStartsWithError(metricAPIURL, http.StatusBadRequest, []byte{})

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := createDefaultTestEventData()

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()
	dh := createQueryProcessing(keptnEvent, httpClient, start, end)

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
			query:     "PV2;problemEntity=status(open)",
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

		ret := createCustomQueryProcessing(keptnEvent, httpClient, keptn.NewCustomQueries(customQueries), startTime, endTime)

		res, err := ret.GetSLIValue(testConfig.indicator)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	}
}

func createQueryProcessing(keptnEvent adapter.EventContentAdapter, httpClient *http.Client, start time.Time, end time.Time) *Processing {
	return createCustomQueryProcessing(
		keptnEvent,
		httpClient,
		keptn.NewEmptyCustomQueries(),
		start,
		end)
}

func createCustomQueryProcessing(keptnEvent adapter.EventContentAdapter, httpClient *http.Client, queries *keptn.CustomQueries, start time.Time, end time.Time) *Processing {
	return NewProcessing(
		dynatrace.NewClientWithHTTP(
			&credentials.DTCredentials{Tenant: "http://dynatrace"},
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

package sli

import (
	"errors"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// create a fake http client for integration tests
func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}

// tests the GETSliValue function to return the proper datapoint
func TestGetSLIValue(t *testing.T) {

	okResponse := `{
		"totalCount": 8,
		"nextPageKey": null,
		"result": [
			{
				"metricId": "builtin:service.response.time:merge(0):percentile(50)",
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

	value, err := runGetSLIValueTest(okResponse)

	assert.NoError(t, err)

	assert.InDelta(t, 8.43340, value, 0.001)
}

// tests the GETSliValue function to return the proper datapoint with the old custom query format
func TestGetSLIValueWithOldandNewCustomQueryFormat(t *testing.T) {

	okResponse := `{
		"totalCount": 8,
		"nextPageKey": null,
		"result": [
			{
				"metricId": "builtin:service.response.time:merge(0):percentile(50)",
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

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okResponse))
	})

	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	keptnEvent := &BaseKeptnEvent{}
	keptnEvent.Project = "sockshop"
	keptnEvent.Stage = "dev"
	keptnEvent.Service = "carts"
	keptnEvent.DeploymentStrategy = ""

	dh := createDynatraceHandler(keptnEvent, httpClient)

	// overwrite custom queries with the new format (starting with metricSelector=)
	customQueries := make(map[string]string)
	customQueries[ResponseTimeP50] = "metricSelector=builtin:service.response.time:merge(0):percentile(50)&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)"

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()
	value, err := dh.GetSLIValue(ResponseTimeP50, start, end, customQueries)

	assert.EqualValues(t, nil, err)
	assert.InDelta(t, 8.43340, value, 0.001)

	// now do the same but with the new format but with ?metricSelector= in front (the ? is not needed/wanted)
	customQueries = make(map[string]string)
	customQueries[ResponseTimeP50] = "?metricSelector=builtin:service.response.time:merge(0):percentile(50)&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)"

	start = time.Unix(1571649084, 0).UTC()
	end = time.Unix(1571649085, 0).UTC()
	value, err = dh.GetSLIValue(ResponseTimeP50, start, end, customQueries)

	assert.EqualValues(t, nil, err)
	assert.InDelta(t, 8.43340, value, 0.001)

	// now do the same but with the old format ($metricName?scope=...)
	customQueries = make(map[string]string)
	customQueries[ResponseTimeP50] = "builtin:service.response.time:merge(0):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)"

	start = time.Unix(1571649084, 0).UTC()
	end = time.Unix(1571649085, 0).UTC()
	value, err = dh.GetSLIValue(ResponseTimeP50, start, end, customQueries)

	assert.EqualValues(t, nil, err)
	assert.InDelta(t, 8.43340, value, 0.001)
}

// Tests GetSLIValue with an empty result (no datapoints)
func TestGetSLIValueWithEmptyResult(t *testing.T) {

	okResponse := `{
    "totalCount": 4,
    "nextPageKey": null,
	"result": [
		{
			"metricId": "builtin:service.response.time:merge(0):percentile(50)",
			"data": [
			]
		}
	]
}`

	value, err := runGetSLIValueTest(okResponse)

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

	value, err := runGetSLIValueTest(okResponse)

	assert.EqualValues(t, errors.New("No result matched the query's metric selector"), err)

	assert.EqualValues(t, 0.0, value)
}

/*
 * Helper function to test GetSLIValue
 */
func runGetSLIValueTest(okResponse string) (float64, error) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okResponse))
	})

	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	keptnEvent := &BaseKeptnEvent{}
	keptnEvent.Project = "sockshop"
	keptnEvent.Stage = "dev"
	keptnEvent.Service = "carts"
	keptnEvent.DeploymentStrategy = ""

	dh := createDynatraceHandler(keptnEvent, httpClient)

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()

	return dh.GetSLIValue(ResponseTimeP50, start, end, nil)
}

func TestGetSLIValueWithMV2Prefix(t *testing.T) {

	metricsQuery := "MV2;Percent;metricSelector=builtin:host.cpu.usage:merge(0):avg:names&entitySelector=type(HOST)"

	if strings.HasPrefix(metricsQuery, "MV2;") {
		metricsQuery = metricsQuery[4:]
		assert.EqualValues(t, metricsQuery, "Percent;metricSelector=builtin:host.cpu.usage:merge(0):avg:names&entitySelector=type(HOST)")
		queryStartIndex := strings.Index(metricsQuery, ";")
		metricUnit := metricsQuery[:queryStartIndex]
		assert.EqualValues(t, metricUnit, "Percent")
		metricsQuery = metricsQuery[queryStartIndex+1:]
		assert.EqualValues(t, metricsQuery, "metricSelector=builtin:host.cpu.usage:merge(0):avg:names&entitySelector=type(HOST)")
	}
}

/*
// Tests what happens if the end-time is in the future
func TestGetSLIEndTimeFuture(t *testing.T) {
	keptnEvent := &BaseKeptnEvent{}
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

	dh := NewDynatraceHandler(ts.URL, keptnEvent, nil, nil)

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
	keptnEvent := &BaseKeptnEvent{}
	keptnEvent.Project = "sockshop"
	keptnEvent.Stage = "dev"
	keptnEvent.Service = "carts"
	keptnEvent.DeploymentStrategy = ""

	dh := NewDynatraceHandler("http://dynatrace", keptnEvent, nil, nil)

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
				"metricId": "builtin:service.response.time:merge(0):percentile(50)",
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

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okResponse))
	})

	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	keptnEvent := &BaseKeptnEvent{}
	keptnEvent.Project = "sockshop"
	keptnEvent.Stage = "dev"
	keptnEvent.Service = "carts"
	keptnEvent.DeploymentStrategy = ""

	dh := createDynatraceHandler(keptnEvent, httpClient)

	start := time.Now().Add(-5 * time.Minute)
	// artificially increase end time to be in the future
	end := time.Now().Add(-80 * time.Second)
	value, err := dh.GetSLIValue(ResponseTimeP50, start, end, nil)

	assert.InDelta(t, 8.43340, value, 0.001)
	assert.Nil(t, err)
}

// Tests the behaviour of the GetSLIValue function in case of a HTTP 400 return code
func TestGetSLIValueWithErrorResponse(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.Write([]byte(response))
		w.WriteHeader(http.StatusBadRequest)
	})

	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	keptnEvent := &BaseKeptnEvent{}
	keptnEvent.Project = "sockshop"
	keptnEvent.Stage = "dev"
	keptnEvent.Service = "carts"
	keptnEvent.DeploymentStrategy = ""

	dh := createDynatraceHandler(keptnEvent, httpClient)

	start := time.Unix(1571649084, 0).UTC()
	end := time.Unix(1571649085, 0).UTC()
	value, err := dh.GetSLIValue(Throughput, start, end, nil)

	assert.EqualValues(t, 0.0, value)
	assert.NotNil(t, err, nil)
}

func createDynatraceHandler(keptnEvent *BaseKeptnEvent, httpClient *http.Client) *Handler {
	dh := NewDynatraceHandler(
		keptnEvent,
		dynatrace.NewClient(
			&credentials.DTCredentials{Tenant: "http://dynatrace"}))
	dh.client.HTTPClient = httpClient

	return dh
}

package dashboard

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestQueryDynatraceDashboardForSLIs(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)
	// we handle these if the URLs are a full match
	handler.AddExact(dynatrace.DashboardsPath, "./testdata/test_get_dashboards.json")
	handler.AddExact(dynatrace.DashboardsPath+"/12345678-1111-4444-8888-123456789012", "./testdata/test_get_dashboards_id.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.processCount", "./testdata/test_get_metrics_processcount.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/test_get_metrics_svcresponsetime.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.mem.workingSetSize", "./testdata/test_get_metrics_workingsetsize.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.cpu.usage", "./testdata/test_get_metrics_cpuusage.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.errors.server.rate", "./testdata/test_get_metrics_errorrate.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.requestCount.total", "./testdata/test_get_metrics_requestcount.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.cpu.usage", "./testdata/test_get_metrics_hostcpuusage.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.mem.usage", "./testdata/test_get_metrics_hostmemusage.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.disk.queueLength", "./testdata/test_get_metrics_hostdiskqueue.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.nonDbChildCallCount", "./testdata/test_get_metrics_nondbcallcount.json")
	handler.AddExact(dynatrace.MetricsPath+"/jmeter.usermetrics.transaction.meantime", "./testdata/test_get_metrics_jmeter_usermetrics_transaction_meantime.json")
	// we handle these if the URL "starts with"
	handler.AddStartsWith(dynatrace.MetricsQueryPath, "./testdata/test_get_metrics_query.json")
	handler.AddStartsWith("/api/v2/slo", "./testdata/test_get_slo_id.json")
	handler.AddStartsWith("/api/v2/problems", "./testdata/test_get_problems.json")
	handler.AddStartsWith("/api/v2/securityProblems", "./testdata/test_get_securityproblems.json")

	querying, _, teardown := createQueryingWithHandler(keptnEvent, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	result, err := querying.GetSLIValues(common.DynatraceConfigDashboardQUERY, startTime, endTime)

	assert.Nil(t, err)
	assert.NotNil(t, result, "No result returned")
	assert.NotNil(t, result.dashboardLink, "No dashboard link label generated")
	assert.NotNil(t, result.dashboard, "No Dashboard JSON returned")
	assert.NotNil(t, result.sli, "No SLI returned")
	assert.NotNil(t, result.slo, "No SLO returned")
	assert.NotNil(t, result.sliResults, "No SLI Results returned")

	const expectedSLOs = 14
	assert.Equal(t, expectedSLOs, len(result.sli.Indicators))
	assert.Equal(t, expectedSLOs, len(result.slo.Objectives))
	assert.EqualValues(t, &keptnapi.SLOScore{Pass: "90%", Warning: "70%"}, result.slo.TotalScore)
	assert.EqualValues(
		t,
		&keptnapi.SLOComparison{
			CompareWith:               "single_result",
			IncludeResultWithScore:    "pass",
			NumberOfComparisonResults: 1,
			AggregateFunction:         "avg",
		},
		result.slo.Comparison)
	assert.Equal(t, expectedSLOs, len(result.sliResults))
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") and there is no dashboard already stored
// in Keptn resources then we do not do anything
func TestNoQueryingOfDashboardNecessaryDueToNotSpecifiedAndNotDashboardInKeptn(t *testing.T) {
	// we don't care about event data in this case
	ev := &test.EventData{}

	// no need for a handler as we should not query Dynatrace API here
	querying, _, teardown := createCustomQuerying(ev, nil, DashboardReaderMock{err: "no dashboard"})
	defer teardown()

	const dashboardID = ""

	result, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

	assert.Nil(t, result)
	assert.Nil(t, err)
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") but there is a dashboard already stored
// in Keptn resources then we do the fallback to querying Dynatrace API for a dashboard
// also we will fail because of a failing request to Dynatrace API
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithFailingDynatraceRequest(t *testing.T) {
	// we don't care about event data in this case
	ev := &test.EventData{}

	// we add a handler to simulate a failing dashboards API request (401 in this case)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath, http.StatusUnauthorized, "./testdata/dynatrace_missing_authorization_error.json")

	// we don't care about the content of the dashboard here, because it just should not be empty!
	// also we don't add a handler to simulate a failing request (404) in this case.
	querying, _, teardown := createCustomQuerying(ev, handler, DashboardReaderMock{content: "some dashboard content"})
	defer teardown()

	const dashboardID = ""

	result, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

	assert.Nil(t, result)
	assert.Nil(t, err)
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") but there is a dashboard already stored
// in Keptn resources then we do the fallback to querying Dynatrace API for a dashboard
//
// we will find dashboards but none with a name, that would match the required format: KQG;project=%project%;service=%service%;stage=%stage%;xxx
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithSuccessfulDynatraceRequest(t *testing.T) {
	// we don't care about event data in this case
	ev := &test.EventData{}

	// we add a handler to simulate an successful Dashboards API request in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath, "./testdata/test_query_dynatrace_dashboard_dashboards.json")

	// we don't care about the content of the dashboard here, because it just should not be empty!
	querying, _, teardown := createCustomQuerying(ev, handler, DashboardReaderMock{content: "some dashboard content"})
	defer teardown()

	const dashboardID = ""

	result, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

	assert.Nil(t, result)
	assert.Nil(t, err)
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") but there is a dashboard already stored
// in Keptn resources then we do the fallback to querying Dynatrace API for a dashboard
//
// we will find dashboards and one with a name, that would match the required format: KQG;project=%project%;service=%service%;stage=%stage%;xxx
// also the retrieved dashboard is the same as the stored one and we have enabled "KQG.QueryBehavior=ParseOnChange" in our markdown tile
// so no need to do anything with the contents of the dashboard here.
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithSuccessfulDynatraceRequestAndMatchingDashboard(t *testing.T) {
	// we need to match project, stage & service against the dashboard name
	ev := createKeptnEvent("project-1", "stage-3", "service-7")

	const matchingDashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9b"
	// we need to make sure to use the "processed" one and not the original Dynatrace JSON, because we do have different
	// models for our DTOs (Tile structures are generic in our part - we need to support any tile type)
	const storedDashboardFile = "./testdata/test_query_dynatrace_dashboard_dashboard_kqg.json"

	// we add a handle to simulate an successful Dashboards API request in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath, "./testdata/test_query_dynatrace_dashboard_dashboards_kqg.json")
	handler.AddExact(dynatrace.DashboardsPath+"/"+matchingDashboardID, storedDashboardFile)

	dashboardContent, err := ioutil.ReadFile(storedDashboardFile)
	if err != nil {
		panic(err)
	}

	// we need to make sure that the mocked reader returns the "processed" Dynatrace Dashboard
	querying, url, teardown := createCustomQuerying(ev, handler, DashboardReaderMock{content: string(dashboardContent)})
	defer teardown()

	const dashboardID = ""

	from := time.Date(2021, 9, 17, 7, 0, 0, 0, time.UTC)
	to := time.Date(2021, 9, 17, 8, 0, 0, 0, time.UTC)
	actualResult, err := querying.GetSLIValues(dashboardID, from, to)

	expectedResult := NewQueryResultFrom(&DashboardLink{
		apiURL:         url,
		startTimestamp: from,
		endTimestamp:   to,
		dashboardID:    matchingDashboardID,
	})

	assert.Nil(t, err)
	assert.EqualValues(t, expectedResult, actualResult)
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// also the retrieved dashboard is the same as the stored one and we have enabled "KQG.QueryBehavior=ParseOnChange" in our markdown tile
// so no need to do anything with the contents of the dashboard here.
// (this is basically the same as above without querying the dashboards API for a matching dashboard beforehand)
func TestRetrieveDashboardWithValidIDAndStoredDashboardInKeptnIsTheSame(t *testing.T) {
	// we need to match project, stage & service against the dashboard name
	ev := createKeptnEvent("project-1", "stage-3", "service-7")

	const dashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9b"
	// we need to make sure to use the "processed" one and not the original Dynatrace JSON, because we do have different
	// models for our DTOs (Tile structures are generic in our part - we need to support any tile type)
	const storedDashboardFile = "./testdata/test_query_dynatrace_dashboard_dashboard_kqg.json"

	// we add a handle to simulate an successful Dashboards API request in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+dashboardID, storedDashboardFile)

	dashboardContent, err := ioutil.ReadFile(storedDashboardFile)
	if err != nil {
		panic(err)
	}

	// we need to make sure that the mocked reader returns the "processed" Dynatrace Dashboard
	querying, url, teardown := createCustomQuerying(ev, handler, DashboardReaderMock{content: string(dashboardContent)})
	defer teardown()

	from := time.Date(2021, 9, 17, 7, 0, 0, 0, time.UTC)
	to := time.Date(2021, 9, 17, 8, 0, 0, 0, time.UTC)
	actualResult, err := querying.GetSLIValues(dashboardID, from, to)

	expectedResult := NewQueryResultFrom(&DashboardLink{
		apiURL:         url,
		startTimestamp: from,
		endTimestamp:   to,
		dashboardID:    dashboardID,
	})

	assert.Nil(t, err)
	assert.EqualValues(t, expectedResult, actualResult)
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// also the ID of the dashboard we try to retrieve was not found
func TestRetrieveDashboardWithUnknownButValidID(t *testing.T) {
	// we need do not care about the event here
	ev := &test.EventData{}

	const dashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9c"

	// we add a handler to simulate a very concrete 404 Dashboards API request/response in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath+"/"+dashboardID, http.StatusNotFound, "./testdata/test_query_dynatrace_dashboard_dashboard_id_not_found.json")

	// we also do not care about the dashboard that would be returned by keptn
	querying, _, teardown := createCustomQuerying(ev, handler, DashboardReaderMock{})
	defer teardown()

	actualResult, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

	assert.Error(t, err)
	assert.Nil(t, actualResult)

	var apiErr *dynatrace.APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusNotFound, apiErr.Code())
		assert.Contains(t, apiErr.Message(), dashboardID)
	}
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// it could happen that you have a copy/paste error in your ID (invalid UUID) so we will see a different error
func TestRetrieveDashboardWithInvalidID(t *testing.T) {
	// we need do not care about the event here
	ev := &test.EventData{}

	const dashboardID = "definitely-invalid-uuid"

	// we add a handler to simulate a very concrete 400 Dashboards API request/response in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath+"/"+dashboardID, http.StatusBadRequest, "./testdata/test_query_dynatrace_dashboard_dashboard_id_not_valid.json")

	// we also do not care about the dashboard that would be returned by keptn
	querying, _, teardown := createCustomQuerying(ev, handler, DashboardReaderMock{})
	defer teardown()

	actualResult, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

	assert.Error(t, err)
	assert.Nil(t, actualResult)

	var apiErr *dynatrace.APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusBadRequest, apiErr.Code())
	}
	assert.Contains(t, err.Error(), "UUID")
}

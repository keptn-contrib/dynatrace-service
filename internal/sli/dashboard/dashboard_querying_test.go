package dashboard

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
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

	// TODO 2021-10-11: Check: these test data files may be out of date as the result data elements do not include a dimensionMap element
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28PROCESS_GROUP_INSTANCE%29&from=1571649084000&metricSelector=builtin%3Atech.generic.processCount%3Amerge%28%22dt.entity.process_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_tech.generic.processCount.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2CentityId%28%22SERVICE-086C46F600BA1DC6%22%29%2Ctag%28%22keptn_deployment%3Aprimary%22%29&from=1571649084000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_service.response.time_merge_service_percentile_95.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1571649084000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2890.000000%29%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_service.response.time_merge_service_percentile_90.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1571649084000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2850.000000%29%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_service.response.time_merge_service_percentile_50.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28PROCESS_GROUP_INSTANCE%29&from=1571649084000&metricSelector=builtin%3Atech.generic.mem.workingSetSize%3Amerge%28%22dt.entity.process_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_tech.generic.mem.workingSetSize.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28PROCESS_GROUP_INSTANCE%29&from=1571649084000&metricSelector=builtin%3Atech.generic.cpu.usage%3Amerge%28%22dt.entity.process_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_tech.generic.cpu.usage.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1571649084000&metricSelector=builtin%3Aservice.errors.server.rate%3Amerge%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_service.errors.server.rate.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1571649084000&metricSelector=builtin%3Aservice.requestCount.total%3Amerge%28%22dt.entity.service%22%29%3Avalue%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_service.requestCount.total.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28HOST%29&from=1571649084000&metricSelector=builtin%3Ahost.cpu.usage%3Amerge%28%22dt.entity.host%22%29%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_host.cpu.usage.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28HOST%29&from=1571649084000&metricSelector=builtin%3Ahost.mem.usage%3Amerge%28%22dt.entity.host%22%29%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_host.mem.usage.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28HOST%29&from=1571649084000&metricSelector=builtin%3Ahost.disk.queueLength%3Amerge%28%22dt.entity.disk%22%29%3Amerge%28%22dt.entity.host%22%29%3Amax%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_host.disk.queueLength.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1571649084000&metricSelector=builtin%3Aservice.nonDbChildCallCount%3Amerge%28%22dt.entity.service%22%29%3Avalue%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_service.nonDbChildCallCount.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=entityId%28SERVICE-FFD81F003E39B468%29&from=1571649084000&metricSelector=jmeter.usermetrics.transaction.meantime%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_jmeter.usermetrics.transaction.meantime.json")

	handler.AddStartsWith("/api/v2/slo", "./testdata/test_get_slo_id.json")
	handler.AddStartsWith("/api/v2/problems", "./testdata/test_get_problems.json")
	handler.AddStartsWith("/api/v2/securityProblems", "./testdata/test_get_securityproblems.json")

	querying, _, teardown := createQueryingWithHandler(keptnEvent, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	result, dbProcessed, err := querying.GetSLIValues(common.DynatraceConfigDashboardQUERY, startTime, endTime)

	assert.Nil(t, err)
	assert.True(t, dbProcessed)
	assert.NotNil(t, result, "No result returned")
	assert.NotNil(t, result.dashboardLink, "No dashboard link label generated")
	assert.NotNil(t, result.dashboard, "No Dashboard JSON returned")
	assert.NotNil(t, result.sli, "No SLI returned")
	assert.NotNil(t, result.slo, "No SLO returned")
	assert.NotNil(t, result.sliResults, "No SLI Results returned")

	const expectedSLOs = 17
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
	actualResult, dbProcessed, err := querying.GetSLIValues(dashboardID, from, to)

	expectedResult := NewQueryResultFrom(&DashboardLink{
		apiURL:         url,
		startTimestamp: from,
		endTimestamp:   to,
		dashboardID:    dashboardID,
	})

	assert.NoError(t, err)
	assert.False(t, dbProcessed)
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

	actualResult, dbProcessed, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

	assert.Error(t, err)
	assert.False(t, dbProcessed)
	assert.Nil(t, actualResult)

	var apiErr *dynatrace.APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusNotFound, apiErr.Code())
		assert.Contains(t, apiErr.Message(), dashboardID)
	}
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will try to retrieve the dashboard via the Dynatrace API.
// this should fail if we cannot access Keptn resources or if a dashboard is stored in Keptn, but has no content
func TestRetrieveDashboardFailingBecauseOfErrorsInKeptn(t *testing.T) {
	testConfigs := []struct {
		name string
		err  error
	}{
		{
			name: "error while retrieving dashboard from Keptn",
			err:  &keptn.ResourceRetrievalFailedError{},
		},
		{
			name: "error because dashboard content is empty",
			err:  &keptn.ResourceEmptyError{},
		},
	}

	const dashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9c"

	// we should not be able to query anything, but fail early
	handler := test.NewFileBasedURLHandler(t)

	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			querying, _, teardown := createCustomQuerying(
				&test.EventData{},
				handler,
				DashboardReaderMock{err: tc.err})
			defer teardown()

			actualResult, dbProcessed, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

			assert.Error(t, err)
			assert.False(t, dbProcessed)
			assert.Nil(t, actualResult)

			var resErr = reflect.New(reflect.TypeOf(tc.err)).Interface()
			assert.ErrorAs(t, err, &resErr)
		})
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

	actualResult, dbProcessed, err := querying.GetSLIValues(dashboardID, time.Now(), time.Now())

	assert.Error(t, err)
	assert.False(t, dbProcessed)
	assert.Nil(t, actualResult)

	var apiErr *dynatrace.APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusBadRequest, apiErr.Code())
	}
	assert.Contains(t, err.Error(), "UUID")
}

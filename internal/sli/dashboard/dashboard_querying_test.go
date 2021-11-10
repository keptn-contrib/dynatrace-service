package dashboard

import (
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

	// TODO 2021-10-11: Check: these test data files may be out of date as the result data elements do not include a dimensionMap element
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=type%28PROCESS_GROUP_INSTANCE%29&from=1571649084000&metricSelector=builtin%3Atech.generic.processCount%3Amerge%28%22dt.entity.process_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		"./testdata/test_get_metrics_query_tech.generic.processCount.json")
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

	handler.AddStartsWith("/api/v2/slo", "./testdata/test_get_slo_id.json")
	handler.AddStartsWith("/api/v2/problems", "./testdata/test_get_problems.json")
	handler.AddStartsWith("/api/v2/securityProblems", "./testdata/test_get_securityproblems.json")

	querying, _, teardown := createQueryingWithHandler(t, keptnEvent, handler)
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

	for _, objective := range result.slo.Objectives {
		assert.NotNil(t, objective)
	}

	for _, sliResult := range result.sliResults {
		assert.True(t, sliResult.Success)
	}
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

	querying, _, teardown := createCustomQuerying(t, ev, handler)
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

	querying, _, teardown := createCustomQuerying(t, ev, handler)
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

package sli

import (
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag tests splitting by key service request and filtering by tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_autotag/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE_METHOD%29%2CfromRelationships.isServiceMethodOfService%28type%28SERVICE%29%2Ctag%28%22keptnmanager%22%29%29", "builtin%3Aservice.keyRequest.totalProcessingTime%3AsplitBy%28%22dt.entity.service_method%22%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_splitby_servicekeyrequest_filterby_autotag.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.keyRequest.totalProcessingTime", filepath.Join(testDataFolder, "metrics_get_by_id_builtin_servicekeyrequest_totalprocessingtime.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query_builtin_servicekeyrequest_totalprocessingtime.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("processing_time_findlocations", 18227.56816390859, expectedMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("processing_time_getjourneybyid", 2860.6086572438163, expectedMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("processing_time_getjourneypagebytenant", 15964.052631578946, expectedMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("processing_time_findjourneys", 23587.584492453388, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIButNoSeries tests a custom charting tile with an SLI name defined, i.e. in the title, but no series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIButNoSeries(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_no_series_test/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_sli_name_no_series.json"))

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("empty_chart"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries tests a custom charting tile with an SLI name defined and two series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_two_series_test/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_sli_name_two_series.json"))

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("services_response_time_two_series"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy tests a custom charting tile with neither split by or filter by defined.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_no_filterby/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_no_splitby_no_filterby.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_get_by_id_builtin_service_responsetime.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query_builtin_service_responsetime.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 54896.44626158664, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod tests a custom charting tile that splits by service key request and filters by service of service method.
// This is will result in a SLIResult with failure, as this is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_serviceofservicemethod/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_splitby_servicekeyrequest_filterby_serviceofservicemethod.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.keyRequest.totalProcessingTime", filepath.Join(testDataFolder, "metrics_get_by_id_builtin_servicekeyrequest_totalprocessingtime.json"))

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("tpt_key_requests_journeyservice"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag tests a custom charting tile that splits by service and filters by tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_autotag/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22keptn_managed%22%29", "builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_splitby_service_filterby_autotag.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_get_by_id_builtin_service_responsetime.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query_builtin_service_responsetime.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags_easytravelservice", 132278.23461853978, expectedMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags_journeyservice", 20256.493055555555, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterBySpecificEntity tests a custom charting tile that splits by service and filters by specific entity.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterBySpecificEntity(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_specificentity/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CentityId%28%22SERVICE-C6876D601CA5DDFD%22%29", "builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_splitby_service_filterby_specificentity.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_get_by_id_builtin_service_responsetime.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query_builtin_service_responsetime.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_specificentity", 57974.262650996854, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech tests a custom charting tile that filters by service software tech.
// This is will result in a SLIResult with failure, as the SERVICE_SOFTWARE_TECH filter is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_filterby_servicesoftwaretech/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_filterby_servicetopg.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_get_by_id_builtin_service_responsetime.json"))

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("svc_rt_p95"))
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_WorkerProcessCount(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/worker_process_count_avg/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28PROCESS_GROUP_INSTANCE%29", "builtin%3Atech.generic.processCount%3AsplitBy%28%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_worker_process_count_avg.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.processCount", filepath.Join(testDataFolder, "metrics_builtin_tech_generic_processCount.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_tech_generic_processCount_avg.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("proc_count", 48.89413206124561, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP90(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p90/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2890.000000%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_response_time_p90.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_builtin_service_response_time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_service_response_time_p90.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_rt_p90", 35008.21102453578, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP50(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p50/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2850.000000%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_response_time_p50.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_builtin_service_response_time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_service_response_time_p50.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_rt_p50", 1500.1115152382822, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessMemoryAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_memory_avg/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28PROCESS_GROUP_INSTANCE%29", "builtin%3Atech.generic.mem.workingSetSize%3AsplitBy%28%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_process_memory_avg.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.mem.workingSetSize", filepath.Join(testDataFolder, "metrics_builtin_tech_generic_mem_workingSetSize.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_tech_generic_mem_workingsetsize_avg.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("process_memory", 1.4800280435645182e+09, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessCPUAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_cpu_avg/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28PROCESS_GROUP_INSTANCE%29", "builtin%3Atech.generic.cpu.usage%3AsplitBy%28%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_process_cpu_avg.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.cpu.usage", filepath.Join(testDataFolder, "metrics_builtin_tech_generic_cpu_usage.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_tech_generic_cpu_usage_avg.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("process_cpu", 14.223367878298156, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_Throughput(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/throughput/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.requestCount.total%3AsplitBy%28%29%3Avalue%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_throughput.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.requestCount.total", filepath.Join(testDataFolder, "metrics_builtin_service_requestcount_total.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_service_requestcount_total_value.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_tp_min", 68044716, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostCPUUsageAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_cpu_usage_avg/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28HOST%29", "builtin%3Ahost.cpu.usage%3AsplitBy%28%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_host_cpu_usage_avg.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.cpu.usage", filepath.Join(testDataFolder, "metrics_builtin_host_cpu_usage.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_host_cpu_usage_avg.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_cpu", 20.309976061722214, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostMemoryUsageAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_mem_usage_avg/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28HOST%29", "builtin%3Ahost.mem.usage%3AsplitBy%28%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_host_mem_usage_avg.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.mem.usage", filepath.Join(testDataFolder, "metrics_builtin_host_mem_usage.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_host_mem_usage_avg.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_mem", 45.443796324058994, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostDiskQueueLengthMax(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_disk_queuelength_max/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28HOST%29", "builtin%3Ahost.disk.queueLength%3AsplitBy%28%29%3Amax%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_host_disk_queuelength_max.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.disk.queueLength", filepath.Join(testDataFolder, "metrics_builtin_host_disk_queuelength.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_host_disk_queuelength_max.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_disk_queue", 100, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_NonDbChildCallCount(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/non_db_child_call_count/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.nonDbChildCallCount%3AsplitBy%28%29%3Avalue%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_non_db_child_call_count.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.nonDbChildCallCount", filepath.Join(testDataFolder, "metrics_builtin_service_nondbchildcallcount.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_service_nondbchildcallcount.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc2svc_calls", 13657068, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile tests an additional custom charting tile with exclude set to true is skipped.
// This results in success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/excluded_tile/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_excluded_tile.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_get_by_id_builtin_service_responsetime.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query_builtin_service_responsetime.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 29313.12208863131, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

package sli

import (
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

var testCustomChartingGetSLIEventData = createTestGetSLIEventDataWithStartAndEnd("2021-09-17T07:00:00.000Z", "2021-09-17T08:00:00.000Z")

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag tests splitting by key service request and filtering by tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_autotag/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_servicekeyrequest_filterby_autotag.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.keyRequest.totalProcessingTime", testDataFolder+"metrics_get_by_id_builtin_servicekeyrequest_totalprocessingtime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE_METHOD%29%2CfromRelationships.isServiceMethodOfService%28type%28SERVICE%29%2Ctag%28%22keptnmanager%22%29%29&from=1631862000000&metricSelector=builtin%3Aservice.keyRequest.totalProcessingTime%3AsplitBy%28%22dt.entity.service_method%22%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_servicekeyrequest_totalprocessingtime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("processing_time_findlocations", 18.22756816390859, "MV2;MicroSecond;entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\"))&metricSelector=builtin:service.keyRequest.totalProcessingTime:splitBy(\"dt.entity.service_method\"):avg:names"),
		createSuccessfulDashboardSLIResultAssertionsFunc("processing_time_getjourneybyid", 2.8606086572438163, "MV2;MicroSecond;entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\"))&metricSelector=builtin:service.keyRequest.totalProcessingTime:splitBy(\"dt.entity.service_method\"):avg:names"),
		createSuccessfulDashboardSLIResultAssertionsFunc("processing_time_getjourneypagebytenant", 15.964052631578946, "MV2;MicroSecond;entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\"))&metricSelector=builtin:service.keyRequest.totalProcessingTime:splitBy(\"dt.entity.service_method\"):avg:names"),
		createSuccessfulDashboardSLIResultAssertionsFunc("processing_time_findjourneys", 23.587584492453388, "MV2;MicroSecond;entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\"))&metricSelector=builtin:service.keyRequest.totalProcessingTime:splitBy(\"dt.entity.service_method\"):avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIButNoSeries tests a custom charting tile with an SLI name defined, i.e. in the title, but no series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIButNoSeries(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_no_series_test/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_sli_name_no_series.json")

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testCustomChartingGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("empty_chart"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries tests a custom charting tile with an SLI name defined and two series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_two_series_test/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_sli_name_two_series.json")

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testCustomChartingGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("services_response_time_two_series"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy tests a custom charting tile with neither split by or filter by defined.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_no_filterby/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_no_splitby_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_get_by_id_builtin_service_responsetime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_service_responsetime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("service_response_time", 29.31312208863131, "MV2;MicroSecond;entitySelector=type(SERVICE)&metricSelector=builtin:service.response.time:splitBy():avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod tests a custom charting tile that splits by service key request and filters by service of service method.
// This is will result in a SLIResult with failure, as this is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_serviceofservicemethod/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_servicekeyrequest_filterby_serviceofservicemethod.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.keyRequest.totalProcessingTime", testDataFolder+"metrics_get_by_id_builtin_servicekeyrequest_totalprocessingtime.json")

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testCustomChartingGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("tpt_key_requests_journeyservice"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag tests a custom charting tile that splits by service and filters by tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_autotag/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_service_filterby_autotag.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_get_by_id_builtin_service_responsetime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28%22keptn_managed%22%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_service_responsetime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags_easytravelservice", 132.27823461853978, "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"keptn_managed\")&metricSelector=builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:names"),
		createSuccessfulDashboardSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags_journeyservice", 20.256493055555555, "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"keptn_managed\")&metricSelector=builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterBySpecificEntity tests a custom charting tile that splits by service and filters by specific entity.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterBySpecificEntity(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_specificentity/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_service_filterby_specificentity.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_get_by_id_builtin_service_responsetime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2CentityId%28%22SERVICE-F2455557EF67362B%22%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_service_responsetime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_specificentity", 20.256493055555555, "MV2;MicroSecond;entitySelector=type(SERVICE),entityId(\"SERVICE-F2455557EF67362B\")&metricSelector=builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech tests a custom charting tile that filters by service software tech.
// This is will result in a SLIResult with failure, as the SERVICE_SOFTWARE_TECH filter is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_filterby_servicesoftwaretech/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_filterby_servicetopg.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_get_by_id_builtin_service_responsetime.json")

	rClient := &uploadErrorResourceClientMock{t: t}
	getSLIEventData := createTestGetSLIEventDataWithStartAndEnd("2019-10-21T09:11:24Z", "2019-10-21T09:11:25Z")
	runAndAssertThatDashboardTestIsCorrect(t, getSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("svc_rt_p95"))
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_WorkerProcessCount(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/worker_process_count_avg/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_worker_process_count_avg.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.processCount", testDataFolder+"metrics_builtin_tech_generic_processCount.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28PROCESS_GROUP_INSTANCE%29&from=1631862000000&metricSelector=builtin%3Atech.generic.processCount%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_tech_generic_processCount_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("proc_count", 48.63491666452461, "entitySelector=type(PROCESS_GROUP_INSTANCE)&metricSelector=builtin:tech.generic.processCount:splitBy():avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP90(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p90/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_response_time_p90.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2890.000000%29%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_service_response_time_p90.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("svc_rt_p90", 35.00002454848894, "MV2;MicroSecond;entitySelector=type(SERVICE)&metricSelector=builtin:service.response.time:splitBy():percentile(90.000000):names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP50(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p50/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_response_time_p50.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2850.000000%29%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_service_response_time_p50.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("svc_rt_p50", 1.500151733421778, "MV2;MicroSecond;entitySelector=type(SERVICE)&metricSelector=builtin:service.response.time:splitBy():percentile(50.000000):names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessMemoryAvg(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_memory_avg/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_process_memory_avg.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.mem.workingSetSize", testDataFolder+"metrics_builtin_tech_generic_mem_workingSetSize.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28PROCESS_GROUP_INSTANCE%29&from=1631862000000&metricSelector=builtin%3Atech.generic.mem.workingSetSize%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_tech_generic_mem_workingsetsize_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("process_memory", 1437907.0484235594, "MV2;Byte;entitySelector=type(PROCESS_GROUP_INSTANCE)&metricSelector=builtin:tech.generic.mem.workingSetSize:splitBy():avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessCPUAvg(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_cpu_avg/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_process_cpu_avg.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:tech.generic.cpu.usage", testDataFolder+"metrics_builtin_tech_generic_cpu_usage.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28PROCESS_GROUP_INSTANCE%29&from=1631862000000&metricSelector=builtin%3Atech.generic.cpu.usage%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_tech_generic_cpu_usage_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("process_cpu", 14.223367878298156, "entitySelector=type(PROCESS_GROUP_INSTANCE)&metricSelector=builtin:tech.generic.cpu.usage:splitBy():avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_Throughput(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/throughput/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_throughput.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.requestCount.total", testDataFolder+"metrics_builtin_service_requestcount_total.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1631862000000&metricSelector=builtin%3Aservice.requestCount.total%3AsplitBy%28%29%3Avalue%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_service_requestcount_total_value.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("svc_tp_min", 68044716, "entitySelector=type(SERVICE)&metricSelector=builtin:service.requestCount.total:splitBy():value:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostCPUUsageAvg(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_cpu_usage_avg/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_host_cpu_usage_avg.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.cpu.usage", testDataFolder+"metrics_builtin_host_cpu_usage.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28HOST%29&from=1631862000000&metricSelector=builtin%3Ahost.cpu.usage%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_host_cpu_usage_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("host_cpu", 20.309976061722214, "entitySelector=type(HOST)&metricSelector=builtin:host.cpu.usage:splitBy():avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostMemoryUsageAvg(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_mem_usage_avg/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_host_mem_usage_avg.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.mem.usage", testDataFolder+"metrics_builtin_host_mem_usage.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28HOST%29&from=1631862000000&metricSelector=builtin%3Ahost.mem.usage%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_host_mem_usage_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("host_mem", 45.443796324058994, "entitySelector=type(HOST)&metricSelector=builtin:host.mem.usage:splitBy():avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostDiskQueueLengthMax(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_disk_queuelength_max/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_host_disk_queuelength_max.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:host.disk.queueLength", testDataFolder+"metrics_builtin_host_disk_queuelength.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28HOST%29&from=1631862000000&metricSelector=builtin%3Ahost.disk.queueLength%3AsplitBy%28%29%3Amax%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_host_disk_queuelength_max.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("host_disk_queue", 100, "entitySelector=type(HOST)&metricSelector=builtin:host.disk.queueLength:splitBy():max:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_NonDbChildCallCount(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/non_db_child_call_count/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_non_db_child_call_count.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.nonDbChildCallCount", testDataFolder+"metrics_builtin_service_nondbchildcallcount.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1631862000000&metricSelector=builtin%3Aservice.nonDbChildCallCount%3AsplitBy%28%29%3Avalue%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_query_builtin_service_nondbchildcallcount.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("svc2svc_calls", 13657068, "entitySelector=type(SERVICE)&metricSelector=builtin:service.nonDbChildCallCount:splitBy():value:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile tests an additional custom charting tile with exclude set to true is skipped.
// This results in success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/excluded_tile/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_excluded_tile.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_get_by_id_builtin_service_responsetime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_service_responsetime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("service_response_time", 29.31312208863131, "MV2;MicroSecond;entitySelector=type(SERVICE)&metricSelector=builtin:service.response.time:splitBy():avg:names"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func runGetSLIsFromDashboardTestAndCheckSLIs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *keptnv2.GetSLIFinishedEventData), sliResultsAssertionsFuncs ...func(t *testing.T, actual *keptnv2.SLIResult)) {
	eventSenderClient := &eventSenderClientMock{}
	rClient := &uploadErrorResourceClientMock{t: t}

	runTestAndAssertNoError(t, getSLIEventData, handler, eventSenderClient, rClient, testDashboardID)
	assertCorrectGetSLIEvents(t, eventSenderClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultsAssertionsFuncs...)
}

func runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *keptnv2.GetSLIFinishedEventData), uploadedSLOsAssertionsFunc func(t *testing.T, actual *keptnapi.ServiceLevelObjectives), sliResultsAssertionsFuncs ...func(t *testing.T, actual *keptnv2.SLIResult)) {
	eventSenderClient := &eventSenderClientMock{}
	rClient := &uploadErrorResourceClientMock{t: t}
	runTestAndAssertNoError(t, getSLIEventData, handler, eventSenderClient, rClient, testDashboardID)
	assertCorrectGetSLIEvents(t, eventSenderClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultsAssertionsFuncs...)
	uploadedSLOsAssertionsFunc(t, rClient.uploadedSLOs)
}

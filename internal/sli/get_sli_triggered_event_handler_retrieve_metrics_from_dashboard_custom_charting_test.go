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

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.keyRequest.totalProcessingTime",
		fullMetricSelector: "builtin:service.keyRequest.totalProcessingTime:splitBy(\"dt.entity.service_method\"):avg:names",
		entitySelector:     "type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\"))",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("processing_time_logupload", 2087.646963562753, expectedMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("processing_time_doallimportantwork10", 34999.99917777245, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIButNoSeries tests a custom charting tile with an SLI name defined, i.e. in the title, but no series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIButNoSeries(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_no_series_test/"

	handler := createHandlerForEarlyFailureCustomChartingTest(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("empty_chart"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries tests a custom charting tile with an SLI name defined and two series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_two_series_test/"

	handler := createHandlerForEarlyFailureCustomChartingTest(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("services_response_time_two_series"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy tests a custom charting tile with neither split by or filter by defined.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_no_filterby/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():avg:names",
		entitySelector:     "type(SERVICE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 54896.50469568841, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod tests a custom charting tile that splits by service key request and filters by service of service method.
// This is will result in a SLIResult with failure, as this is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_serviceofservicemethod/"

	handler := createHandlerForLateFailureCustomChartingTest(t, testDataFolder, "builtin:service.keyRequest.totalProcessingTime")
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("tpt_key_requests_journeyservice"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag tests a custom charting tile that splits by service and filters by tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_autotag/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:names",
		entitySelector:     "type(SERVICE),tag(\"keptn_managed\")",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags_bookingservice", 357668.85193320084, expectedMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags__", 645.8395061728395, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterBySpecificEntity tests a custom charting tile that splits by service and filters by specific entity.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterBySpecificEntity(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_specificentity/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:names",
		entitySelector:     "type(SERVICE),entityId(\"SERVICE-C6876D601CA5DDFD\")",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_specificentity", 57974.262650996854, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech tests a custom charting tile that filters by service software tech.
// This is will result in a SLIResult with failure, as the SERVICE_SOFTWARE_TECH filter is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_filterby_servicesoftwaretech/"

	handler := createHandlerForLateFailureCustomChartingTest(t, testDataFolder, "builtin:service.response.time")
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("svc_rt_p95"))
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_WorkerProcessCount(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/worker_process_count_avg/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:tech.generic.processCount",
		fullMetricSelector: "builtin:tech.generic.processCount:splitBy():avg:names",
		entitySelector:     "type(PROCESS_GROUP_INSTANCE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("proc_count", 48.89432551431819, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP90(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p90/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():percentile(90.000000):names",
		entitySelector:     "type(SERVICE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_rt_p90", 35007.25927374065, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP50(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p50/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():percentile(50.000000):names",
		entitySelector:     "type(SERVICE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_rt_p50", 1500.1086807956667, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessMemoryAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_memory_avg/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:tech.generic.mem.workingSetSize",
		fullMetricSelector: "builtin:tech.generic.mem.workingSetSize:splitBy():avg:names",
		entitySelector:     "type(PROCESS_GROUP_INSTANCE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("process_memory", 1480033899.1845968, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessCPUAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_cpu_avg/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:tech.generic.cpu.usage",
		fullMetricSelector: "builtin:tech.generic.cpu.usage:splitBy():avg:names",
		entitySelector:     "type(PROCESS_GROUP_INSTANCE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("process_cpu", 14.29287304299379, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_Throughput(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/throughput/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.requestCount.total",
		fullMetricSelector: "builtin:service.requestCount.total:splitBy():value:names",
		entitySelector:     "type(SERVICE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_tp_min", 2099456590, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostCPUUsageAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_cpu_usage_avg/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:host.cpu.usage",
		fullMetricSelector: "builtin:host.cpu.usage:splitBy():avg:names",
		entitySelector:     "type(HOST)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_cpu", 20.41917825744766, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostMemoryUsageAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_mem_usage_avg/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:host.mem.usage",
		fullMetricSelector: "builtin:host.mem.usage:splitBy():avg:names",
		entitySelector:     "type(HOST)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_mem", 45.433269610961815, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostDiskQueueLengthMax(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_disk_queuelength_max/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:host.disk.queueLength",
		fullMetricSelector: "builtin:host.disk.queueLength:splitBy():max:names",
		entitySelector:     "type(HOST)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_disk_queue", 100, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_NonDbChildCallCount(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/non_db_child_call_count/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.nonDbChildCallCount",
		fullMetricSelector: "builtin:service.nonDbChildCallCount:splitBy():value:names",
		entitySelector:     "type(SERVICE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc2svc_calls", 341746808, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile tests an additional custom charting tile with exclude set to true is skipped.
// This results in success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile(t *testing.T) {
	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     "./testdata/dashboards/custom_charting/excluded_tile/",
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():avg:names",
		entitySelector:     "type(SERVICE)",
	})

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 54896.50469568841, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_UnitTransformMilliseconds tests a custom charting tile with units set to milliseconds.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_UnitTransformMilliseconds(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/unit_transform_milliseconds/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():avg:names",
		entitySelector:     "type(SERVICE)",
	})

	handler.AddExact(buildMetricsUnitsConvertRequest("MicroSecond", 54896.48858596068, "MilliSecond"), filepath.Join(testDataFolder, "metrics_units_convert1.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 54.89648858596068, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_UnitTransformError tests a custom charting tile with invalid units generates the expected error.
func TestRetrieveMetricsFromDashboardCustomChartingTile_UnitTransformError(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/unit_transform_error/"

	requestBuilder := newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():avg:names").copyWithEntitySelector("type(SERVICE)")

	handler, _ := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: requestBuilder.metricSelector(),
		entitySelector:     "type(SERVICE)",
	})

	handler.AddExactError(buildMetricsUnitsConvertRequest("MicroSecond", 54896.48858596068, "Byte"), 400, filepath.Join(testDataFolder, "metrics_units_convert_error.json"))
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc("service_response_time", requestBuilder.build()))
}

type successfulCustomChartingTestHandlerConfiguration struct {
	testDataFolder     string
	baseMetricSelector string
	fullMetricSelector string
	entitySelector     string
}

func createHandlerForSuccessfulCustomChartingTest(t *testing.T, config successfulCustomChartingTestHandlerConfiguration) (*test.FileBasedURLHandler, string) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(config.testDataFolder, "dashboard.json"))
	queryBuilder := newMetricsV2QueryRequestBuilder(config.fullMetricSelector).copyWithEntitySelector(config.entitySelector)

	expectedFirstMetricsRequest := queryBuilder.build()
	expectedSecondMetricsRequest := queryBuilder.copyWithResolution(resolutionInf).build()

	handler.AddExact(buildMetricsV2DefinitionRequestString(config.baseMetricSelector), filepath.Join(config.testDataFolder, "metrics_get_by_id_base.json"))
	handler.AddExact(buildMetricsV2DefinitionRequestString(config.fullMetricSelector), filepath.Join(config.testDataFolder, "metrics_get_by_id_full.json"))
	handler.AddExact(expectedFirstMetricsRequest, filepath.Join(config.testDataFolder, "metrics_get_by_query_first.json"))
	handler.AddExact(expectedSecondMetricsRequest, filepath.Join(config.testDataFolder, "metrics_get_by_query_second.json"))

	return handler, expectedSecondMetricsRequest
}

func createHandlerForEarlyFailureCustomChartingTest(t *testing.T, testDataFolder string) *test.FileBasedURLHandler {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	return handler
}

func createHandlerForLateFailureCustomChartingTest(t *testing.T, testDataFolder string, baseMetricSelector string) *test.FileBasedURLHandler {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(buildMetricsV2DefinitionRequestString(baseMetricSelector), filepath.Join(testDataFolder, "metrics_get_by_id_base.json"))
	return handler
}

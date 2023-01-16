package sli

import (
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag tests splitting by key service request and filtering by tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_autotag/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.keyRequest.totalProcessingTime")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.keyRequest.totalProcessingTime:splitBy(\"dt.entity.service_method\"):avg:names").copyWithEntitySelector("type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\"))"))

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

	handler := createHandlerWithDashboard(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("empty_chart"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries tests a custom charting tile with an SLI name defined and two series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardCustomChartingTile_WithSLIAndTwoSeries(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_two_series_test/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("services_response_time_two_series"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy tests a custom charting tile with neither split by or filter by defined.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_no_filterby/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():avg:names").copyWithEntitySelector("type(SERVICE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 54896.50469568841, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod tests a custom charting tile that splits by service key request and filters by service of service method.
// This is will result in a SLIResult with failure, as this is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_serviceofservicemethod/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.keyRequest.totalProcessingTime")
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("tpt_key_requests_journeyservice"))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag tests a custom charting tile that splits by service and filters by tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_autotag/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:names").copyWithEntitySelector("type(SERVICE),tag(\"keptn_managed\")"))

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

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:names").copyWithEntitySelector("type(SERVICE),entityId(\"SERVICE-C6876D601CA5DDFD\")"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_specificentity", 57974.262650996854, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech tests a custom charting tile that filters by service software tech.
// This is will result in a SLIResult with failure, as the SERVICE_SOFTWARE_TECH filter is not supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_NoSplitByFilterByServiceSoftwareTech(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/no_splitby_filterby_servicesoftwaretech/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("svc_rt_p95"))
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_WorkerProcessCount(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/worker_process_count_avg/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:tech.generic.processCount")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:tech.generic.processCount:splitBy():avg:names").copyWithEntitySelector("type(PROCESS_GROUP_INSTANCE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("proc_count", 48.89432551431819, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP90(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p90/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():percentile(90.000000):names").copyWithEntitySelector("type(SERVICE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_rt_p90", 35007.25927374065, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ResponseTimeP50(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/response_time_p50/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():percentile(50.000000):names").copyWithEntitySelector("type(SERVICE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_rt_p50", 1500.1086807956667, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessMemoryAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_memory_avg/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:tech.generic.mem.workingSetSize")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:tech.generic.mem.workingSetSize:splitBy():avg:names").copyWithEntitySelector("type(PROCESS_GROUP_INSTANCE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("process_memory", 1480033899.1845968, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_ProcessCPUAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/process_cpu_avg/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:tech.generic.cpu.usage")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:tech.generic.cpu.usage:splitBy():avg:names").copyWithEntitySelector("type(PROCESS_GROUP_INSTANCE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("process_cpu", 14.29287304299379, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_Throughput(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/throughput/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.requestCount.total")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.requestCount.total:splitBy():value:names").copyWithEntitySelector("type(SERVICE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_tp_min", 2099456590, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostCPUUsageAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_cpu_usage_avg/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:host.cpu.usage")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:host.cpu.usage:splitBy():avg:names").copyWithEntitySelector("type(HOST)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_cpu", 20.41917825744766, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostMemoryUsageAvg(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_mem_usage_avg/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:host.mem.usage")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:host.mem.usage:splitBy():avg:names").copyWithEntitySelector("type(HOST)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_mem", 45.433269610961815, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_HostDiskQueueLengthMax(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/host_disk_queuelength_max/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:host.disk.queueLength")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:host.disk.queueLength:splitBy():max:names").copyWithEntitySelector("type(HOST)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("host_disk_queue", 100, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_OldTest_NonDbChildCallCount(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/old_tests/non_db_child_call_count/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.nonDbChildCallCount")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.nonDbChildCallCount:splitBy():value:names").copyWithEntitySelector("type(SERVICE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc2svc_calls", 341746808, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile tests an additional custom charting tile with exclude set to true is skipped.
// This results in success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_ExcludedTile(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/excluded_tile/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():avg:names").copyWithEntitySelector("type(SERVICE)"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 54896.50469568841, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_UnitTransformMilliseconds tests a custom charting tile with units set to milliseconds.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardCustomChartingTile_UnitTransformMilliseconds(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/unit_transform_milliseconds/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():avg:names").copyWithEntitySelector("type(SERVICE)"),
		createToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("service_response_time", 54.896485186544574, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_IncompatibleUnitTransformIsIgnored tests a custom charting tile that performs and incompatible unit transform performs no conversion and still succeeds due to the underlying API.
// This is would require a misconfigured Custom charting tile to occur.
func TestRetrieveMetricsFromDashboardCustomChartingTile_IncompatibleUnitTransformIsIgnored(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/unit_transform_error/"

	requestBuilder := newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():avg:names").copyWithEntitySelector("type(SERVICE)")

	handler := createHandlerWithDashboard(t, testDataFolder)
	addRequestToHandlerForBaseMetricDefinition(handler, testDataFolder, "builtin:service.response.time")

	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testDataFolder, requestBuilder, createToUnitConversionSnippet(microSecondUnitID, byteUnitID))

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("service_response_time", 54896.485186544574, expectedMetricsRequest))
}

// TestRetrieveMetricsFromDashboardCustomChartingTile_ManagementZonesWork tests applying management zones to the dashboard and tile work as expected.
func TestRetrieveMetricsFromDashboardCustomChartingTile_ManagementZonesWork(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/custom_charting/management_zones_work/"

	dashboardFilterWithManagementZone := dynatrace.DashboardFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "2311420533206603714",
			Name: "ap_mz_1",
		},
	}

	dashboardFilterWithAllManagementZones := dynatrace.DashboardFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "all",
			Name: "All",
		},
	}

	emptyTileFilter := dynatrace.TileFilter{}

	tileFilterWithManagementZone := dynatrace.TileFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "-6219736993013608218",
			Name: "ap_mz_2",
		},
	}

	tileFilterWithAllManagementZones := dynatrace.TileFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "all",
			Name: "All",
		},
	}

	requestBuilderWithNoManagementZone := newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():avg:names").copyWithEntitySelector("type(SERVICE)")
	requestBuilderWithManagementZone1 := requestBuilderWithNoManagementZone.copyWithMZSelector("mzName(\"ap_mz_1\")")
	requestBuilderWithManagementZone2 := requestBuilderWithNoManagementZone.copyWithMZSelector("mzName(\"ap_mz_2\")")

	tests := []struct {
		name             string
		dashboardFilter  *dynatrace.DashboardFilter
		tileFilter       dynatrace.TileFilter
		requestBuilder   *metricsV2QueryRequestBuilder
		expectedSLIValue float64
	}{
		{
			name:             "no_dashboard_filter_and_empty_tile_filter",
			dashboardFilter:  nil,
			tileFilter:       emptyTileFilter,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.48722844951,
		},
		{
			name:             "dashboard_filter_with_mz_and_empty_tile_filter",
			dashboardFilter:  &dashboardFilterWithManagementZone,
			tileFilter:       emptyTileFilter,
			requestBuilder:   requestBuilderWithManagementZone1,
			expectedSLIValue: 115445.40697872869,
		},
		{
			name:             "dashboard_filter_with_all_mz_and_empty_tile_filter",
			dashboardFilter:  &dashboardFilterWithAllManagementZones,
			tileFilter:       emptyTileFilter,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.48722844951,
		},
		{
			name:             "no_dashboard_filter_and_tile_filter_with_mz",
			dashboardFilter:  nil,
			tileFilter:       tileFilterWithManagementZone,
			requestBuilder:   requestBuilderWithManagementZone2,
			expectedSLIValue: 1519500.493859082,
		},
		{
			name:             "dashboard_filter_with_mz_and_tile_filter_with_mz",
			dashboardFilter:  &dashboardFilterWithManagementZone,
			tileFilter:       tileFilterWithManagementZone,
			requestBuilder:   requestBuilderWithManagementZone2,
			expectedSLIValue: 1519500.493859082,
		},
		{
			name:             "dashboard_filter_with_all_mz_and_tile_filter_with_mz",
			dashboardFilter:  &dashboardFilterWithAllManagementZones,
			tileFilter:       tileFilterWithManagementZone,
			requestBuilder:   requestBuilderWithManagementZone2,
			expectedSLIValue: 1519500.493859082,
		},
		{
			name:             "no_dashboard_filter_and_tile_filter_with_all_mz",
			dashboardFilter:  nil,
			tileFilter:       tileFilterWithAllManagementZones,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.48722844951,
		},
		{
			name:             "dashboard_filter_with_mz_and_tile_filter_with_all_mz",
			dashboardFilter:  &dashboardFilterWithManagementZone,
			tileFilter:       tileFilterWithAllManagementZones,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.48722844951,
		},
		{
			name:             "dashboard_filter_with_all_mz_and_tile_filter_with_all_mz",
			dashboardFilter:  &dashboardFilterWithAllManagementZones,
			tileFilter:       tileFilterWithAllManagementZones,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.48722844951,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				filepath.Join(testDataFolder, "dashboard.template.json"),
				struct {
					DashboardFilterString string
					TileFilterString      string
				}{
					DashboardFilterString: convertToJSONStringOrEmptyIfNil(t, tt.dashboardFilter),
					TileFilterString:      convertToJSONString(t, tt.tileFilter),
				},
			)

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)
			handler.AddExactFile(buildMetricsV2DefinitionRequestString("builtin:service.response.time"), filepath.Join(testVariantDataFolder, "metrics_get_by_id0.json"))
			metricsQueryRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, tt.requestBuilder)
			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("service_response_time", tt.expectedSLIValue, metricsQueryRequest))
		})
	}
}

package sli

import (
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

var getSLIFinishedEventSuccessAssertionsFunc = func(t *testing.T, data *keptnv2.GetSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultPass, data.Result)
	assert.Empty(t, data.Message)
}

var getSLIFinishedEventFailureAssertionsFunc = func(t *testing.T, data *keptnv2.GetSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
	assert.NotEmpty(t, data.Message)
}

var testCustomChartingGetSLIEventData = createTestGetSLIEventDataWithStartAndEnd("2021-09-17T07:00:00.000Z", "2021-09-17T08:00:00.000Z")

func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_autotag/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_servicekeyrequest_filterby_autotag.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.keyRequest.totalProcessingTime", testDataFolder+"metrics_get_by_id_builtin_servicekeyrequest_totalprocessingtime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE_METHOD%29%2CfromRelationships.isServiceMethodOfService%28type%28SERVICE%29%2Ctag%28%22keptnmanager%22%29%29&from=1631862000000&metricSelector=builtin%3Aservice.keyRequest.totalProcessingTime%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_servicekeyrequest_totalprocessingtime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("processing_time_getJourneyPageByTenant", 15.964052631578946),
		createSuccessfulSLIResultAssertionsFunc("processing_time_findLocations", 18.22756816390859),
		createSuccessfulSLIResultAssertionsFunc("processing_time_getJourneyById", 2.8606086572438163),
		createSuccessfulSLIResultAssertionsFunc("processing_time_findJourneys", 23.587584492453388),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "processing_time_getJourneyPageByTenant", "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-7E279028E3327C67)")
		assertSLIDefinitionIsPresent(t, actual, "processing_time_findLocations", "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-935D5E52D2E0C97E)")
		assertSLIDefinitionIsPresent(t, actual, "processing_time_getJourneyById", "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-3CE2B68F45050ED9)")
		assertSLIDefinitionIsPresent(t, actual, "processing_time_findJourneys", "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-00542666DD40A496)")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestDashboardCustomChartingTile_WithSLIButNoSeries(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/sli_name_no_series_test/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_sli_name_no_series.json")

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testCustomChartingGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("empty_chart"))
}

func TestDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByServiceOfServiceMethod(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_serviceofservicemethod/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_servicekeyrequest_filterby_serviceofservicemethod.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.keyRequest.totalProcessingTime", testDataFolder+"metrics_get_by_id_builtin_servicekeyrequest_totalprocessingtime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE_METHOD%29%2CfromRelationships.isServiceMethodOfService%28type%28SERVICE%29%2Ctag%28%22keptnmanager%22%29%29&from=1631862000000&metricSelector=builtin%3Aservice.keyRequest.totalProcessingTime%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_servicekeyrequest_totalprocessingtime.json")

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testCustomChartingGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("tpt_key_requests_journeyService"))
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterByAutoTag(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_autotag/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_service_filterby_autotag.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_get_by_id_builtin_service_responsetime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28%22keptn_managed%22%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_service_responsetime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags_JourneyService", 20.256493055555555),
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_autotags_EasytravelService", 132.27823461853978),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "services_response_time_splitby_service_filterby_autotags_JourneyService", "MV2;MicroSecond;metricSelector=builtin:service.response.time:avg:names&entitySelector=type(SERVICE),tag(\"keptn_managed\"),entityId(SERVICE-F2455557EF67362B)")
		assertSLIDefinitionIsPresent(t, actual, "services_response_time_splitby_service_filterby_autotags_EasytravelService", "MV2;MicroSecond;metricSelector=builtin:service.response.time:avg:names&entitySelector=type(SERVICE),tag(\"keptn_managed\"),entityId(SERVICE-B67B3EC4C95E0FA7)")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceFilterBySpecificEntity(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_service_filterby_specificentity/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_service_filterby_specificentity.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_get_by_id_builtin_service_responsetime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2CentityId%28%22SERVICE-F2455557EF67362B%22%29&from=1631862000000&metricSelector=builtin%3Aservice.response.time%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_service_responsetime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("services_response_time_splitby_service_filterby_specificentity", 20.256493055555555),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "services_response_time_splitby_service_filterby_specificentity", "MV2;MicroSecond;metricSelector=builtin:service.response.time:avg:names&entitySelector=type(SERVICE),entityId(\"SERVICE-F2455557EF67362B\")")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testCustomChartingGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

func runGetSLIsFromDashboardTestAndCheckSLIs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *keptnv2.GetSLIFinishedEventData), uploadedSLIsAssertionsFunc func(t *testing.T, actual *dynatrace.SLI), sliResultsAssertionsFuncs ...func(t *testing.T, actual *keptnv2.SLIResult)) {
	kClient := &keptnClientMock{}
	rClient := &uploadErrorResourceClientMock{t: t}

	runTestAndAssertNoError(t, getSLIEventData, handler, kClient, rClient, testDashboardID)
	assertCorrectGetSLIEvents(t, kClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultsAssertionsFuncs...)
	uploadedSLIsAssertionsFunc(t, rClient.uploadedSLIs)
}

func assertSLIDefinitionIsPresent(t *testing.T, slis *dynatrace.SLI, metric string, definition string) {
	if !assert.NotNil(t, slis) {
		return
	}
	assert.Contains(t, slis.Indicators, metric)
	assert.EqualValues(t, definition, slis.Indicators[metric])
}

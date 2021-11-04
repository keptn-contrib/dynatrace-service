package sli

import (
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

type expectedDashboardSLIResult struct {
	name       string
	value      float64
	definition string
}

func TestRetrieveMetricsFromDashboardCustomChartingTile_SplitByServiceKeyRequestFilterByAutoTag(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/custom_charting/splitby_servicekeyrequest_filterby_autotag/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_custom_charting_splitby_servicekeyrequest_filterby_autotag.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.keyRequest.totalProcessingTime", testDataFolder+"metrics_get_by_id_builtin_servicekeyrequest_totalprocessingtime.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE_METHOD%29%2CfromRelationships.isServiceMethodOfService%28type%28SERVICE%29%2Ctag%28%22keptnmanager%22%29%29&from=1631862000000&metricSelector=builtin%3Aservice.keyRequest.totalProcessingTime%3Aavg%3Anames&resolution=Inf&to=1631865600000",
		testDataFolder+"metrics_get_by_query_builtin_servicekeyrequest_totalprocessingtime.json")

	getSLIEventData := makeTestGetSLIEventDataWithStartAndEnd("2021-09-17T07:00:00.000Z", "2021-09-17T08:00:00.000Z")

	expectedDashboardSLIResults := []expectedDashboardSLIResult{
		{
			name:       "processing_time_getJourneyPageByTenant",
			definition: "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-7E279028E3327C67)",
			value:      15.964052631578946,
		},
		{
			name:       "processing_time_findLocations",
			definition: "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-935D5E52D2E0C97E)",
			value:      18.22756816390859,
		},
		{
			name:       "processing_time_getJourneyById",
			definition: "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-3CE2B68F45050ED9)",
			value:      2.8606086572438163,
		},
		{
			name:       "processing_time_findJourneys",
			definition: "MV2;MicroSecond;metricSelector=builtin:service.keyRequest.totalProcessingTime:avg:names&entitySelector=type(SERVICE_METHOD),fromRelationships.isServiceMethodOfService(type(SERVICE),tag(\"keptnmanager\")),entityId(SERVICE_METHOD-00542666DD40A496)",
			value:      23.587584492453388,
		},
	}

	runGetSLIsFromDashboardTestAndExpectSuccess(t, handler, getSLIEventData, expectedDashboardSLIResults)
}

func runGetSLIsFromDashboardTestAndExpectSuccess(t *testing.T, handler http.Handler, ev *getSLIEventData, expectedResults []expectedDashboardSLIResult) {
	kClient := &keptnClientMock{}
	rClient := &uploadErrorResourceClientMock{t: t}

	setupTestAndAssertNoError(t, ev, handler, kClient, rClient, testDashboardID)

	eventAssertionsFunc := func(data *keptnv2.GetSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultPass, data.Result)
		assert.Empty(t, data.Message)
	}

	getSLIFinishedEventData := assertThatEventsAreThere(t, kClient.eventSink, eventAssertionsFunc)
	verifySLIResultsSucceeded(t, expectedResults, getSLIFinishedEventData.GetSLI.IndicatorValues)
	verifyUploadedSLIDefinitions(t, expectedResults, rClient.uploadedSLIs)
}

func verifySLIResultsSucceeded(t *testing.T, expectedResults []expectedDashboardSLIResult, sliResults []*keptnv2.SLIResult) {
	if !assert.EqualValues(t, len(expectedResults), len(sliResults), "number of assertions should match number of SLI indicator values") {
		return
	}

	for i, expectedResult := range expectedResults {
		verifySLIResultSucceeded(t, expectedResult, sliResults[i])
	}
}

func verifySLIResultSucceeded(t *testing.T, expectedResult expectedDashboardSLIResult, sliResult *keptnv2.SLIResult) {
	assert.True(t, sliResult.Success)
	assert.EqualValues(t, expectedResult.name, sliResult.Metric)
	assert.EqualValues(t, expectedResult.value, sliResult.Value)
	assert.Empty(t, sliResult.Message)
}

func verifyUploadedSLIDefinitions(t *testing.T, expectedResults []expectedDashboardSLIResult, uploadedSLIs *dynatrace.SLI) {
	if !assert.NotNil(t, uploadedSLIs) {
		return
	}
	assert.EqualValues(t, len(expectedResults), len(uploadedSLIs.Indicators))
	for _, expectedResult := range expectedResults {
		definition, hasIndicator := uploadedSLIs.Indicators[expectedResult.name]
		assert.True(t, hasIndicator)
		assert.EqualValues(t, expectedResult.definition, definition)
	}
}

func makeTestGetSLIEventDataWithStartAndEnd(sliStart string, sliEnd string) *getSLIEventData {
	return &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
		sliStart:   sliStart,
		sliEnd:     sliEnd,
	}
}

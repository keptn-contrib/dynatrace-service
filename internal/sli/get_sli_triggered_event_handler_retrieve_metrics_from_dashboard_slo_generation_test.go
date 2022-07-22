package sli

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

// TestRetrieveMetrics_SLOObjectiveGeneratedFromSupportedDataExplorerTile tests that an SLO objective is created for a supported data explorer tile.
func TestRetrieveMetrics_SLOObjectiveGeneratedFromSupportedDataExplorerTile(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/slo_generation/supported_data_explorer_tile/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000",
		testDataFolder+"metrics_query_builtin_service_response_time.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("srt", 29.192929640271974, "MV2;MicroSecond;metricSelector=builtin:service.response.time:splitBy():avg:names"),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "srt",
			DisplayName: "srt",
			Pass:        []*keptnapi.SLOCriteria{{Criteria: []string{"<28"}}},
			Weight:      1,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testDataExplorerGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetrics_SLOObjectiveNotGeneratedFromUnsupportedDataExplorerTile tests that an SLO objective is also created for an unsupported data explorer tile.
func TestRetrieveMetrics_SLOObjectiveNotGeneratedFromUnsupportedDataExplorerTile(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/slo_generation/unsupported_data_explorer_tile/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("response_time"),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "response_time",
			DisplayName: "response_time",
			Pass:        []*keptnapi.SLOCriteria{{Criteria: []string{"<1200"}}},
			Weight:      1,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testDataExplorerGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetrics_SLOObjectiveGeneratedForNoDataFromDataExplorerTile tests that an SLO objective is created for a data explorer tile which results in a metrics query that returns no data.
func TestRetrieveMetrics_SLOObjectiveGeneratedForNoDataFromDataExplorerTile(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/slo_generation/data_explorer_tile_no_data/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=entityId%28SERVICE-C33B8A4C73748469%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000",
		testDataFolder+"metrics_query_builtin_service_response_time.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("srt_service", "entitySelector=entityId(SERVICE-C33B8A4C73748469)&metricSelector=builtin:service.response.time:splitBy():avg:names"),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "srt_service",
			DisplayName: "srt_service",
			Pass:        []*keptnapi.SLOCriteria{{Criteria: []string{"<10"}}},
			Weight:      1,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testDataExplorerGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

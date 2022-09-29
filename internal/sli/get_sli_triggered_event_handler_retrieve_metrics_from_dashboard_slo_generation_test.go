package sli

import (
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
)

// TestRetrieveMetrics_SLOObjectiveGeneratedFromSupportedDataExplorerTile tests that an SLO objective is created for a supported data explorer tile.
func TestRetrieveMetrics_SLOObjectiveGeneratedFromSupportedDataExplorerTile(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_generation/supported_data_explorer_tile/"

	expectedMetricsRequest := buildMetricsV2RequestString("%28builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_builtin_service_response_time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_service_response_time.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt", 54896.44626158664, expectedMetricsRequest),
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

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetrics_SLOObjectiveNotGeneratedFromUnsupportedDataExplorerTile tests that an SLO objective is also created for an unsupported data explorer tile.
func TestRetrieveMetrics_SLOObjectiveNotGeneratedFromUnsupportedDataExplorerTile(t *testing.T) {
	// TODO: 25-08-2022: Check if this test is still needed
	t.Skip("Investigate if this test is still needed")
	const testDataFolder = "./testdata/dashboards/slo_generation/unsupported_data_explorer_tile/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_builtin_service_response_time.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
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

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetrics_SLOObjectiveGeneratedForNoDataFromDataExplorerTile tests that an SLO objective is created for a data explorer tile which results in a metrics query that returns no data.
func TestRetrieveMetrics_SLOObjectiveGeneratedForNoDataFromDataExplorerTile(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_generation/data_explorer_tile_no_data/"

	expectedMetricsRequest := buildMetricsV2RequestString("%28builtin%3Aservice.response.time%3Afilter%28and%28or%28in%28%22dt.entity.service%22%2CentitySelector%28%22type%28service%29%2CentityId%28~%22SERVICE-C33B8A4C73748469~%22%29%22%29%29%29%29%29%3AsplitBy%28%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metrics_builtin_service_response_time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_builtin_service_response_time.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("srt_service", expectedMetricsRequest),
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

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

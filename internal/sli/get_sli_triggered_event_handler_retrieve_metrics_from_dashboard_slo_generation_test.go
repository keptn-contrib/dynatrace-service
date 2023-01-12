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

	handler := createHandlerWithDashboard(t, testDataFolder)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt", 54896.50447404383, expectedMetricsRequest),
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
	const testDataFolder = "./testdata/dashboards/slo_generation/unsupported_data_explorer_tile/"

	// TODO: 25-08-2022: Check if this test is still needed
	t.Skip("Investigate if this test is still needed")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))

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

	requestBuilder := newMetricsV2QueryRequestBuilder("(builtin:service.response.time:filter(and(or(in(\"dt.entity.service\",entitySelector(\"type(service),entityId(~\"SERVICE-C33B8A4C73748469~\")\"))))):splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names")

	handler := createHandlerWithDashboard(t, testDataFolder)
	_ = addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testDataFolder, requestBuilder)

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

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, uploadedSLOsAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc("srt_service", requestBuilder.build()))
}

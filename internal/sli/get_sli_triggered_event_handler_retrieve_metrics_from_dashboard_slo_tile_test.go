package sli

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
)

var testSLOTileGetSLIEventData = createTestGetSLIEventDataWithStartAndEnd("2021-09-17T07:00:00.000Z", "2021-09-17T08:00:00.000Z")

func TestRetrieveMetricsFromDashboardSLOTile_SLOFound(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/passing_slo/"

	const expectedSLORequest = dynatrace.SLOPath + "/7d07efde-b714-3e6e-ad95-08490e2540c4?from=1631862000000&timeFrame=GTF&to=1631865600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(expectedSLORequest, testDataFolder+"slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("static_slo_-_pass", 95, expectedSLORequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:     "static_slo_-_pass",
			Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{">=90.000000"}}},
			Warning: []*keptnapi.SLOCriteria{{Criteria: []string{">=75.000000"}}},
			Weight:  1,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testSLOTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardSLOTile_TileWithNoIDs tests that an unsuccessful tile result is produced for SLO tiles reference no SLOs.
func TestRetrieveMetricsFromDashboardSLOTile_TileWithNoIDs(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/tile_no_slo_ids/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultAssertionsFunc("slo_tile_without_slo"),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		assert.Nil(t, actual)
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testSLOTileGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardSLOTile_TileWithEmptyID tests that an unsuccessful tile result is produced for SLO tiles containing an empty SLO ID.
func TestRetrieveMetricsFromDashboardSLOTile_TileWithEmptyID(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/tile_empty_slo_id/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultAssertionsFunc("slo_without_id"),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		assert.Nil(t, actual)
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testSLOTileGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardSLOTile_TileWithUnknownID tests that an unsuccessful tile result is produced for SLO tiles containing an unknown SLO ID.
func TestRetrieveMetricsFromDashboardSLOTile_TileWithUnknownID(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/unknown_slo_id/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.SLOPath+"/7d07efde-b714-3e6e-ad95-08490e2540c5?from=1631862000000&timeFrame=GTF&to=1631865600000", testDataFolder+"slo_7d07efde-b714-3e6e-ad95-08490e2540c5.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultAssertionsFunc("slo_7d07efde-b714-3e6e-ad95-08490e2540c5"),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		assert.Nil(t, actual)
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testSLOTileGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

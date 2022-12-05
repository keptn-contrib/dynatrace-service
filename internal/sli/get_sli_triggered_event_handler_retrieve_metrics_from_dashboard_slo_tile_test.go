package sli

import (
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
)

func TestRetrieveMetricsFromDashboardSLOTile_SLOFound(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/passing_slo/"

	expectedSLORequest := buildSLORequest("7d07efde-b714-3e6e-ad95-08490e2540c4")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedSLORequest, filepath.Join(testDataFolder, "slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc(testIndicatorStaticSLOPass, 95, expectedSLORequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:     testIndicatorStaticSLOPass,
			Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{">=90.000000"}}},
			Warning: []*keptnapi.SLOCriteria{{Criteria: []string{">=75.000000"}}},
			Weight:  1,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardSLOTile_TileWithNoIDs tests that an unsuccessful tile result is produced for SLO tiles reference no SLOs.
func TestRetrieveMetricsFromDashboardSLOTile_TileWithNoIDs(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/tile_no_slo_ids/"

	const sliName = "slo_tile_without_slo"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createSLOsAssertionsFuncForSingleInformationalSLO(sliName), createFailedSLIResultAssertionsFunc(sliName))
}

// TestRetrieveMetricsFromDashboardSLOTile_TileWithEmptyID tests that an unsuccessful tile result is produced for SLO tiles containing an empty SLO ID.
func TestRetrieveMetricsFromDashboardSLOTile_TileWithEmptyID(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/tile_empty_slo_id/"

	const sliName = "slo_without_id"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createSLOsAssertionsFuncForSingleInformationalSLO(sliName), createFailedSLIResultAssertionsFunc(sliName))
}

// TestRetrieveMetricsFromDashboardSLOTile_TileWithUnknownID tests that an unsuccessful tile result is produced for SLO tiles containing an unknown SLO ID.
func TestRetrieveMetricsFromDashboardSLOTile_TileWithUnknownID(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/slo_tiles/unknown_slo_id/"

	const sloID = "7d07efde-b714-3e6e-ad95-08490e2540c5"
	const sliName = "slo_" + sloID

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(buildSLORequest(sloID), filepath.Join(testDataFolder, "slo_7d07efde-b714-3e6e-ad95-08490e2540c5.json"))

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createSLOsAssertionsFuncForSingleInformationalSLO(sliName), createFailedSLIResultAssertionsFunc(sliName))
}

func createSLOsAssertionsFuncForSingleInformationalSLO(name string) func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
	return func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    name,
			Weight: 1,
		}, actual.Objectives[0])
	}
}

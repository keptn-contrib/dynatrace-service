package sli

import (
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

var testProblemTileGetSLIEventData = createTestGetSLIEventDataWithStartAndEnd("2021-09-17T07:00:00.000Z", "2021-09-17T08:00:00.000Z")

// TestRetrieveMetricsFromDashboardProblemTile_Success tests the success case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
func TestRetrieveMetricsFromDashboardProblemTile_Success(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/problem_tile_success/"

	const expectedProblemsRequest = dynatrace.ProblemsV2Path + "?from=1631862000000&problemSelector=status%28%22open%22%29&to=1631865600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(expectedProblemsRequest, testDataFolder+"problems_status_open.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("problems", 0, expectedProblemsRequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testProblemTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardProblemTile_CustomManagementZone tests retrieving the problem and security problem count SLIs in response to a problems dashboard tile with a custom management zone.
func TestRetrieveMetricsFromDashboardProblemTile_CustomManagementZone(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/custom_management_zone/"

	const expectedProblemsRequest = dynatrace.ProblemsV2Path + "?from=1631862000000&problemSelector=status%28%22open%22%29%2CmanagementZoneIds%289130632296508575249%29&to=1631865600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(expectedProblemsRequest, testDataFolder+"problems_status_open.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulDashboardSLIResultAssertionsFunc("problems", 10, expectedProblemsRequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testProblemTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardProblemTile_MissingScopes tests the failure case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
// Retrieving SLIs fails because the the API token is missing the required scopes.
func TestRetrieveMetricsFromDashboardProblemTile_MissingScopes(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/missing_scopes/"

	const expectedProblemsRequest = dynatrace.ProblemsV2Path + "?from=1631862000000&problemSelector=status%28%22open%22%29&to=1631865600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExactError(expectedProblemsRequest, 403, testDataFolder+"problems_missing_scope.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("problems", dynatrace.ProblemsV2Path+"?from=1631862000000&problemSelector=status%28%22open%22%29&to=1631865600000"),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testProblemTileGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

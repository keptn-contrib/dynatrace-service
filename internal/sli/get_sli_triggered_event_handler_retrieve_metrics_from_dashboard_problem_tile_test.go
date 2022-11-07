package sli

import (
	"path/filepath"
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// TestRetrieveMetricsFromDashboardProblemTile_Success tests the success case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
func TestRetrieveMetricsFromDashboardProblemTile_Success(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/problem_tile_success/"

	expectedProblemsRequest := buildProblemsV2Request("status(\"open\")")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedProblemsRequest, filepath.Join(testDataFolder, "problems_status_open.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("problems", 42, expectedProblemsRequest),
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

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardProblemTile_CustomManagementZone tests retrieving the problem and security problem count SLIs in response to a problems dashboard tile with a custom management zone.
func TestRetrieveMetricsFromDashboardProblemTile_CustomManagementZone(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/custom_management_zone/"

	expectedProblemsRequest := buildProblemsV2Request("status(\"open\"),managementZones(\"Easytravel\")")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedProblemsRequest, filepath.Join(testDataFolder, "problems_status_open.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("problems", 22, expectedProblemsRequest),
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

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardProblemTile_MissingScopes tests the failure case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
// Retrieving SLIs fails because the the API token is missing the required scopes.
func TestRetrieveMetricsFromDashboardProblemTile_MissingScopes(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/missing_scopes/"

	expectedProblemsRequest := buildProblemsV2Request("status(\"open\")")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExactError(expectedProblemsRequest, 403, filepath.Join(testDataFolder, "problems_missing_scope.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("problems", expectedProblemsRequest),
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

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

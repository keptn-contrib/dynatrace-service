package sli

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

var testProblemTileGetSLIEventData = createTestGetSLIEventDataWithStartAndEnd("2021-09-17T07:00:00.000Z", "2021-09-17T08:00:00.000Z")

// TestRetrieveMetricsFromDashboardProblemTile_Success tests the success case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
func TestRetrieveMetricsFromDashboardProblemTile_Success(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/problem_tile/problem_tile_success/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.ProblemsV2Path+"?from=1631862000000&problemSelector=status%28%22open%22%29&to=1631865600000", testDataFolder+"problems_status_open.json")
	handler.AddExact(dynatrace.SecurityProblemsPath+"?from=1631862000000&securityProblemSelector=status%28%22open%22%29&to=1631865600000", testDataFolder+"security_problems_status_open.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("problems", 0),
		createSuccessfulSLIResultAssertionsFunc("security_problems", 103),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "problems", "PV2;problemSelector=status(\"open\")")
		assertSLIDefinitionIsPresent(t, actual, "security_problems", "SECPV2;securityProblemSelector=status(\"open\")")
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 2, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[0])

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "security_problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[1])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testProblemTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardProblemTile_MissingScopes tests the failure case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
// Retrieving SLIs fails because the the API token is missing the required scopes.
func TestRetrieveMetricsFromDashboardProblemTile_MissingScopes(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/problem_tile/missing_scopes/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExactError(dynatrace.ProblemsV2Path+"?from=1631862000000&problemSelector=status%28%22open%22%29&to=1631865600000", 403, testDataFolder+"problems_missing_scope.json")
	handler.AddExactError(dynatrace.SecurityProblemsPath+"?from=1631862000000&securityProblemSelector=status%28%22open%22%29&to=1631865600000", 403, testDataFolder+"security_problems_missing_scope.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("problems"),
		createFailedSLIResultAssertionsFunc("security_problems"),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "problems", "PV2;problemSelector=status(\"open\")")
		assertSLIDefinitionIsPresent(t, actual, "security_problems", "SECPV2;securityProblemSelector=status(\"open\")")
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 2, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[0])

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "security_problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[1])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testProblemTileGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLIsAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

package sli

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// TestNoErrorIsReturnedWhenSLOFileWritingSucceeds tests that no error is returned if retrieving (a single) SLI from a dashboard works and the resulting SLO file is uploaded.
//
// prerequisites:
//   - we use a valid dashboard ID
//   - all processing and SLI result retrieval works
func TestNoErrorIsReturnedWhenSLOFileWritingSucceeds(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/success/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2895.000000%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_single_sli.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metric_definition_service-response-time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_1_result.json"))

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultPass, actual.Result)
		assert.Empty(t, actual.Message)
	}
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 210597.59603729026, expectedMetricsRequest))
}

// TestErrorIsReturnedWhenSLOFileWritingFails tests that an error is returned if retrieving (a single) SLI from a dashboard works but the upload of the SLO file fails.
//
// prerequisites:
//   - we use a valid dashboard ID
//   - all processing and SLI result retrieval works
//   - if an upload of the SLO file fails, then the test must fail
func TestErrorIsReturnedWhenSLOFileWritingFails(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/slo_writing_fails/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2895.000000%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_single_sli.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metric_definition_service-response-time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_1_result.json"))

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
		assert.Contains(t, actual.Message, "upload failed")
	}

	runGetSLIsFromDashboardTestWithConfigClientAndCheckSLIs(t, handler, testGetSLIEventData, newConfigClientMockThatErrorsUploadSLOs(t, errors.New("SLO upload failed")), getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95))
}

// TestThatThereIsNoFallbackToSLIsFromDashboard tests that retrieving a dashboard by ID works, and we ignore the outdated parse behaviour.
//
// prerequisites:
//   - we use a valid dashboard ID and it is returned by Dynatrace API
//   - The dashboard has 'KQG.QueryBehavior=ParseOnChange' set to only reparse the dashboard if it changed  (we do no longer consider this behaviour)
//   - we will not fallback to processing SLI files, but process the dashboard again
func TestThatThereIsNoFallbackToSLIsFromDashboard(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/no_fallback_to_slis/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2895.000000%29%3Anames")

	// we need metrics definition, because we will be retrieving metrics from dashboard
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_single_sli_parse_only_on_change.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metric_definition_service-response-time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_1_result.json"))

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		assert.NotNil(t, actual)
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 210597.59603729026, expectedMetricsRequest))
}

// TestDashboardThatProducesNoDataProducesError tests retrieving (a single) SLI from a dashboard that returns no data.
//
// prerequisites:
//   - we use a valid dashboard ID
//   - all processing works, but SLI result retrieval failed with 0 results (no data available)
func TestDashboardThatProducesNoDataProducesError(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/no_data/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2895.000000%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_single_sli.json"))
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", filepath.Join(testDataFolder, "metric_definition_service-response-time.json"))
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_0_results.json"))

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultWarning, actual.Result)
		assert.Contains(t, actual.Message, "Metrics API v2 returned zero metric series")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest))
}

// TestDashboardThatProducesNoResultsProducesError tests that processing a dashboard which produce no results produces an error.
//
// prerequisites:
//   - we use a valid dashboard ID and it is returned by Dynatrace API
//   - the dashboard does have a CustomCharting tile, but not the correct tile name, that would qualify it as SLI/SLO source
func TestDashboardThatProducesNoResultsProducesError(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/no_results/"

	// we do not need metrics definition and metrics query, because we will should not be looking into the tile
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard_custom_charting_without_matching_tile_name.json"))

	// no SLOs should be uploaded
	configClient := &uploadSLOsWillFailConfigClientMock{t: t}

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
		assert.Contains(t, actual.Message, "any SLI results")
	}

	runGetSLIsFromDashboardTestWithConfigClientAndCheckSLIs(t, handler, testGetSLIEventData, configClient, getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95))
}

// TestQueryDynatraceDashboardForSLIs tests that querying for a dashboard (i.e. dashboard=query) works as expected.
func TestQueryDynatraceDashboardForSLIs(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/dashboard_query/"

	expectedSLORequest := buildSLORequest("7d07efde-b714-3e6e-ad95-08490e2540c4")
	expectedProblemsV2Request := buildProblemsV2Request("status%28%22open%22%29%2CmanagementZoneIds%287030365576649815430%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath, filepath.Join(testDataFolder, "dashboards_query.json"))
	handler.AddExact(dynatrace.DashboardsPath+"/12345678-1111-4444-8888-123456789012", filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedSLORequest, filepath.Join(testDataFolder, "slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json"))
	handler.AddExact(expectedProblemsV2Request, filepath.Join(testDataFolder, "problems.json"))

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		assert.Equal(t, 2, len(actual.Objectives))
		assert.EqualValues(t, &keptnapi.SLOScore{Pass: "90%", Warning: "70%"}, actual.TotalScore)
		assert.EqualValues(
			t,
			&keptnapi.SLOComparison{
				CompareWith:               "single_result",
				IncludeResultWithScore:    "pass",
				NumberOfComparisonResults: 1,
				AggregateFunction:         "avg",
			},
			actual.Comparison)
	}

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_faster_500ms", 95, expectedSLORequest),
		createSuccessfulSLIResultAssertionsFunc("problems", 1, expectedProblemsV2Request),
	}

	runGetSLIsFromDashboardTestWithDashboardParameterAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, common.DynatraceConfigDashboardQUERY, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveDashboardWithUnknownButValidID tests requesting a dashboard with a valid but unknown ID fails as expected.
// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// also the ID of the dashboard we try to retrieve was not found
func TestRetrieveDashboardWithUnknownButValidID(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/valid_but_unknown_id/"

	const dashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9c"

	// we add a handler to simulate a very concrete 404 Dashboards API request/response in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath+"/"+dashboardID, http.StatusNotFound, filepath.Join(testDataFolder, "dashboards_query.json"))

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
		assert.Contains(t, actual.Message, dashboardID)
		assert.Contains(t, actual.Message, "not found")
	}

	runGetSLIsFromDashboardTestWithDashboardParameterAndCheckSLIs(t, handler, testGetSLIEventData, dashboardID, getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc("no metric"))
}

// TestRetrieveDashboardWithInvalidID tests that requesting a dashboard with an invalid ID fails as expected.
// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// it could happen that you have a copy/paste error in your ID (invalid UUID) so we will see a different error
func TestRetrieveDashboardWithInvalidID(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/invalid_id/"

	const dashboardID = "definitely-invalid-uuid"

	// we add a handler to simulate a very concrete 400 Dashboards API request/response in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath+"/"+dashboardID, http.StatusBadRequest, filepath.Join(testDataFolder, "dashboards_query.json"))

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
		assert.Contains(t, actual.Message, "Not a valid UUID")
	}

	runGetSLIsFromDashboardTestWithDashboardParameterAndCheckSLIs(t, handler, testGetSLIEventData, dashboardID, getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc("no metric"))
}

type uploadSLOsWillFailConfigClientMock struct {
	t *testing.T
}

func (m *uploadSLOsWillFailConfigClientMock) GetSLIs(_ context.Context, _ string, _ string, _ string) (map[string]string, error) {
	m.t.Fatalf("GetSLIs() should not be needed in this mock!")
	return nil, nil
}

func (m *uploadSLOsWillFailConfigClientMock) GetSLOs(_ context.Context, _ string, _ string, _ string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")
	return nil, nil
}

func (m *uploadSLOsWillFailConfigClientMock) UploadSLOs(_ context.Context, _ string, _ string, _ string, _ *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLOs() should not be needed in this mock!")
	return nil
}

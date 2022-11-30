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

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():percentile(95.000000):names",
		entitySelector:     "type(SERVICE)",
	},
	)

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 210598.1424830455, expectedMetricsRequest))
}

// TestErrorIsReturnedWhenSLOFileWritingFails tests that an error is returned if retrieving (a single) SLI from a dashboard works but the upload of the SLO file fails.
//
// prerequisites:
//   - we use a valid dashboard ID
//   - all processing and SLI result retrieval works
//   - if an upload of the SLO file fails, then the test must fail
func TestErrorIsReturnedWhenSLOFileWritingFails(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/slo_writing_fails/"

	handler, _ := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():percentile(95.000000):names",
		entitySelector:     "type(SERVICE)",
	},
	)

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
		assert.Contains(t, actual.Message, "upload failed")
	}

	runGetSLIsFromDashboardTestWithConfigClientAndCheckSLIs(t, handler, testGetSLIEventData, newConfigClientMockThatErrorsUploadSLOs(t, errors.New("SLO upload failed")), getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorNoMetric))
}

// TestThatThereIsNoFallbackToSLIsFromDashboard tests that retrieving a dashboard by ID works, and we ignore the outdated parse behaviour.
//
// prerequisites:
//   - we use a valid dashboard ID and it is returned by Dynatrace API
//   - The dashboard has 'KQG.QueryBehavior=ParseOnChange' set to only reparse the dashboard if it changed  (we do no longer consider this behaviour)
//   - we will not fallback to processing SLI files, but process the dashboard again
func TestThatThereIsNoFallbackToSLIsFromDashboard(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/no_fallback_to_slis/"

	handler, expectedMetricsRequest := createHandlerForSuccessfulCustomChartingTest(t, successfulCustomChartingTestHandlerConfiguration{
		testDataFolder:     testDataFolder,
		baseMetricSelector: "builtin:service.response.time",
		fullMetricSelector: "builtin:service.response.time:splitBy():percentile(95.000000):names",
		entitySelector:     "type(SERVICE)",
	},
	)

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		assert.NotNil(t, actual)
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 210598.14198018494, expectedMetricsRequest))
}

// TestDashboardThatProducesNoDataProducesError tests retrieving (a single) SLI from a dashboard that returns no data.
//
// prerequisites:
//   - we use a valid dashboard ID
//   - all processing works, but SLI result retrieval failed with 0 results (no data available)
func TestDashboardThatProducesNoDataProducesError(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/no_data/"

	expectedMetricsRequest := newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():percentile(95.000000):names").copyWithEntitySelector("type(SERVICE),entityId(\"SERVICE-F6B97183A8968C3A\")").build()

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(buildMetricsV2DefinitionRequestString("builtin:service.response.time"), filepath.Join(testDataFolder, "metric_definition_service-response-time.json"))
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
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))

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
	expectedProblemsV2Request := buildProblemsV2Request("status(\"open\"),managementZones(\"Keptn: keptn07project\")")

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
		createSuccessfulSLIResultAssertionsFunc("static_slo_-_pass", 95, expectedSLORequest),
		createSuccessfulSLIResultAssertionsFunc("problems", 0, expectedProblemsV2Request),
	}

	runGetSLIsFromDashboardTestWithDashboardParameterAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, "query", getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
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

// TestDashboardWithInformationalSLOWithNoData demonstrates that an informational SLO with no data  tests retrieving (a single) SLI from a dashboard that returns no data.

func TestDashboardWithInformationalSLOWithNoData(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/no_data_informational_sli/"

	handler := createHandlerForEarlyFailureDataExplorerTest(t, testDataFolder)
	expectedMetricsRequest := newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names").build()
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query1.json"))
	expectedSLORequest := buildSLORequest("7d07efde-b714-3e6e-ad95-08490e2540c4")
	handler.AddExact(expectedSLORequest, filepath.Join(testDataFolder, "slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("service_response_time", expectedMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("static_slo_-_pass", 95, expectedSLORequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 2, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "service_response_time",
			DisplayName: "Service response time",
			Pass:        nil,
			Warning:     nil,
			Weight:      1,
		}, actual.Objectives[0])

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:     "static_slo_-_pass",
			Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{">=90.000000"}}},
			Warning: []*keptnapi.SLOCriteria{{Criteria: []string{">=75.000000"}}},
			Weight:  1,
		}, actual.Objectives[1])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)

}

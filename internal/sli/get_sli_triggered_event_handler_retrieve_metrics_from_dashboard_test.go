package sli

import (
	"context"
	"errors"
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// Retrieving (a single) SLI from a dashboard works, but Upload of dashboard, SLO or SLI file could fail
//
// prerequisites:
//   * we use a valid dashboard ID
//   * all processing and SLI result retrieval works
//   * if an upload of either SLO, SLI or dashboard file fails, then the test must fail
func TestErrorIsReturnedWhenSLISLOOrDashboardFileWritingFails(t *testing.T) {
	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000"

	failureAssertionsFunc := createFailedSLIResultAssertionsFunc(indicator)

	testConfigs := []struct {
		name                    string
		resourceClientMock      resourceClientInterface
		sliResultAssertionsFunc func(t *testing.T, actual sliResult)
		shouldFail              bool
	}{
		{
			name: "SLO upload fails",
			resourceClientMock: &uploadErrorResourceClientMock{
				t:              t,
				uploadSLOError: errors.New("SLO upload failed"),
			},
			sliResultAssertionsFunc: failureAssertionsFunc,
			shouldFail:              true,
		},
		// success case:
		{
			name: "upload of all files works",
			resourceClientMock: &uploadErrorResourceClientMock{
				t: t,
			},
			sliResultAssertionsFunc: createSuccessfulSLIResultAssertionsFunc(indicator, 12.439619479902443, expectedMetricsRequest),
			shouldFail:              false,
		},
	}

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/sli_via_dashboard_test/metric_definition_service-response-time.json")
	handler.AddExact(expectedMetricsRequest, "./testdata/sli_via_dashboard_test/response_time_p95_200_1_result.json")

	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
				if tc.shouldFail {
					assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
					assert.Contains(t, actual.Message, "upload failed")
				} else {
					assert.EqualValues(t, keptnv2.ResultPass, actual.Result)
					assert.Empty(t, actual.Message)
				}
			}

			runAndAssertThatDashboardTestIsCorrect(t, testGetSLIEventDataWithDefaultStartAndEnd, handler, tc.resourceClientMock, getSLIFinishedEventAssertionsFunc, tc.sliResultAssertionsFunc)
		})
	}
}

// Retrieving a dashboard by ID works, and we ignore the outdated parse behaviour
//
// prerequisites:
//   * we use a valid dashboard ID and it is returned by Dynatrace API
//   * The dashboard has 'KQG.QueryBehavior=ParseOnChange' set to only reparse the dashboard if it changed  (we do no longer consider this behaviour)
//   * we will not fallback to processing the stored SLI files, but process the dashboard again
func TestThatThereIsNoFallbackToSLIsFromDashboard(t *testing.T) {

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000"

	// we need metrics definition, because we will be retrieving metrics from dashboard
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli_parse_only_on_change.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/sli_via_dashboard_test/metric_definition_service-response-time.json")
	handler.AddExact(expectedMetricsRequest, "./testdata/sli_via_dashboard_test/response_time_p95_200_1_result.json")

	// sli and slo upload works
	rClient := &uploadErrorResourceClientMock{t: t}

	// value is divided by 1000 from dynatrace API result!
	runAndAssertThatDashboardTestIsCorrect(t, testGetSLIEventDataWithDefaultStartAndEnd, handler, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(indicator, 12.439619479902443, expectedMetricsRequest))
	assert.True(t, rClient.slosUploaded)
}

type uploadWillFailResourceClientMock struct {
	t *testing.T
}

func (m *uploadWillFailResourceClientMock) GetSLIs(_ context.Context, _ string, _ string, _ string) (map[string]string, error) {
	m.t.Fatalf("GetSLIs() should not be needed in this mock!")
	return nil, nil
}

func (m *uploadWillFailResourceClientMock) GetSLOs(_ context.Context, _ string, _ string, _ string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")
	return nil, nil
}

func (m *uploadWillFailResourceClientMock) UploadSLOs(_ context.Context, _ string, _ string, _ string, _ *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLOs() should not be needed in this mock!")
	return nil
}

// Retrieving (a single) SLI from a dashboard did not work, but no empty SLI or SLO files would be written
//
// prerequisites:
//   * we use a valid dashboard ID
//   * all processing works, but SLI result retrieval failed with 0 results (no data available)
//   * therefore SLI and SLO should be empty and an upload of either SLO or SLI should fail the test
func TestEmptySLOAndSLIAreNotWritten(t *testing.T) {
	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/sli_via_dashboard_test/metric_definition_service-response-time.json")
	handler.AddExact(expectedMetricsRequest, "./testdata/sli_via_dashboard_test/response_time_p95_200_0_results.json")

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultWarning, actual.Result)
		assert.Contains(t, actual.Message, "Metrics API v2 returned zero data points")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventDataWithDefaultStartAndEnd, getSLIFinishedEventAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(indicator, expectedMetricsRequest))
}

// Retrieving a dashboard by ID works, but dashboard processing did not produce any results, so we expect an error
//
// prerequisites:
//   * we use a valid dashboard ID and it is returned by Dynatrace API
//   * the dashboard does have a CustomCharting tile, but not the correct tile name, that would qualify it as SLI/SLO source
func TestThatFallbackToSLIsFromDashboardIfDashboardDidNotChangeWorks(t *testing.T) {

	// we do not need metrics definition and metrics query, because we will should not be looking into the tile
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/sli_via_dashboard_test/dashboard_custom_charting_without_matching_tile_name.json")
	// sli and slo should not happen, otherwise we fail
	rClient := &uploadWillFailResourceClientMock{t: t}

	getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
		assert.Contains(t, actual.Message, "any SLI results")
	}

	runAndAssertThatDashboardTestIsCorrect(t, testGetSLIEventDataWithDefaultStartAndEnd, handler, rClient, getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc(indicator))
}

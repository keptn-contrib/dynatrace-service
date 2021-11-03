package sli

import (
	"errors"
	"net/http"
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

type uploadErrorResourceClientMock struct {
	t              *testing.T
	uploadSLOError error
	sloUploaded    bool
	uploadSLIError error
	sliUploaded    bool
}

func (m *uploadErrorResourceClientMock) GetSLOs(project string, stage string, service string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")

	return nil, nil
}

func (m *uploadErrorResourceClientMock) UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error {
	if m.uploadSLIError != nil {
		return m.uploadSLIError
	}

	m.sliUploaded = true
	return nil
}

func (m *uploadErrorResourceClientMock) UploadSLOs(project string, stage string, service string, dashboardSLOs *keptnapi.ServiceLevelObjectives) error {
	if m.uploadSLOError != nil {
		return m.uploadSLOError
	}

	m.sloUploaded = true
	return nil
}

// Retrieving (a single) SLI from a dashboard works, but Upload of dashboard, SLO or SLI file could fail
//
// prerequisites:
//   * we use a valid dashboard ID
//   * all processing and SLI result retrieval works
//   * if an upload of either SLO, SLI or dashboard file fails, then the test must fail
func TestErrorIsReturnedWhenSLISLOOrDashboardFileWritingFails(t *testing.T) {

	failureAssertionsFunc := createSLIResultAssertionsFunc(indicator, 0, false)

	testConfigs := []struct {
		name                    string
		resourceClientMock      keptn.ResourceClientInterface
		sliResultAssertionsFunc func(t *testing.T, actual *keptnv2.SLIResult)
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
		{
			name: "SLI upload fails",
			resourceClientMock: &uploadErrorResourceClientMock{
				t:              t,
				uploadSLIError: errors.New("SLI upload failed"),
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
			sliResultAssertionsFunc: createSLIResultAssertionsFunc(indicator, 12.439619479902443, true),
			shouldFail:              false,
		},
	}

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/sli_via_dashboard_test/metric_definition_service-response-time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000",
		"./testdata/sli_via_dashboard_test/response_time_p95_200_1_result.json")

	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			// we do not need custom queries, as we are using the dashboard here
			// we need to instantiate this guy here!
			kClient := &keptnClientMock{}

			eventAssertionsFunc := func(data *keptnv2.GetSLIFinishedEventData) {
				if tc.shouldFail {
					assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
					assert.Contains(t, data.Message, "upload failed")
				} else {
					assert.EqualValues(t, keptnv2.ResultPass, data.Result)
					assert.Empty(t, data.Message)
				}
			}

			assertThatDashboardTestIsCorrect(t, handler, kClient, tc.resourceClientMock, eventAssertionsFunc, tc.sliResultAssertionsFunc)
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

	// we need metrics definition, because we will be retrieving metrics from dashboard
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli_parse_only_on_change.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/sli_via_dashboard_test/metric_definition_service-response-time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000",
		"./testdata/sli_via_dashboard_test/response_time_p95_200_1_result.json")

	// we do not need custom queries, as we are using the dashboard
	kClient := &keptnClientMock{}

	// sli and slo upload works
	rClient := &uploadErrorResourceClientMock{t: t}

	eventAssertionsFunc := func(data *keptnv2.GetSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultPass, data.Result)
		assert.Empty(t, data.Message)
	}

	// value is divided by 1000 from dynatrace API result!
	assertThatDashboardTestIsCorrect(t, handler, kClient, rClient, eventAssertionsFunc, createSLIResultAssertionsFunc(indicator, 12.439619479902443, true))
	assert.True(t, rClient.sliUploaded)
	assert.True(t, rClient.sloUploaded)
}

type uploadWillFailResourceClientMock struct {
	t *testing.T
}

func (m *uploadWillFailResourceClientMock) GetSLOs(project string, stage string, service string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")

	return nil, nil
}

func (m *uploadWillFailResourceClientMock) UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error {
	m.t.Fatalf("UploadSLI() should not be needed in this mock!")

	return nil
}

func (m *uploadWillFailResourceClientMock) UploadSLOs(project string, stage string, service string, dashboardSLOs *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLO() should not be needed in this mock!")

	return nil
}

// Retrieving (a single) SLI from a dashboard did not work, but no empty SLI or SLO files would be written
//
// prerequisites:
//   * we use a valid dashboard ID
//   * all processing works, but SLI result retrieval failed with 0 results (no data available)
//   * therefore SLI and SLO should be empty and an upload of either SLO or SLI should fail the test
func TestEmptySLOAndSLIAreNotWritten(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/sli_via_dashboard_test/metric_definition_service-response-time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000",
		"./testdata/sli_via_dashboard_test/response_time_p95_200_0_results.json")

	// we do not need custom queries, as we are using the dashboard here
	kClient := &keptnClientMock{}

	// if an upload of sli would be triggered then this test would fail
	rClient := &uploadWillFailResourceClientMock{t: t}

	eventAssertionsFunc := func(data *keptnv2.GetSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
		assert.Contains(t, data.Message, "any SLI results")
	}

	assertThatDashboardTestIsCorrect(t, handler, kClient, rClient, eventAssertionsFunc, createSLIResultAssertionsFunc(indicator, 0, false))
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

	// we do not need custom queries, as we are using the dashboard
	kClient := &keptnClientMock{}

	// sli and slo should not happen, otherwise we fail
	rClient := &uploadWillFailResourceClientMock{t: t}

	eventAssertionsFunc := func(data *keptnv2.GetSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
		assert.Contains(t, data.Message, "any SLI results")
	}

	assertThatDashboardTestIsCorrect(t, handler, kClient, rClient, eventAssertionsFunc, createSLIResultAssertionsFunc(indicator, 0, false))
}

func assertThatDashboardTestIsCorrect(t *testing.T, handler http.Handler, kClient *keptnClientMock, rClient keptn.ResourceClientInterface, eventAssertionsFunc func(data *keptnv2.GetSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual *keptnv2.SLIResult)) {
	setupTestAndAssertNoError(t, handler, kClient, rClient, testDashboardID)

	assertThatEventHasExpectedPayloadWithMatchingFunc(t, kClient.eventSink, eventAssertionsFunc, sliResultAssertionsFuncs...)
}

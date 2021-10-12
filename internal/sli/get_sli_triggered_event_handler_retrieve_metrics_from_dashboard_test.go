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
	t                    *testing.T
	dashboardContent     string
	uploadDashboardError error
	uploadSLOError       error
	uploadSLIError       error
}

func (m *uploadErrorResourceClientMock) GetSLOs(project string, stage string, service string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")

	return nil, nil
}

func (m *uploadErrorResourceClientMock) UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error {
	if m.uploadSLIError != nil {
		return m.uploadSLIError
	}

	return nil
}

func (m *uploadErrorResourceClientMock) UploadSLOs(project string, stage string, service string, dashboardSLOs *keptnapi.ServiceLevelObjectives) error {
	if m.uploadSLOError != nil {
		return m.uploadSLOError
	}

	return nil
}

func (m *uploadErrorResourceClientMock) GetDashboard(project string, stage string, service string) (string, error) {
	return m.dashboardContent, nil
}

func (m *uploadErrorResourceClientMock) UploadDashboard(project string, stage string, service string, dashboard *dynatrace.Dashboard) error {
	if m.uploadDashboardError != nil {
		return m.uploadDashboardError
	}

	return nil
}

// Retrieving (a single) SLI from a dashboard works, but Upload of dashboard, SLO or SLI file could fail
//
// prerequisites:
//   * we use a valid dashboard ID
//   * all processing and SLI result retrieval works
//   * if an upload of either SLO, SLI or dashboard file fails, then the test must fail
func TestErrorIsReturnedWhenSLISLOOrDashboardFileWritingFails(t *testing.T) {

	failureAssertionFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.False(t, actual.Success)
	}

	testConfigs := []struct {
		name               string
		resourceClientMock keptn.ResourceClientInterface
		assertFunc         func(t *testing.T, actual *keptnv2.SLIResult)
		shouldFail         bool
	}{
		// failure cases:
		{
			name: "dashboard upload fails",
			resourceClientMock: &uploadErrorResourceClientMock{
				t:                    t,
				uploadDashboardError: errors.New("dashboard upload failed"),
			},
			assertFunc: failureAssertionFunc,
			shouldFail: true,
		},
		{
			name: "SLO upload fails",
			resourceClientMock: &uploadErrorResourceClientMock{
				t:              t,
				uploadSLOError: errors.New("SLO upload failed"),
			},
			assertFunc: failureAssertionFunc,
			shouldFail: true,
		},
		{
			name: "SLI upload fails",
			resourceClientMock: &uploadErrorResourceClientMock{
				t:              t,
				uploadSLIError: errors.New("SLI upload failed"),
			},
			assertFunc: failureAssertionFunc,
			shouldFail: true,
		},
		// success case:
		{
			name: "upload of all files works",
			resourceClientMock: &uploadErrorResourceClientMock{
				t: t,
			},
			assertFunc: func(t *testing.T, actual *keptnv2.SLIResult) {
				assert.EqualValues(t, indicator, actual.Metric)
				assert.EqualValues(t, 12.439619479902443, actual.Value) // div by 1000 from dynatrace API result!
				assert.True(t, actual.Success)
			},
			shouldFail: false,
		},
	}

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/12345678-1111-4444-8888-123456789012", "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli.json")
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

			assertThatDashboardTestIsCorrect(t, handler, kClient, tc.resourceClientMock, tc.assertFunc, eventAssertionsFunc)
		})
	}
}

// Retrieving (a single) SLI from a dashboard did not work, but no empty SLI or SLO files would be written
//
// prerequisites:
//   * we use a valid dashboard ID
//   * all processing works, but SLI result retrieval failed with 0 results (no data available)
//   * therefore SLI and SLO should be empty and an upload of either SLO or SLI should fail the test
func TestEmptySLOAndSLIAreNotWritten(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/12345678-1111-4444-8888-123456789012", "./testdata/sli_via_dashboard_test/dashboard_custom_charting_single_sli.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", "./testdata/sli_via_dashboard_test/metric_definition_service-response-time.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000",
		"./testdata/sli_via_dashboard_test/response_time_p95_200_0_results.json")

	// we do not need custom queries, as we are using the dashboard here
	kClient := &keptnClientMock{}

	// if an upload of sli would be triggered then this test should fail, because the result fails
	rClient := &uploadErrorResourceClientMock{
		t:              t,
		uploadSLOError: errors.New("SLO upload failed"),
		uploadSLIError: errors.New("SLI upload failed"),
	}

	eventAssertionsFunc := func(data *keptnv2.GetSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultPass, data.Result)
		assert.Empty(t, data.Message)
	}

	// we do not want to assert the expected result, because there won't be any in this case
	assertThatDashboardTestIsCorrect(t, handler, kClient, rClient, nil, eventAssertionsFunc)
}

func assertThatDashboardTestIsCorrect(t *testing.T, handler http.Handler, kClient *keptnClientMock, rClient keptn.ResourceClientInterface, assertionsFunc func(t *testing.T, actual *keptnv2.SLIResult), eventAssertionsFunc func(data *keptnv2.GetSLIFinishedEventData)) {
	setupTestAndAssertNoError(t, handler, kClient, rClient, "12345678-1111-4444-8888-123456789012")

	assertThatEventHasExpectedPayloadWithMatchingFunc(t, assertionsFunc, kClient.eventSink, eventAssertionsFunc)
}
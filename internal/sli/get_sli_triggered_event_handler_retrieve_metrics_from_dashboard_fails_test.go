package sli

import (
	"net/http"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// Retrieving a dashboard by an invalid ID returns an error
//
// prerequisites:
//   * we use an invalid dashboard ID Dynatrace API returns a 400 error
//   * the event can have multiple indicators or none. (There is an SLO file in Keptn and the SLO files may contain indicators)
//
// We do not want to see the error attached to any indicator coming from SLO files, but attached to a "no metric" indicator
func TestThatInvalidDashboardIDProducesErrorMessageInNoMetricIndicatorEvenIfThereAreIndicators(t *testing.T) {

	testConfigs := []struct {
		name            string
		eventIndicators []string
	}{
		{
			name:            "no indicators defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{},
		},
		{
			name:            "one indicator defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{"single-indicator-from-slo-file"},
		},
		{
			name:            "multiple indicators defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{"first-indicator-from-slo-file", "second-indicator-from-slo-file", "third-indicator-from-slo-file"},
		},
	}

	const dashboardID = "some-invalid-dashboard-id"
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath+"/"+dashboardID, 400, "./testdata/sli_via_dashboard_test/dashboard_invalid_uuid_400.json")

	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			testEvent := &getSLIEventData{
				project:    "sockshop",
				stage:      "staging",
				service:    "carts",
				indicators: tc.eventIndicators,
				sliStart:   "", // use defaults here
				sliEnd:     "", // use defaults here
			}

			// sli and slo upload works
			rClient := &uploadErrorResourceClientMock{t: t}

			getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *keptnv2.GetSLIFinishedEventData) {
				assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
				assert.Contains(t, actual.Message, dashboardID)
				assert.Contains(t, actual.Message, "400")
				assert.Contains(t, actual.Message, "Constraints violated")
			}

			runAndAssertDashboardTest(t, testEvent, handler, rClient, dashboardID, getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc("no metric"))
		})
	}
}

func runAndAssertDashboardTest(t *testing.T, getSLIEventData *getSLIEventData, handler http.Handler, rClient keptn.ResourceClientInterface, dashboardID string, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *keptnv2.GetSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual *keptnv2.SLIResult)) {

	// we do not need custom queries, as we are using the dashboard
	kClient := &keptnClientMock{}

	runTestAndAssertNoError(t, getSLIEventData, handler, kClient, rClient, dashboardID)

	assertCorrectGetSLIEvents(t, kClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
}

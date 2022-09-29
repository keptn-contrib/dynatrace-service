package sli

import (
	"path/filepath"
	"strconv"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// Retrieving a dashboard by an invalid ID returns an error
//
// prerequisites:
//   - we use
//   - an invalid dashboard ID and Dynatrace API returns a 400 error, or
//   - a valid, but not found dashboard ID and Dynatrace API returns a 404
//   - the event can have multiple indicators or none. (There is an SLO file in Keptn and the SLO files may contain indicators)
//
// We do not want to see the error attached to any indicator coming from SLO files, but attached to a "no metric" indicator
func TestThatInvalidDashboardIDProducesErrorMessageInNoMetricIndicatorEvenIfThereAreIndicators(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/basic/no_metric_errors/"

	type definition struct {
		errorCode    int
		errorMessage string
		dashboardID  string
		payload      string
	}

	invalidID := definition{
		errorCode:    400,
		errorMessage: "Constraints violated",
		dashboardID:  "some-invalid-dashboard-id",
		payload:      filepath.Join(testDataFolder, "dashboard_invalid_uuid_400.json"),
	}
	idNotFound := definition{
		errorCode:    404,
		errorMessage: "not found",
		dashboardID:  testDashboardID,
		payload:      filepath.Join(testDataFolder, "dashboard_id_not_found_404.json"),
	}

	testConfigs := []struct {
		name            string
		eventIndicators []string
		def             definition
	}{
		{
			name:            "no indicators defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{},
			def:             invalidID,
		},
		{
			name:            "one indicator defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{"single-indicator-from-slo-file"},
			def:             invalidID,
		},
		{
			name:            "multiple indicators defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{"first-indicator-from-slo-file", "second-indicator-from-slo-file", "third-indicator-from-slo-file"},
			def:             invalidID,
		},
		{
			name:            "no indicators defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{},
			def:             idNotFound,
		},
		{
			name:            "one indicator defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{"single-indicator-from-slo-file"},
			def:             idNotFound,
		},
		{
			name:            "multiple indicators defined in event (SLO file) will produce single SLI result with name 'no metric'",
			eventIndicators: []string{"first-indicator-from-slo-file", "second-indicator-from-slo-file", "third-indicator-from-slo-file"},
			def:             idNotFound,
		},
	}

	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExactError(dynatrace.DashboardsPath+"/"+tc.def.dashboardID, tc.def.errorCode, tc.def.payload)

			getSLIFinishedEventAssertionsFunc := func(t *testing.T, actual *getSLIFinishedEventData) {
				assert.EqualValues(t, keptnv2.ResultFailed, actual.Result)
				assert.Contains(t, actual.Message, tc.def.dashboardID)
				assert.Contains(t, actual.Message, strconv.Itoa(tc.def.errorCode))
				assert.Contains(t, actual.Message, tc.def.errorMessage)
			}

			runGetSLIsFromDashboardTestWithConfigClientAndDashboardParameterAndCheckSLIs(t, handler, newConfigClientMockThatAllowsUploadSLOs(t), testGetSLIEventData, tc.def.dashboardID, getSLIFinishedEventAssertionsFunc, createFailedSLIResultAssertionsFunc(NoMetricIndicator))
		})
	}
}

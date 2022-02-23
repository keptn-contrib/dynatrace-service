package dashboard

import (
	"net/http"
	"testing"
	"time"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestQueryDynatraceDashboardForSLIs(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)
	// we handle these if the URLs are a full match
	handler.AddExact(dynatrace.DashboardsPath, "./testdata/test_get_dashboards.json")
	handler.AddExact(dynatrace.DashboardsPath+"/12345678-1111-4444-8888-123456789012", "./testdata/test_get_dashboards_id.json")

	handler.AddStartsWith("/api/v2/slo", "./testdata/test_get_slo_id.json")
	handler.AddStartsWith("/api/v2/problems", "./testdata/test_get_problems.json")
	handler.AddStartsWith("/api/v2/securityProblems", "./testdata/test_get_securityproblems.json")

	querying, _, teardown := createQueryingWithHandler(t, keptnEvent, handler)
	defer teardown()

	timeframe, err := common.NewTimeframeParser("2019-10-21T09:11:24Z", "2019-10-21T09:11:25Z").Parse()
	assert.NoError(t, err)
	result, err := querying.GetSLIValues(common.DynatraceConfigDashboardQUERY, *timeframe)

	assert.Nil(t, err)
	assert.NotNil(t, result, "No result returned")
	assert.NotNil(t, result.dashboardLink, "No dashboard link label generated")
	assert.NotNil(t, result.dashboard, "No Dashboard JSON returned")
	assert.NotNil(t, result.sli, "No SLI returned")
	assert.NotNil(t, result.slo, "No SLO returned")
	assert.NotNil(t, result.sliResults, "No SLI Results returned")

	const expectedSLOs = 3
	assert.Equal(t, expectedSLOs, len(result.sli.Indicators))
	assert.Equal(t, expectedSLOs, len(result.slo.Objectives))
	assert.EqualValues(t, &keptnapi.SLOScore{Pass: "90%", Warning: "70%"}, result.slo.TotalScore)
	assert.EqualValues(
		t,
		&keptnapi.SLOComparison{
			CompareWith:               "single_result",
			IncludeResultWithScore:    "pass",
			NumberOfComparisonResults: 1,
			AggregateFunction:         "avg",
		},
		result.slo.Comparison)
	assert.Equal(t, expectedSLOs, len(result.sliResults))

	for _, objective := range result.slo.Objectives {
		assert.NotNil(t, objective)
	}

	for _, sliResult := range result.sliResults {
		assert.True(t, sliResult.Success)
	}
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// also the ID of the dashboard we try to retrieve was not found
func TestRetrieveDashboardWithUnknownButValidID(t *testing.T) {
	// we need do not care about the event here
	ev := &test.EventData{}

	const dashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9c"

	// we add a handler to simulate a very concrete 404 Dashboards API request/response in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath+"/"+dashboardID, http.StatusNotFound, "./testdata/test_query_dynatrace_dashboard_dashboard_id_not_found.json")

	querying, _, teardown := createCustomQuerying(t, ev, handler)
	defer teardown()

	timeframe, err := common.NewTimeframe(time.Now(), time.Now())
	assert.NoError(t, err)
	actualResult, err := querying.GetSLIValues(dashboardID, *timeframe)

	assert.Error(t, err)
	assert.Nil(t, actualResult)

	var apiErr *dynatrace.APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusNotFound, apiErr.Code())
		assert.Contains(t, apiErr.Message(), dashboardID)
	}
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// it could happen that you have a copy/paste error in your ID (invalid UUID) so we will see a different error
func TestRetrieveDashboardWithInvalidID(t *testing.T) {
	// we need do not care about the event here
	ev := &test.EventData{}

	const dashboardID = "definitely-invalid-uuid"

	// we add a handler to simulate a very concrete 400 Dashboards API request/response in this case.
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(dynatrace.DashboardsPath+"/"+dashboardID, http.StatusBadRequest, "./testdata/test_query_dynatrace_dashboard_dashboard_id_not_valid.json")

	querying, _, teardown := createCustomQuerying(t, ev, handler)
	defer teardown()

	timeframe, err := common.NewTimeframe(time.Now(), time.Now())
	assert.NoError(t, err)
	actualResult, err := querying.GetSLIValues(dashboardID, *timeframe)

	assert.Error(t, err)
	assert.Nil(t, actualResult)

	var apiErr *dynatrace.APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusBadRequest, apiErr.Code())
	}
	assert.Contains(t, err.Error(), "UUID")
}

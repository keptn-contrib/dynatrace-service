package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueryDynatraceDashboardForSLIs(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	result, err := dh.QueryDynatraceDashboardForSLIs(keptnEvent, common.DynatraceConfigDashboardQUERY, startTime, endTime)

	assert.Nil(t, err)
	assert.NotNil(t, result, "No result returned")
	assert.NotNil(t, result.dashboardLink, "No dashboard link label generated")
	assert.NotNil(t, result.dashboard, "No Dashboard JSON returned")
	assert.NotNil(t, result.sli, "No SLI returned")
	assert.NotNil(t, result.slo, "No SLO returned")
	assert.NotNil(t, result.sliResults, "No SLI Results returned")

	const expectedSLOs = 14
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
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") and there is no dashboard already stored
// in Keptn resources then we do not do anything
func TestNoQueryingOfDashboardNecessaryDueToNotSpecifiedAndNotDashboardInKeptn(t *testing.T) {
	// we don't care about event data in this case
	ev := &GetSLITriggeredEvent{}

	// no need for a handler as we should not query Dynatrace API here
	retrieval, _, teardown := createCustomRetrieval(ev, nil, KeptnClientMock{}, DashboardReaderMock{err: "no dashboard"})
	defer teardown()

	const dashboardID = ""

	result, err := retrieval.QueryDynatraceDashboardForSLIs(ev, dashboardID, time.Now(), time.Now())

	assert.Nil(t, result)
	assert.Nil(t, err)
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") but there is a dashboard already stored
// in Keptn resources then we do the fallback to querying Dynatrace API for a dashboard
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithFailingDynatraceRequest(t *testing.T) {
	// we don't care about event data in this case
	ev := &GetSLITriggeredEvent{}

	// we don't care about the content of the dashboard here, because it just should not be empty!
	// also we don't add a handler to simulate a failing request (404) in this case.
	retrieval, _, teardown := createCustomRetrieval(ev, nil, KeptnClientMock{}, DashboardReaderMock{content: "some dashboard content"})
	defer teardown()

	const dashboardID = ""

	result, err := retrieval.QueryDynatraceDashboardForSLIs(ev, dashboardID, time.Now(), time.Now())

	assert.Nil(t, result)
	assert.Nil(t, err)
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") but there is a dashboard already stored
// in Keptn resources then we do the fallback to querying Dynatrace API for a dashboard
//
// we will find dashboards but none with a name, that would match the required format: KQG;project=%project%;service=%service%;stage=%stage%;xxx
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithSuccessfulDynatraceRequest(t *testing.T) {
	// we don't care about event data in this case
	ev := &GetSLITriggeredEvent{}

	// we add a handler to simulate an successful Dashboards API request in this case.
	handler := test.NewURLHandler()
	handler.AddExact("/api/config/v1/dashboards", "./testfiles/test_query_dynatrace_dashboard_dashboards.json")

	// we don't care about the content of the dashboard here, because it just should not be empty!
	retrieval, _, teardown := createCustomRetrieval(ev, handler, KeptnClientMock{}, DashboardReaderMock{content: "some dashboard content"})
	defer teardown()

	const dashboardID = ""

	result, err := retrieval.QueryDynatraceDashboardForSLIs(ev, dashboardID, time.Now(), time.Now())

	assert.Nil(t, result)
	assert.Nil(t, err)
}

// If you do not specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "") but there is a dashboard already stored
// in Keptn resources then we do the fallback to querying Dynatrace API for a dashboard
//
// we will find dashboards and one with a name, that would match the required format: KQG;project=%project%;service=%service%;stage=%stage%;xxx
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithSuccessfulDynatraceRequestAndMatchingDashboard(t *testing.T) {
	// we need to match project, stage & service against the dashboard name
	ev := createKeptnEvent("project-1", "stage-3", "service-7")

	// we add a handler to simulate an successful Dashboards API request in this case.
	handler := test.NewURLHandler()
	handler.AddExact("/api/config/v1/dashboards", "./testfiles/test_query_dynatrace_dashboard_dashboards_kqg.json")
	handler.AddExact("/api/config/v1/dashboards/e03f4be0-4712-4f12-96ee-8c486d001e9b", "./testfiles/test_query_dynatrace_dashboard_dashboard_kqg.json")

	// we don't care about the content of the dashboard here, because it just should not be empty!
	retrieval, _, teardown := createCustomRetrieval(ev, handler, KeptnClientMock{}, DashboardReaderMock{content: "some dashboard content"})
	defer teardown()

	const dashboardID = ""

	result, err := retrieval.QueryDynatraceDashboardForSLIs(ev, dashboardID, time.Now(), time.Now())

	assert.Nil(t, err)
	// TODO 2021-09-16: refine for result checks on equality here!
	assert.NotNil(t, result)
}

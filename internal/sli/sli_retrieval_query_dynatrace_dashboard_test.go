package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const dashboardURL = "/api/config/v1/dashboards"

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
// also we will fail because of a failing request to Dynatrace API
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithFailingDynatraceRequest(t *testing.T) {
	// we don't care about event data in this case
	ev := &GetSLITriggeredEvent{}

	// we add a handler to simulate a failing dashboards API request (401 in this case)
	handler := test.NewURLHandler()
	handler.AddExactError(dashboardURL, http.StatusUnauthorized, "./testfiles/dynatrace_missing_authorization_error.json")

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
	handler.AddExact(dashboardURL, "./testfiles/test_query_dynatrace_dashboard_dashboards.json")

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
// also the retrieved dashboard is the same as the stored one and we have enabled "KQG.QueryBehavior=ParseOnChange" in our markdown tile
// so no need to do anything with the contents of the dashboard here.
func TestQueryingOfDashboardNecessaryDueToNotSpecifiedButStoredDashboardInKeptnWithSuccessfulDynatraceRequestAndMatchingDashboard(t *testing.T) {
	// we need to match project, stage & service against the dashboard name
	ev := createKeptnEvent("project-1", "stage-3", "service-7")

	const matchingDashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9b"
	// we need to make sure to use the "processed" one and not the original Dynatrace JSON, because we do have different
	// models for our DTOs (Tile structures are generic in our part - we need to support any tile type)
	const storedDashboardFile = "./testfiles/test_query_dynatrace_dashboard_dashboard_kqg.json"

	// we add a handle to simulate an successful Dashboards API request in this case.
	handler := test.NewURLHandler()
	handler.AddExact(dashboardURL, "./testfiles/test_query_dynatrace_dashboard_dashboards_kqg.json")
	handler.AddExact(dashboardURL+"/"+matchingDashboardID, storedDashboardFile)

	dashboardContent, err := ioutil.ReadFile(storedDashboardFile)
	if err != nil {
		panic(err)
	}

	// we need to make sure that the mocked reader returns the "processed" Dynatrace Dashboard
	retrieval, url, teardown := createCustomRetrieval(ev, handler, KeptnClientMock{}, DashboardReaderMock{content: string(dashboardContent)})
	defer teardown()

	const dashboardID = ""

	from := time.Date(2021, 9, 17, 7, 0, 0, 0, time.UTC)
	to := time.Date(2021, 9, 17, 8, 0, 0, 0, time.UTC)
	actualResult, err := retrieval.QueryDynatraceDashboardForSLIs(ev, dashboardID, from, to)

	expectedResult := newDashboardQueryResultFrom(&DashboardLink{
		apiURL:         url,
		startTimestamp: from,
		endTimestamp:   to,
		dashboardID:    matchingDashboardID,
	})

	assert.Nil(t, err)
	assert.EqualValues(t, expectedResult, actualResult)
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// also the retrieved dashboard is the same as the stored one and we have enabled "KQG.QueryBehavior=ParseOnChange" in our markdown tile
// so no need to do anything with the contents of the dashboard here.
// (this is basically the same as above without querying the dashboards API for a matching dashboard beforehand)
func TestRetrieveDashboardWithValidIDAndStoredDashboardInKeptnIsTheSame(t *testing.T) {
	// we need to match project, stage & service against the dashboard name
	ev := createKeptnEvent("project-1", "stage-3", "service-7")

	const dashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9b"
	// we need to make sure to use the "processed" one and not the original Dynatrace JSON, because we do have different
	// models for our DTOs (Tile structures are generic in our part - we need to support any tile type)
	const storedDashboardFile = "./testfiles/test_query_dynatrace_dashboard_dashboard_kqg.json"

	// we add a handle to simulate an successful Dashboards API request in this case.
	handler := test.NewURLHandler()
	handler.AddExact(dashboardURL+"/"+dashboardID, storedDashboardFile)

	dashboardContent, err := ioutil.ReadFile(storedDashboardFile)
	if err != nil {
		panic(err)
	}

	// we need to make sure that the mocked reader returns the "processed" Dynatrace Dashboard
	retrieval, url, teardown := createCustomRetrieval(ev, handler, KeptnClientMock{}, DashboardReaderMock{content: string(dashboardContent)})
	defer teardown()

	from := time.Date(2021, 9, 17, 7, 0, 0, 0, time.UTC)
	to := time.Date(2021, 9, 17, 8, 0, 0, 0, time.UTC)
	actualResult, err := retrieval.QueryDynatraceDashboardForSLIs(ev, dashboardID, from, to)

	expectedResult := newDashboardQueryResultFrom(&DashboardLink{
		apiURL:         url,
		startTimestamp: from,
		endTimestamp:   to,
		dashboardID:    dashboardID,
	})

	assert.Nil(t, err)
	assert.EqualValues(t, expectedResult, actualResult)
}

// If you do specify a Dashboard in dynatrace.conf.yaml (-> dashboard: "<some-dashboard-uuid>") then we will retrieve the
// dashboard via the Dynatrace API.
// also the ID of the dashboard we try to retrieve was not found
func TestRetrieveDashboardWithInvalidID(t *testing.T) {
	// we need do not care about the event here
	ev := &GetSLITriggeredEvent{}

	const dashboardID = "e03f4be0-4712-4f12-96ee-8c486d001e9c"

	// we add a handler to simulate a very concrete 404 Dashboards API request/response in this case.
	handler := test.NewURLHandler()
	handler.AddExactError(dashboardURL+"/"+dashboardID, http.StatusNotFound, "./testfiles/test_query_dynatrace_dashboard_dashboard_id_not_found.json")

	// we also do not care about the dashboard that would be returned by keptn
	retrieval, _, teardown := createCustomRetrieval(ev, handler, KeptnClientMock{}, DashboardReaderMock{})
	defer teardown()

	actualResult, err := retrieval.QueryDynatraceDashboardForSLIs(ev, dashboardID, time.Now(), time.Now())

	assert.Error(t, err)
	var apiErr *dynatrace.APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.Code())
	assert.Contains(t, apiErr.Message(), dashboardID)
	assert.Nil(t, actualResult)
}

package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"net/http"
	"testing"
)

func TestFindDynatraceDashboardSuccess(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/config/v1/dashboards", "./testdata/test_get_dashboards.json")

	dh, teardown := createDashboardRetrieval(keptnEvent, handler)
	defer teardown()

	dashboardID, err := dh.findDynatraceDashboard()

	if err != nil {
		t.Error(err)
	}

	if dashboardID != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("findDynatraceDashboard not finding quality gate dashboard")
	}
}

func TestFindDynatraceDashboardNoneExistingDashboard(t *testing.T) {
	keptnEvent := createKeptnEvent("BAD PROJECT", QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/config/v1/dashboards", "./testdata/test_get_dashboards.json")

	dh, teardown := createDashboardRetrieval(keptnEvent, handler)
	defer teardown()

	dashboardID, err := dh.findDynatraceDashboard()

	if err != nil {
		t.Error(err)
	}

	if dashboardID != "" {
		t.Errorf("findDynatraceDashboard found a dashboard that should not have been found: " + dashboardID)
	}
}

func TestLoadDynatraceDashboardWithQUERY(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/config/v1/dashboards", "./testdata/test_get_dashboards.json")
	handler.AddExact("/api/config/v1/dashboards/12345678-1111-4444-8888-123456789012", "./testdata/test_get_dashboards_id.json")

	dh, teardown := createDashboardRetrieval(keptnEvent, handler)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.Retrieve(common.DynatraceConfigDashboardQUERY)

	if dashboardJSON == nil {
		t.Errorf("Didnt query dashboard for quality gate project even though it shoudl exist: " + dashboard)
	}

	if dashboard != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("Didnt query the dashboard that matches the project/stage/service names: " + dashboard)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoadDynatraceDashboardWithID(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/config/v1/dashboards/12345678-1111-4444-8888-123456789012", "./testdata/test_get_dashboards_id.json")

	dh, teardown := createDashboardRetrieval(keptnEvent, handler)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.Retrieve(QUALITYGATE_DASHBOARD_ID)

	if dashboardJSON == nil {
		t.Errorf("Didnt query dashboard for quality gate project even though it should exist by ID")
	}

	if dashboard != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("loadDynatraceDashboard should return the passed in dashboard id")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoadDynatraceDashboardWithEmptyDashboard(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)

	dh, teardown := createDashboardRetrieval(keptnEvent, handler)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.Retrieve("")

	if dashboardJSON != nil {
		t.Errorf("No dashboard should be loaded if no dashboard is passed")
	}

	if dashboard != "" {
		t.Errorf("dashboard should be empty as by default we dont load a dashboard")
	}

	if err != nil {
		t.Error(err)
	}
}

func createDashboardRetrieval(eventData adapter.EventContentAdapter, handler http.Handler) (*Retrieval, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	retrieval := NewRetrieval(
		dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		eventData)

	return retrieval, teardown
}

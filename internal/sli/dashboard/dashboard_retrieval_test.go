package dashboard

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestLoadDynatraceDashboardWithEmptyDashboard(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)

	handler := test.NewFileBasedURLHandler(t)

	dh, teardown := createDashboardRetrieval(t, keptnEvent, handler)
	defer teardown()

	dashboardJSON, dashboard, err := dh.Retrieve(context.TODO(), "")

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid 'dashboard'")
	}
	assert.Nil(t, dashboardJSON)
	assert.Empty(t, dashboard)
}

func createDashboardRetrieval(t *testing.T, eventData adapter.EventContentAdapter, handler http.Handler) (*Retrieval, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	retrieval := NewRetrieval(
		dynatrace.NewClientWithHTTP(createDynatraceCredentials(t, url), httpClient),
		eventData)

	return retrieval, teardown
}

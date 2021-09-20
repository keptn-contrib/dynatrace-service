package dashboard

import (
	"errors"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const QUALITYGATE_DASHBOARD_ID = "12345678-1111-4444-8888-123456789012"
const QUALITYGATE_PROJECT = "qualitygate"
const QUALTIYGATE_SERVICE = "evalservice"
const QUALITYGATE_STAGE = "qualitystage"

// createKeptnEvent creates a new Keptn Event for project, stage and service
func createKeptnEvent(project string, stage string, service string) adapter.EventContentAdapter {
	return &test.EventData{
		Project: project,
		Stage:   stage,
		Service: service,
	}
}

func createQueryingWithHandler(keptnEvent adapter.EventContentAdapter, handler http.Handler) (*Querying, string, func()) {
	return createCustomQuerying(keptnEvent, handler, DashboardReaderMock{})
}

func createCustomQuerying(keptnEvent adapter.EventContentAdapter, handler http.Handler, reader keptn.DashboardResourceReaderInterface) (*Querying, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	dh := NewQuerying(
		keptnEvent,
		nil,
		dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		reader)

	return dh, url, teardown
}

func TestCreateQueryingWithHandler(t *testing.T) {
	keptnEvent := createKeptnEvent("sockshop", "dev", "carts")
	dh, url, teardown := createQueryingWithHandler(keptnEvent, nil)
	defer teardown()

	c := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	assert.EqualValues(t, c, dh.dtClient.Credentials())
	assert.EqualValues(t, keptnEvent, dh.eventData)
	assert.EqualValues(t, DashboardReaderMock{}, dh.dashboardReader)
}

type DashboardReaderMock struct {
	content string
	err     string
}

func (m DashboardReaderMock) GetDashboard(project string, stage string, service string) (string, error) {
	if m.err != "" {
		return "", errors.New(m.err)
	}

	return m.content, nil
}

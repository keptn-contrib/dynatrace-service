package dashboard

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
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
	dynatraceClient, url, teardown := createDynatraceClient(handler)

	dh := NewQuerying(
		keptnEvent,
		nil,
		dynatraceClient,
		reader)

	return dh, url, teardown
}

// TODO: 2021-10-08: Can this be moved to test package and shared?
func createDynatraceClient(handler http.Handler) (dynatrace.ClientInterface, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DynatraceCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	dh := dynatrace.NewClientWithHTTP(dtCredentials, httpClient)

	return dh, url, teardown
}

func TestCreateQueryingWithHandler(t *testing.T) {
	keptnEvent := createKeptnEvent("sockshop", "dev", "carts")
	dh, url, teardown := createQueryingWithHandler(keptnEvent, nil)
	defer teardown()

	c := &credentials.DynatraceCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	assert.EqualValues(t, c, dh.dtClient.Credentials())
	assert.EqualValues(t, keptnEvent, dh.eventData)
	assert.EqualValues(t, DashboardReaderMock{}, dh.dashboardReader)
}

type DashboardReaderMock struct {
	content string
	err     error
}

func (m DashboardReaderMock) GetDashboard(project string, stage string, service string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.content, nil
}

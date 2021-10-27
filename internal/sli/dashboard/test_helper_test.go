package dashboard

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const testDynatraceAPIToken = "dtOc01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"

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

func createDynatraceCredentials(t *testing.T, url string) *credentials.DynatraceCredentials {
	dynatraceCredentials, err := credentials.NewDynatraceCredentials(url, testDynatraceAPIToken)
	assert.NoError(t, err)
	return dynatraceCredentials
}

func createQueryingWithHandler(t *testing.T, keptnEvent adapter.EventContentAdapter, handler http.Handler) (*Querying, string, func()) {
	return createCustomQuerying(t, keptnEvent, handler)
}

func createCustomQuerying(t *testing.T, keptnEvent adapter.EventContentAdapter, handler http.Handler) (*Querying, string, func()) {
	dynatraceClient, url, teardown := createDynatraceClient(t, handler)

	dh := NewQuerying(
		keptnEvent,
		nil,
		dynatraceClient)

	return dh, url, teardown
}

// TODO: 2021-10-08: Can this be moved to test package and shared?
func createDynatraceClient(t *testing.T, handler http.Handler) (dynatrace.ClientInterface, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dh := dynatrace.NewClientWithHTTP(createDynatraceCredentials(t, url), httpClient)

	return dh, url, teardown
}

func TestCreateQueryingWithHandler(t *testing.T) {
	keptnEvent := createKeptnEvent("sockshop", "dev", "carts")
	dh, url, teardown := createQueryingWithHandler(t, keptnEvent, nil)
	defer teardown()

	assert.EqualValues(t, createDynatraceCredentials(t, url), dh.dtClient.Credentials())
	assert.EqualValues(t, keptnEvent, dh.eventData)
}

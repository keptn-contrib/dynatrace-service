package dynatrace

import (
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func createDynatraceClient(t *testing.T, handler http.Handler) (ClientInterface, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials, err := credentials.NewDynatraceCredentials(url, "test")
	assert.NoError(t, err)

	dh := NewClientWithHTTP(dtCredentials, httpClient)

	return dh, url, teardown
}

package dynatrace

import (
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"net/http"
)

func createDynatraceClient(handler http.Handler) (ClientInterface, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	dh := NewClientWithHTTP(dtCredentials, httpClient)

	return dh, url, teardown
}

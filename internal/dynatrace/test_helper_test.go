package dynatrace

import (
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func createDynatraceClient(handler http.Handler) (ClientInterface, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DynatraceCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	dh := NewClientWithHTTP(dtCredentials, httpClient)

	return dh, url, teardown
}

package dynatrace

import (
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestDynatraceHelper_createClient_with_proxy(t *testing.T) {
	const mockTenant = "https://mySampleEnv.live.dynatrace.com"
	const mockProxy = "https://proxy-abcdefgh123:8080"

	os.Setenv("HTTP_PROXY", mockProxy)
	os.Setenv("HTTPS_PROXY", mockProxy)
	os.Setenv("NO_PROXY", "localhost")

	dt := NewClient(createDynatraceCredentials(t, mockTenant))
	_, _, url, err := dt.restClient.Get("/api/v1/config/clusterversion")

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "proxy-abcdefgh123")
	}
	assert.Empty(t, url)

	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("NO_PROXY")
}

func TestExecuteDynatraceREST(t *testing.T) {
	expected := []byte("my-error")
	expectedStatusCode := http.StatusNotFound
	h := test.CreateHandler(expected, expectedStatusCode)

	client, teardown := testingDynatraceClient(t, h)
	defer teardown()

	actual, err := client.Get("/invalid-url")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), string(expected))
	assert.Contains(t, err.Error(), strconv.Itoa(expectedStatusCode))
	assert.EqualValues(t, expected, actual)
}

func TestExecuteDynatraceRESTBadRequest(t *testing.T) {
	expected := []byte("my-message")
	h := test.CreateHandler(expected, 200)

	client, teardown := testingDynatraceClient(t, h)
	defer teardown()

	actual, err := client.Get("/valid-url")

	assert.Nil(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestDynatraceClient(t *testing.T) {
	response := []byte("response")
	payload := []byte("payload")

	tests := []struct {
		name               string
		expectedResponse   []byte
		expectedStatusCode int
		responseFunc       func(*Client) ([]byte, error)
		shouldBeAPIError   bool
	}{
		{
			name:               "GET, 200",
			expectedResponse:   response,
			expectedStatusCode: http.StatusOK,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Get("/valid-url") },
		},
		{
			name:               "GET, 404",
			expectedResponse:   response,
			expectedStatusCode: http.StatusNotFound,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Get("/not-found-url") },
			shouldBeAPIError:   true,
		},
		{
			name:               "POST, 200",
			expectedResponse:   response,
			expectedStatusCode: http.StatusOK,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Post("/valid-url", payload) },
		},
		{
			name:               "POST, 204",
			expectedResponse:   []byte{},
			expectedStatusCode: http.StatusNoContent,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Post("/valid-url", payload) },
		},
		{
			name:               "POST, 404",
			expectedResponse:   response,
			expectedStatusCode: http.StatusNotFound,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Post("/not-found-url", payload) },
			shouldBeAPIError:   true,
		},
		{
			name:               "PUT, 200",
			expectedResponse:   response,
			expectedStatusCode: http.StatusOK,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Put("/valid-url", payload) },
		},
		{
			name:               "PUT, 204",
			expectedResponse:   []byte{},
			expectedStatusCode: http.StatusNoContent,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Put("/valid-url", payload) },
		},
		{
			name:               "PUT, 404",
			expectedResponse:   response,
			expectedStatusCode: http.StatusNotFound,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Put("/not-found-url", payload) },
			shouldBeAPIError:   true,
		},
		{
			name:               "DELETE, 200",
			expectedResponse:   response,
			expectedStatusCode: http.StatusOK,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Delete("/valid-url") },
		},
		{
			name:               "DELETE, 204",
			expectedResponse:   []byte{},
			expectedStatusCode: http.StatusNoContent,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Delete("/valid-url") },
		},
		{
			name:               "DELETE, 404",
			expectedResponse:   response,
			expectedStatusCode: http.StatusNotFound,
			responseFunc:       func(client *Client) ([]byte, error) { return client.Delete("/not-found-url") },
			shouldBeAPIError:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := test.CreateHandler(tt.expectedResponse, tt.expectedStatusCode)

			client, teardown := testingDynatraceClient(t, h)
			defer teardown()

			actualResponse, err := tt.responseFunc(client)

			if tt.shouldBeAPIError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), string(tt.expectedResponse))
				assert.Contains(t, err.Error(), strconv.Itoa(tt.expectedStatusCode))
				assert.EqualValues(t, tt.expectedResponse, actualResponse)
			} else {
				assert.Nil(t, err)
				assert.EqualValues(t, tt.expectedResponse, actualResponse)
			}
		})
	}
}

func testingDynatraceClient(t *testing.T, handler http.Handler) (*Client, func()) {
	httpClient, teardown := test.CreateHTTPClient(handler)

	client := NewClientWithHTTP(
		createDynatraceCredentials(t, "http://my-tenant.dynatrace.com"),
		httpClient)

	return client, teardown
}

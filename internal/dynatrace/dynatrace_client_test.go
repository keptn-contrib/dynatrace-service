package dynatrace

import (
	"bytes"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

func TestDynatraceHelper_createClient(t *testing.T) {

	mockTenant := "https://mySampleEnv.live.dynatrace.com"
	mockReq, err := http.NewRequest("GET", mockTenant+"/api/v1/config/clusterversion", bytes.NewReader(make([]byte, 100)))
	if err != nil {
		t.Errorf("Client.createClient(): unable to make mock request: error = %v", err)
		return
	}

	mockProxy := "https://proxy:8080"
	t.Logf("Using mock proxy: %v", mockProxy)

	type proxyEnvVars struct {
		httpProxy  string
		httpsProxy string
		noProxy    string
	}
	type fields struct {
		DynatraceCreds *credentials.DTCredentials
	}
	type args struct {
		req *http.Request
	}

	// only one test can be run in a single test run due to the ProxyConfig environment being cached
	// see envProxyFunc() in transport.go for details
	tests := []struct {
		name         string
		proxyEnvVars proxyEnvVars
		fields       fields
		args         args
		wantErr      bool
		wantProxy    string
	}{
		{
			name: "testWithProxy",
			proxyEnvVars: proxyEnvVars{
				httpProxy:  mockProxy,
				httpsProxy: mockProxy,
				noProxy:    "localhost",
			},
			fields: fields{
				DynatraceCreds: &credentials.DTCredentials{
					Tenant:   mockTenant,
					ApiToken: "",
				},
			},
			args: args{
				req: mockReq,
			},
			wantProxy: mockProxy,
		},
		/*{
			name: "testWithNoProxy",
			fields: fields{
				DynatraceCreds: &credentials.DTCredentials{
					Tenant:   mockTenant,
					ApiToken: "",
				},
				Logger: keptncommon.NewLogger("", "", ""),
			},
			args: args{
				req: mockReq,
			},
			wantProxy: "",
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("HTTP_PROXY", tt.proxyEnvVars.httpProxy)
			os.Setenv("HTTPS_PROXY", tt.proxyEnvVars.httpsProxy)
			os.Setenv("NO_PROXY", tt.proxyEnvVars.noProxy)

			dt := NewClient(tt.fields.DynatraceCreds)

			gotTransport := dt.httpClient.Transport.(*http.Transport)
			gotProxyUrl, err := gotTransport.Proxy(tt.args.req)
			if err != nil {
				t.Errorf("Client.createClient() error = %v", err)
				return
			}

			if gotProxyUrl == nil {
				if tt.wantProxy != "" {
					t.Errorf("Client.createClient() error, got proxy is nil, wanted = %v", tt.wantProxy)
				}
			} else {
				gotProxy := gotProxyUrl.String()
				if tt.wantProxy == "" {
					t.Errorf("Client.createClient() error, got proxy = %v, wanted nil", gotProxy)
				} else if gotProxy != tt.wantProxy {
					t.Errorf("Client.createClient() error, got proxy = %v, wanted = %v", gotProxy, tt.wantProxy)
				}
			}

			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("NO_PROXY")
		})
	}
}

func TestExecuteDynatraceREST(t *testing.T) {
	expected := []byte("my-error")
	expectedStatusCode := http.StatusNotFound
	h := test.CreateHandler(expected, expectedStatusCode)

	client, teardown := testingDynatraceClient(h)
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

	client, teardown := testingDynatraceClient(h)
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

			client, teardown := testingDynatraceClient(h)
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

func testingDynatraceClient(handler http.Handler) (*Client, func()) {
	httpClient, teardown := test.CreateHTTPClient(handler)

	client := NewClientWithHTTP(
		&credentials.DTCredentials{
			Tenant:   "http://my-tenant.dynatrace.com",
			ApiToken: "abcdefgh12345678",
		},
		httpClient)

	return client, teardown
}

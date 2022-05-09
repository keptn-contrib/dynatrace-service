package dynatrace

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
	"github.com/keptn-contrib/dynatrace-service/internal/rest"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

type EnvironmentAPIv2Error struct {
	Error struct {
		Code                 int                  `json:"code"`
		Message              string               `json:"message"`
		ConstraintViolations ConstraintViolations `json:"constraintViolations"`
	} `json:"error"`
}

type ConstraintViolation struct {
	Path              string `json:"path"`
	Message           string `json:"message"`
	ParameterLocation string `json:"parameterLocation"`
	Location          string `json:"location"`
}

func (v ConstraintViolation) String() string {
	return v.Message
}

type ConstraintViolations []ConstraintViolation

func (vs ConstraintViolations) String() string {
	messages := make([]string, len(vs))
	for i, v := range vs {
		messages[i] = fmt.Sprintf("[path: %s - msg: %s]", v.Path, v.Message)
	}

	return strings.TrimRight(
		strings.Join(messages, ", "),
		", ")
}

type APIError struct {
	code    int
	message string
	uri     string
	details *EnvironmentAPIv2Error
}

func (e *APIError) Code() int {
	return e.code
}

func (e *APIError) Message() string {
	return e.message
}

func (e *APIError) Error() string {
	if e.details != nil {
		return fmt.Sprintf("Dynatrace API error (%d): %s %s - URL: %s", e.code, e.message, e.details.Error.ConstraintViolations, e.uri)
	}

	return fmt.Sprintf("Dynatrace API error (%d): %s - URL: %s", e.code, e.message, e.uri)
}

func createAdditionalHeaders(token string) rest.HTTPHeader {
	header := rest.HTTPHeader{}
	header.Add("Authorization", "Api-Token "+token)

	return header
}

type ClientInterface interface {
	// Get performs a get request.
	Get(ctx context.Context, apiPath string) ([]byte, error)

	// Post performs a post request.
	Post(ctx context.Context, apiPath string, body []byte) ([]byte, error)

	// Put performs a put request.
	Put(ctx context.Context, apiPath string, body []byte) ([]byte, error)

	// Delete performs a delete request.
	Delete(ctx context.Context, apiPath string) ([]byte, error)

	// Credentials returns the credentials associated with the client.
	Credentials() *credentials.DynatraceCredentials
}

type Client struct {
	credentials *credentials.DynatraceCredentials
	restClient  rest.ClientInterface
}

// NewClient creates a new Client
func NewClient(dynatraceCredentials *credentials.DynatraceCredentials) *Client {
	return NewClientWithHTTP(
		dynatraceCredentials,
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !env.IsHttpSSLVerificationEnabled()},
				Proxy: http.ProxyFromEnvironment,
			},
		},
	)
}

func NewClientWithHTTP(dynatraceCredentials *credentials.DynatraceCredentials, httpClient *http.Client) *Client {
	return &Client{
		credentials: dynatraceCredentials,
		restClient: rest.NewClient(
			httpClient,
			dynatraceCredentials.GetTenant(),
			createAdditionalHeaders(dynatraceCredentials.GetAPIToken())),
	}
}

// Get performs a get request.
func (dt *Client) Get(ctx context.Context, apiPath string) ([]byte, error) {
	body, status, url, err := dt.restClient.Get(ctx, apiPath)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

// Post performs a post request.
func (dt *Client) Post(ctx context.Context, apiPath string, body []byte) ([]byte, error) {
	body, status, url, err := dt.restClient.Post(ctx, apiPath, body)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

// Put performs a put request.
func (dt *Client) Put(ctx context.Context, apiPath string, body []byte) ([]byte, error) {
	body, status, url, err := dt.restClient.Put(ctx, apiPath, body)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

// Delete performs a delete request.
func (dt *Client) Delete(ctx context.Context, apiPath string) ([]byte, error) {
	body, status, url, err := dt.restClient.Delete(ctx, apiPath)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

// validates the response and returns the payload or Keptn API error
func validateResponse(body []byte, status int, url string) ([]byte, error) {
	if status < 200 || status >= 300 {

		// try to get the error information
		dtAPIError := &EnvironmentAPIv2Error{}
		err := json.Unmarshal(body, dtAPIError)
		if err != nil {
			return body, &APIError{
				code:    status,
				message: string(body),
				uri:     url,
			}
		}
		return body, &APIError{
			code:    dtAPIError.Error.Code,
			message: dtAPIError.Error.Message,
			details: dtAPIError,
			uri:     url,
		}
	}

	return body, nil
}

// Credentials returns the credentials associated with the client.
func (dt *Client) Credentials() *credentials.DynatraceCredentials {
	return dt.credentials
}

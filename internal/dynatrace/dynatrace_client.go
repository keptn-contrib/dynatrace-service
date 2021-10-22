package dynatrace

import (
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
	Get(apiPath string) ([]byte, error)
	Post(apiPath string, body []byte) ([]byte, error)
	Put(apiPath string, body []byte) ([]byte, error)
	Delete(apiPath string) ([]byte, error)

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

func (dt *Client) Get(apiPath string) ([]byte, error) {
	body, status, url, err := dt.restClient.Get(apiPath)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

func (dt *Client) Post(apiPath string, body []byte) ([]byte, error) {
	body, status, url, err := dt.restClient.Post(apiPath, body)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

func (dt *Client) Put(apiPath string, body []byte) ([]byte, error) {
	body, status, url, err := dt.restClient.Put(apiPath, body)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

func (dt *Client) Delete(apiPath string) ([]byte, error) {
	body, status, url, err := dt.restClient.Delete(apiPath)
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

func (dt *Client) Credentials() *credentials.DynatraceCredentials {
	return dt.credentials
}

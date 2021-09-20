package dynatrace

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
	log "github.com/sirupsen/logrus"

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
		messages[i] = v.Message
	}

	return strings.Join(messages, ", ")
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

type ClientError struct {
	message string
	cause   error
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("Dynatrace client error: %s [%v]", e.message, e.cause)
}

type ClientInterface interface {
	Get(apiPath string) ([]byte, error)
	Post(apiPath string, body []byte) ([]byte, error)
	Put(apiPath string, body []byte) ([]byte, error)
	Delete(apiPath string) ([]byte, error)

	Credentials() *credentials.DTCredentials
}

type Client struct {
	credentials *credentials.DTCredentials
	httpClient  *http.Client
}

// NewClient creates a new Client
func NewClient(dynatraceCreds *credentials.DTCredentials) *Client {
	return NewClientWithHTTP(
		dynatraceCreds,
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !env.IsHttpSSLVerificationEnabled()},
				Proxy: http.ProxyFromEnvironment,
			},
		},
	)
}

func NewClientWithHTTP(dynatraceCreds *credentials.DTCredentials, httpClient *http.Client) *Client {
	return &Client{
		credentials: dynatraceCreds,
		httpClient:  httpClient,
	}
}

func (dt *Client) Get(apiPath string) ([]byte, error) {
	return dt.sendRequest(apiPath, http.MethodGet, nil)
}

func (dt *Client) Post(apiPath string, body []byte) ([]byte, error) {
	return dt.sendRequest(apiPath, http.MethodPost, body)
}

func (dt *Client) Put(apiPath string, body []byte) ([]byte, error) {
	return dt.sendRequest(apiPath, http.MethodPut, body)
}

func (dt *Client) Delete(apiPath string) ([]byte, error) {
	return dt.sendRequest(apiPath, http.MethodDelete, nil)
}

// sendRequest makes an Dynatrace API request and returns the response
func (dt *Client) sendRequest(apiPath string, method string, body []byte) ([]byte, error) {

	req, err := dt.createRequest(apiPath, method, body)
	if err != nil {
		return nil, err
	}

	response, err := dt.doRequest(req)
	if err != nil {
		return response, err
	}

	return response, nil
}

// creates http request for api call with appropriate headers including authorization
func (dt *Client) createRequest(apiPath string, method string, body []byte) (*http.Request, error) {
	var url string
	if !strings.HasPrefix(dt.credentials.Tenant, "http://") && !strings.HasPrefix(dt.credentials.Tenant, "https://") {
		url = "https://" + dt.credentials.Tenant + apiPath
	} else {
		url = dt.credentials.Tenant + apiPath
	}

	log.WithFields(log.Fields{"method": method, "url": url}).Debug("creating Dynatrace API request")

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, &ClientError{
			message: "failed to create request",
			cause:   err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Token "+dt.credentials.ApiToken)
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+os.Getenv("version"))

	return req, nil
}

// performs the request and reads the response
func (dt *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := dt.httpClient.Do(req)
	if err != nil {
		return nil, &ClientError{
			message: "failed to send request",
			cause:   err,
		}
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &ClientError{
			message: "failed to read response body",
			cause:   err,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {

		// try to get the error information
		dtAPIError := &EnvironmentAPIv2Error{}
		err := json.Unmarshal(responseBody, dtAPIError)
		if err != nil {
			return responseBody, &APIError{
				code:    resp.StatusCode,
				message: string(responseBody),
				uri:     req.URL.String(),
			}
		}
		return responseBody, &APIError{
			code:    dtAPIError.Error.Code,
			message: dtAPIError.Error.Message,
			details: dtAPIError,
			uri:     req.URL.String(),
		}
	}

	return responseBody, nil
}

func (dt *Client) Credentials() *credentials.DTCredentials {
	return dt.credentials
}

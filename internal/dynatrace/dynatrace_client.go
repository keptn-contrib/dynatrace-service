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
		Code                 int    `json:"code"`
		Message              string `json:"message"`
		ConstraintViolations []struct {
			Path              string `json:"path"`
			Message           string `json:"message"`
			ParameterLocation string `json:"parameterLocation"`
			Location          string `json:"location"`
		} `json:"constraintViolations"`
	} `json:"error"`
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
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	response, err := dt.doRequest(req)
	if err != nil {
		return response, fmt.Errorf("failed to do request: %v", err)
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
		return nil, fmt.Errorf("failed to create new request: %v", err)
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
		return nil, fmt.Errorf("failed to send Dynatrace API request: %v", err)
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {

		// try to get the error information
		dtAPIError := &EnvironmentAPIv2Error{}
		err := json.Unmarshal(responseBody, dtAPIError)
		if err != nil {
			return responseBody, fmt.Errorf("request to Dynatrace API returned status code %d and response %s", resp.StatusCode, string(responseBody))
		}
		return responseBody, fmt.Errorf("request to Dynatrace API returned error %d: %s", dtAPIError.Error.Code, dtAPIError.Error.Message)
	}

	return responseBody, nil
}

func (dt *Client) Credentials() *credentials.DTCredentials {
	return dt.credentials
}

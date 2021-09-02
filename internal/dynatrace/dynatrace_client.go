package dynatrace

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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

type Client struct {
	DynatraceCreds *credentials.DTCredentials
	HTTPClient     *http.Client
}

// NewClient creates a new Client
func NewClient(dynatraceCreds *credentials.DTCredentials) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !lib.IsHttpSSLVerificationEnabled()},
		Proxy:           http.ProxyFromEnvironment,
	}
	return &Client{
		DynatraceCreds: dynatraceCreds,
		HTTPClient:     &http.Client{Transport: tr},
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
	if !strings.HasPrefix(dt.DynatraceCreds.Tenant, "http://") && !strings.HasPrefix(dt.DynatraceCreds.Tenant, "https://") {
		url = "https://" + dt.DynatraceCreds.Tenant + apiPath
	} else {
		url = dt.DynatraceCreds.Tenant + apiPath
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Token "+dt.DynatraceCreds.ApiToken)
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+os.Getenv("version"))

	return req, nil
}

// performs the request and reads the response
func (dt *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := dt.HTTPClient.Do(req)
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

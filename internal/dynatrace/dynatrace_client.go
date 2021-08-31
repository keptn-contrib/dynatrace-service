package dynatrace

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

type Client struct {
	DynatraceCreds *credentials.DTCredentials
}

// NewClient creates a new Client
func NewClient(dynatraceCreds *credentials.DTCredentials) *Client {
	return &Client{
		DynatraceCreds: dynatraceCreds,
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

	if common.RunLocal || common.RunLocalTest {
		log.WithFields(
			log.Fields{
				"tenant": dt.DynatraceCreds.Tenant,
				"body":   string(body),
			}).Info("Dynatrace.sendRequest(RUNLOCAL) - not sending event to tenant")
		return nil, nil
	}

	req, err := dt.createRequest(apiPath, method, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	client, err := dt.createClient(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	response, err := dt.doRequest(client, req)
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

// creates http client with proxy and TLS configuration
func (dt *Client) createClient(req *http.Request) (*http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !lib.IsHttpSSLVerificationEnabled()},
		Proxy:           http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: tr}

	return client, nil
}

// performs the request and reads the response
func (dt *Client) doRequest(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Dynatrace API request: %v", err)
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseBody, fmt.Errorf("api request failed with status %s and response %s", resp.Status, string(responseBody))
	}

	return responseBody, nil
}

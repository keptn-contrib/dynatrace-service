package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const NoStatus = -1

type ClientInterface interface {
	Get(apiPath string) ([]byte, int, string, error)
	Post(apiPath string, body []byte) ([]byte, int, string, error)
	Put(apiPath string, body []byte) ([]byte, int, string, error)
	Delete(apiPath string) ([]byte, int, string, error)
}

type HTTPHeader map[string][]string

func (h HTTPHeader) Add(key string, value string) {
	h[key] = append(h[key], value)
}

type ClientError struct {
	message string
	cause   error
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("HTTP client error: %s [%v]", e.message, e.cause)
}

type Client struct {
	httpClient       *http.Client
	baseURL          string
	additionalHeader HTTPHeader
}

// NewClient creates a new Client
func NewClient(httpClient *http.Client, baseURL string, additionalHeader HTTPHeader) *Client {
	return &Client{
		httpClient:       httpClient,
		baseURL:          baseURL,
		additionalHeader: additionalHeader,
	}
}

// NewDefaultClient creates a new Client with a default HTTP client set up
func NewDefaultClient(httpClient *http.Client, baseURL string) *Client {
	return NewClient(httpClient, baseURL, HTTPHeader{})
}

func (c *Client) Get(apiPath string) ([]byte, int, string, error) {
	return c.sendRequest(apiPath, http.MethodGet, nil)
}

func (c *Client) Post(apiPath string, body []byte) ([]byte, int, string, error) {
	return c.sendRequest(apiPath, http.MethodPost, body)
}

func (c *Client) Put(apiPath string, body []byte) ([]byte, int, string, error) {
	return c.sendRequest(apiPath, http.MethodPut, body)
}

func (c *Client) Delete(apiPath string) ([]byte, int, string, error) {
	return c.sendRequest(apiPath, http.MethodDelete, nil)
}

// sendRequest makes an API request and returns the response and the status code or an error
// the response will not contain any data in case of an error
func (c *Client) sendRequest(apiPath string, method string, body []byte) ([]byte, int, string, error) {

	req, err := c.createRequest(apiPath, method, body)
	if err != nil {
		return nil, NoStatus, "", err
	}

	return c.doRequest(req)
}

// creates http request for api call with appropriate headers including authorization
func (c *Client) createRequest(apiPath string, method string, body []byte) (*http.Request, error) {
	var url = c.baseURL + apiPath

	log.WithFields(log.Fields{"method": method, "url": url}).Debug("creating Dynatrace API request")

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, &ClientError{
			message: "failed to create request",
			cause:   err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+os.Getenv("version"))

	for key, values := range c.additionalHeader {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}

// performs the request and reads the response
func (c *Client) doRequest(req *http.Request) ([]byte, int, string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NoStatus, "", &ClientError{
			message: "failed to send request",
			cause:   err,
		}
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NoStatus, "", &ClientError{
			message: "failed to read response body",
			cause:   err,
		}
	}

	return responseBody, resp.StatusCode, req.URL.String(), nil
}

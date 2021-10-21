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
	Get(apiPath string) ([]byte, int, error)
	Post(apiPath string, body []byte) ([]byte, int, error)
	Put(apiPath string, body []byte) ([]byte, int, error)
	Delete(apiPath string) ([]byte, int, error)
}

type ClientError struct {
	message string
	cause   error
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("HTTP client error: %s [%v]", e.message, e.cause)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Client
func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// NewDefaultClient creates a new Client with a default HTTP client set up
func NewDefaultClient(baseURL string) *Client {
	return NewClient(&http.Client{}, baseURL)
}

func (c *Client) Get(apiPath string) ([]byte, int, error) {
	return c.sendRequest(apiPath, http.MethodGet, nil)
}

func (c *Client) Post(apiPath string, body []byte) ([]byte, int, error) {
	return c.sendRequest(apiPath, http.MethodPost, body)
}

func (c *Client) Put(apiPath string, body []byte) ([]byte, int, error) {
	return c.sendRequest(apiPath, http.MethodPut, body)
}

func (c *Client) Delete(apiPath string) ([]byte, int, error) {
	return c.sendRequest(apiPath, http.MethodDelete, nil)
}

// sendRequest makes an API request and returns the response and the status code or an error
// the response will not contain any data in case of an error
func (c *Client) sendRequest(apiPath string, method string, body []byte) ([]byte, int, error) {

	req, err := c.createRequest(apiPath, method, body)
	if err != nil {
		return nil, NoStatus, err
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

	return req, nil
}

// performs the request and reads the response
func (c *Client) doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NoStatus, &ClientError{
			message: "failed to send request",
			cause:   err,
		}
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NoStatus, &ClientError{
			message: "failed to read response body",
			cause:   err,
		}
	}

	return responseBody, resp.StatusCode, nil
}

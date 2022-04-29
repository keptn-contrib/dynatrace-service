package rest

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
	log "github.com/sirupsen/logrus"
)

const NoStatus = -1

type ClientInterface interface {
	// Get performs an HTTP get request.
	Get(ctx context.Context, apiPath string) ([]byte, int, string, error)

	// Post performs an HTTP post request.
	Post(ctx context.Context, apiPath string, body []byte) ([]byte, int, string, error)

	// Put performs an HTTP put request.
	Put(ctx context.Context, apiPath string, body []byte) ([]byte, int, string, error)

	// Delete performs an HTTP delete request.
	Delete(ctx context.Context, apiPath string) ([]byte, int, string, error)
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

// NewClient creates a new Client.
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

// Get performs an HTTP get request.
func (c *Client) Get(ctx context.Context, apiPath string) ([]byte, int, string, error) {
	return c.sendRequest(ctx, apiPath, http.MethodGet, nil)
}

// Post performs an HTTP post request.
func (c *Client) Post(ctx context.Context, apiPath string, body []byte) ([]byte, int, string, error) {
	return c.sendRequest(ctx, apiPath, http.MethodPost, body)
}

// Put performs an HTTP put request.
func (c *Client) Put(ctx context.Context, apiPath string, body []byte) ([]byte, int, string, error) {
	return c.sendRequest(ctx, apiPath, http.MethodPut, body)
}

// Delete performs an HTTP delete request.
func (c *Client) Delete(ctx context.Context, apiPath string) ([]byte, int, string, error) {
	return c.sendRequest(ctx, apiPath, http.MethodDelete, nil)
}

// sendRequest makes an API request and returns the response and the status code or an error.
// The response will not contain any data in case of an error.
func (c *Client) sendRequest(ctx context.Context, apiPath string, method string, body []byte) ([]byte, int, string, error) {

	req, err := c.createRequest(ctx, apiPath, method, body)
	if err != nil {
		return nil, NoStatus, "", err
	}

	return c.doRequest(req)
}

// createRequest creates an HTTP request for an API call with appropriate headers including authorization.
func (c *Client) createRequest(ctx context.Context, apiPath string, method string, body []byte) (*http.Request, error) {
	var url = c.baseURL + apiPath

	log.WithFields(log.Fields{"method": method, "url": url}).Debug("creating HTTP request")

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, &ClientError{
			message: "failed to create request",
			cause:   err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+env.GetVersion())

	for key, values := range c.additionalHeader {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}

// doRequest performs the request and reads the response.
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

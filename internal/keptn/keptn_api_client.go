package keptn

import (
	"encoding/json"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/rest"
)

type APIClientInterface interface {
	Post(apiPath string, body []byte) ([]byte, error)
}

type APIClient struct {
	restClient rest.ClientInterface
}

func NewAPIClient(client rest.ClientInterface) *APIClient {
	return &APIClient{
		restClient: client,
	}
}

func (c *APIClient) Post(apiPath string, body []byte) ([]byte, error) {
	body, status, url, err := c.restClient.Post(apiPath, body)
	if err != nil {
		return nil, err
	}

	return validateResponse(body, status, url)
}

// genericAPIErrorDTO will support multiple Keptn API errors
type genericAPIErrorDTO struct {
	Code      int    `json:"code"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

func (e *genericAPIErrorDTO) status() int {
	if e.Code != 0 {
		return e.Code
	}

	return e.ErrorCode
}

type APIError struct {
	status  int
	message string
	url     string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Keptn API error (%d): %s - URL: %s", e.status, e.message, e.url)
}

func validateResponse(body []byte, status int, url string) ([]byte, error) {
	if status < 200 || status >= 300 {
		// try to get the error information
		keptnAPIError := &genericAPIErrorDTO{}
		err := json.Unmarshal(body, keptnAPIError)
		if err != nil {
			return body, &APIError{
				status:  status,
				message: string(body),
				url:     url,
			}
		}

		return nil, &APIError{
			status:  keptnAPIError.status(),
			message: keptnAPIError.Message,
			url:     url,
		}
	}

	return body, nil
}

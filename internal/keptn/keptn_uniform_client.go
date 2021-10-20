package keptn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

type registrationResponse struct {
	ID string `json:"id"`
}

type UniformClientInterface interface {
	GetServiceNames(project string, stage string) ([]string, error)
	CreateServiceInProject(project string, service string) error
}

type UniformClient struct {
	httpClient *http.Client
}

func NewDefaultUniformClient() *UniformClient {
	return NewUniformClient(
		&http.Client{})
}

func NewUniformClient(httpClient *http.Client) *UniformClient {
	return &UniformClient{
		httpClient: httpClient,
	}
}

func (c *UniformClient) GetIntegrationIDFor(integrationName string) (string, error) {
	// TODO 2021-10-20: extract the http client and consolidate with services (and with dynatrace http client)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/uniform/registration?name=%s", common.GetShipyardControllerURL(), integrationName), bytes.NewReader(nil))
	if err != nil {
		return "", fmt.Errorf("could not create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("request failed with %d: %s", resp.StatusCode, string(body))
	}

	var responses []registrationResponse
	err = json.Unmarshal(body, &responses)
	if err != nil {
		return "", fmt.Errorf("could not parse Keptn Uniform API response")
	}

	if len(responses) == 0 {
		return "", fmt.Errorf("could not retrieve integration ID for %s", integrationName)
	}

	if len(responses) > 1 {
		return "", fmt.Errorf("there are more than one integrations with name %s - this is not supported", integrationName)
	}

	return responses[0].ID, nil
}

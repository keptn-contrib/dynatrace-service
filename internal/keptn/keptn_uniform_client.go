package keptn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/rest"
)

type registrationResponse struct {
	ID string `json:"id"`
}

const UniformPath = "/v1/uniform/registration"

type UniformClientInterface interface {
	GetServiceNames(project string, stage string) ([]string, error)
	CreateServiceInProject(project string, service string) error
}

type UniformClient struct {
	client APIClientInterface
}

func NewDefaultUniformClient() *UniformClient {
	return NewUniformClient(
		&http.Client{})
}

func NewUniformClient(httpClient *http.Client) *UniformClient {
	return &UniformClient{
		client: NewAPIClient(
			rest.NewDefaultClient(
				httpClient,
				common.GetShipyardControllerURL())),
	}
}

func (c *UniformClient) GetIntegrationIDFor(integrationName string) (string, error) {
	body, err := c.client.Get(UniformPath + "?name=" + integrationName)
	if err != nil {
		return "", err
	}

	var responses []registrationResponse
	err = json.Unmarshal(body, &responses)
	if err != nil {
		return "", fmt.Errorf("could not parse Keptn Uniform API response: %w", err)
	}

	if len(responses) == 0 {
		return "", fmt.Errorf("could not retrieve integration ID for %s", integrationName)
	}

	if len(responses) > 1 {
		return "", fmt.Errorf("there are more than one integrations with name %s - this is not supported", integrationName)
	}

	return responses[0].ID, nil
}

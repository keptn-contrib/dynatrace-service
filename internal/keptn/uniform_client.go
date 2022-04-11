package keptn

import (
	"fmt"

	api "github.com/keptn/go-utils/pkg/api/utils"
)

// UniformClientInterface provides access to Keptn Uniform.
type UniformClientInterface interface {
	// GetIntegrationIDByName gets the ID of the integration with specified name or returns an error if none or more than one exist with that name.
	GetIntegrationIDByName(integrationName string) (string, error)
}

// UniformClient is an implementation of UniformClientInterface using api.UniformV1Interface.
type UniformClient struct {
	client api.UniformV1Interface
}

// NewUniformClient creates a new UniformClient using the specified api.UniformV1Interface.
func NewUniformClient(client api.UniformV1Interface) *UniformClient {
	return &UniformClient{
		client: client,
	}
}

// GetIntegrationIDByName gets the ID of the integration with specified name or returns an error if none or more than one exist with that name.
func (c *UniformClient) GetIntegrationIDByName(integrationName string) (string, error) {
	integrations, err := c.client.GetRegistrations()
	if err != nil {
		return "", fmt.Errorf("could not get Keptn Uniform registrations: %w", err)
	}

	var integrationIDs []string
	for _, integration := range integrations {
		if integration.Name == integrationName {
			integrationIDs = append(integrationIDs, integration.ID)
		}
	}

	if len(integrationIDs) == 0 {
		return "", fmt.Errorf("could not retrieve integration ID for %s", integrationName)
	}

	if len(integrationIDs) > 1 {
		return "", fmt.Errorf("there are more than one integrations with name %s - this is not supported", integrationName)
	}

	return integrationIDs[0], nil
}

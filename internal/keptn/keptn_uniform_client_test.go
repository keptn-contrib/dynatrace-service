package keptn

import (
	"testing"

	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/stretchr/testify/assert"
)

type mockUniformClient struct {
	registrations []*models.Integration
}

func (c mockUniformClient) Ping(integrationID string) (*models.Integration, error) {
	panic("Ping should not be called on mockUniformClient")
}

func (c mockUniformClient) RegisterIntegration(integration models.Integration) (string, error) {
	panic("RegisterIntegration should not be called on mockUniformClient")
}

func (c mockUniformClient) CreateSubscription(integrationID string, subscription models.EventSubscription) (string, error) {
	panic("CreateSubscription should not be called on mockUniformClient")
}

func (c mockUniformClient) UnregisterIntegration(integrationID string) error {
	panic("UnregisterIntegration should not be called on mockUniformClient")
}

func (c mockUniformClient) GetRegistrations() ([]*models.Integration, error) {
	return c.registrations, nil
}

func TestUniformClient_GetIntegrationIDByName(t *testing.T) {

	tests := []struct {
		name                   string
		uniformClient          api.UniformV1Interface
		integrationName        string
		expectedIntegrationID  string
		expectError            bool
		expectedErrorSubstring string
	}{
		{
			name: "one integration with name - should work",
			uniformClient: &mockUniformClient{
				registrations: []*models.Integration{
					{
						ID:   "5d64eb87c4ce3e23935758c418df9d980c16a3b1",
						Name: "webhook-service",
					},
					{
						ID:   "e8039ff0b65c7e4d326a0473a18f04cabfefe747",
						Name: "dynatrace-service",
					},
				},
			},
			integrationName:       "dynatrace-service",
			expectedIntegrationID: "e8039ff0b65c7e4d326a0473a18f04cabfefe747",
		},
		{
			name: "two integrations with name - should fail",
			uniformClient: &mockUniformClient{
				registrations: []*models.Integration{
					{
						ID:   "5d64eb87c4ce3e23935758c418df9d980c16a3b1",
						Name: "webhook-service",
					},
					{
						ID:   "e8039ff0b65c7e4d326a0473a18f04cabfefe747",
						Name: "dynatrace-service",
					},
					{
						ID:   "d73082b9c42aa147935fe2592a91eb5d2b224038",
						Name: "dynatrace-service",
					},
				},
			},
			integrationName:        "dynatrace-service",
			expectError:            true,
			expectedErrorSubstring: "more than one integrations with name",
		},
		{
			name: "no integration with name - should fail",
			uniformClient: &mockUniformClient{
				registrations: []*models.Integration{
					{
						ID:   "5d64eb87c4ce3e23935758c418df9d980c16a3b1",
						Name: "webhook-service",
					},
				},
			},
			integrationName:        "dynatrace-service",
			expectError:            true,
			expectedErrorSubstring: "could not retrieve integration ID for",
		},
		{
			name: "no integrations at all - should fail",
			uniformClient: &mockUniformClient{
				registrations: []*models.Integration{},
			},
			integrationName:        "dynatrace-service",
			expectError:            true,
			expectedErrorSubstring: "could not retrieve integration ID for",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := UniformClient{
				client: tt.uniformClient,
			}
			integrationID, err := c.GetIntegrationIDByName(tt.integrationName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorSubstring)
			} else {
				assert.EqualValues(t, tt.expectedIntegrationID, integrationID)
				assert.NoError(t, err)
			}
		})
	}
}

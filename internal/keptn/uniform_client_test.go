package keptn

import (
	"context"
	"testing"

	"github.com/keptn/go-utils/pkg/api/models"
	v2 "github.com/keptn/go-utils/pkg/api/utils/v2"
	"github.com/stretchr/testify/assert"
)

type mockUniformClient struct {
	registrations []*models.Integration
}

func (c mockUniformClient) Ping(_ context.Context, _ string, _ v2.UniformPingOptions) (*models.Integration, error) {
	panic("Ping should not be called on mockUniformClient")
}

func (c mockUniformClient) RegisterIntegration(_ context.Context, _ models.Integration, _ v2.UniformRegisterIntegrationOptions) (string, error) {
	panic("RegisterIntegration should not be called on mockUniformClient")
}

func (c mockUniformClient) CreateSubscription(_ context.Context, _ string, _ models.EventSubscription, _ v2.UniformCreateSubscriptionOptions) (string, error) {
	panic("CreateSubscription should not be called on mockUniformClient")
}

func (c mockUniformClient) UnregisterIntegration(_ context.Context, _ string, _ v2.UniformUnregisterIntegrationOptions) error {
	panic("UnregisterIntegration should not be called on mockUniformClient")
}

func (c mockUniformClient) GetRegistrations(_ context.Context, _ v2.UniformGetRegistrationsOptions) ([]*models.Integration, error) {
	return c.registrations, nil
}

func TestUniformClient_GetIntegrationIDByName(t *testing.T) {

	tests := []struct {
		name                   string
		uniformClient          v2.UniformInterface
		integrationName        string
		expectedIntegrationID  string
		expectError            bool
		expectedErrorSubstring string
	}{
		{
			name: "one integration with name - should work",
			uniformClient: &mockUniformClient{
				registrations: []*models.Integration{
					createIntegration("5d64eb87c4ce3e23935758c418df9d980c16a3b1", "webhook-service"),
					createIntegration("e8039ff0b65c7e4d326a0473a18f04cabfefe747", "dynatrace-service"),
				},
			},
			integrationName:       "dynatrace-service",
			expectedIntegrationID: "e8039ff0b65c7e4d326a0473a18f04cabfefe747",
		},
		{
			name: "two integrations with name - should fail",
			uniformClient: &mockUniformClient{
				registrations: []*models.Integration{
					createIntegration("5d64eb87c4ce3e23935758c418df9d980c16a3b1", "webhook-service"),
					createIntegration("e8039ff0b65c7e4d326a0473a18f04cabfefe747", "dynatrace-service"),
					createIntegration("d73082b9c42aa147935fe2592a91eb5d2b224038", "dynatrace-service"),
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
					createIntegration("5d64eb87c4ce3e23935758c418df9d980c16a3b1", "webhook-service"),
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
			integrationID, err := c.GetIntegrationIDByName(context.Background(), tt.integrationName)
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

func createIntegration(id string, name string) *models.Integration {
	return &models.Integration{
		ID:   id,
		Name: name,
	}
}

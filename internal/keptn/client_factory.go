package keptn

import (
	"fmt"
	"net/http"

	api "github.com/keptn/go-utils/pkg/api/utils"
)

// ClientFactoryInterface provides a factories for clients.
type ClientFactoryInterface interface {
	CreateEventClient() EventClientInterface
	CreateResourceClient() ResourceClientInterface
	CreateServiceClient() ServiceClientInterface
	CreateUniformClient() UniformClientInterface
}

// ClientFactory is an implementation of ClientFactoryInterface.
type ClientFactory struct {
	apiSet *api.InternalAPISet
}

// NewClientFactory creates a new ClientFactory.
func NewClientFactory() (*ClientFactory, error) {
	internalAPISet, err := api.NewInternal(&http.Client{}, GetInClusterAPIMappings())
	if err != nil {
		return nil, fmt.Errorf("could not create internal Keptn API set: %w", err)
	}

	return &ClientFactory{apiSet: internalAPISet}, nil
}

// CreateEventClient creates an EventClientInterface.
func (c *ClientFactory) CreateEventClient() EventClientInterface {
	return NewEventClient(c.apiSet.EventsV1())
}

// CreateResourceClient creates a ResourceClientInterface.
func (c *ClientFactory) CreateResourceClient() ResourceClientInterface {
	return NewResourceClient(c.apiSet.ResourcesV1())
}

// CreateServiceClient creates a ServiceClientInterface.
func (c *ClientFactory) CreateServiceClient() ServiceClientInterface {
	return NewServiceClient(
		c.apiSet.ServicesV1(),
		c.apiSet.APIV1())
}

// CreateUniformClient creates a UniformClientInterface.
func (c *ClientFactory) CreateUniformClient() UniformClientInterface {
	return NewUniformClient(c.apiSet.UniformV1())
}

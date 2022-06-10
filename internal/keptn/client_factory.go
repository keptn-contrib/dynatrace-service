package keptn

import (
	"fmt"
	"net/http"

	v2 "github.com/keptn/keptn/cp-common/v2"
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
	apiSet *v2.InternalAPISet
}

// NewClientFactory creates a new ClientFactory.
func NewClientFactory() (*ClientFactory, error) {
	internalAPISet, err := v2.NewInternal(&http.Client{}, GetV2InClusterAPIMappings())
	if err != nil {
		return nil, fmt.Errorf("could not create internal Keptn API set: %w", err)
	}

	return &ClientFactory{apiSet: internalAPISet}, nil
}

// CreateEventClient creates an EventClientInterface.
func (c *ClientFactory) CreateEventClient() EventClientInterface {
	return NewEventClient(c.apiSet.Events())
}

// CreateResourceClient creates a ResourceClientInterface.
func (c *ClientFactory) CreateResourceClient() ResourceClientInterface {
	return NewResourceClient(c.apiSet.Resources())
}

// CreateServiceClient creates a ServiceClientInterface.
func (c *ClientFactory) CreateServiceClient() ServiceClientInterface {
	return NewServiceClient(
		c.apiSet.Services(),
		c.apiSet.API())
}

// CreateUniformClient creates a UniformClientInterface.
func (c *ClientFactory) CreateUniformClient() UniformClientInterface {
	return NewUniformClient(c.apiSet.Uniform())
}

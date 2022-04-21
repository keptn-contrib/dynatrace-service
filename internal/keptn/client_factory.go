package keptn

import (
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
}

// NewClientFactory creates a new ClientFactory.
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// CreateEventClient creates an EventClientInterface.
func (c *ClientFactory) CreateEventClient() EventClientInterface {
	return NewEventClient(api.NewEventHandler(GetDatastoreURL()))
}

// CreateResourceClient creates a ResourceClientInterface.
func (c *ClientFactory) CreateResourceClient() ResourceClientInterface {
	return NewResourceClient(api.NewResourceHandler(GetConfigurationServiceURL()))
}

// CreateServiceClient creates a ServiceClientInterface.
func (c *ClientFactory) CreateServiceClient() ServiceClientInterface {
	return NewServiceClient(
		api.NewServiceHandler(GetShipyardControllerURL()),
		api.NewAuthenticatedAPIHandler(GetShipyardControllerURL(), "", "", &http.Client{}, "http"))
}

// CreateUniformClient creates a UniformClientInterface.
func (c *ClientFactory) CreateUniformClient() UniformClientInterface {
	return NewUniformClient(api.NewUniformHandler(GetShipyardControllerURL()))
}

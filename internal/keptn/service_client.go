package keptn

import (
	"context"
	"fmt"

	apimodels "github.com/keptn/go-utils/pkg/api/models"
	v2 "github.com/keptn/go-utils/pkg/api/utils/v2"
)

// ServiceClientInterface provides access to Keptn services.
type ServiceClientInterface interface {
	// GetServiceNames gets the names of the services in the specified project and stage or returns an error.
	GetServiceNames(ctx context.Context, project string, stage string) ([]string, error)

	// CreateServiceInProject creates a service in all stages of the specified project or returns an error.
	CreateServiceInProject(ctx context.Context, project string, service string) error
}

// ServiceClient is an implementation of ServiceClientInterface using api.ServicesV1Interface and APIClientInterface.
type ServiceClient struct {
	servicesClient v2.ServicesInterface
	apiClient      v2.APIInterface
}

// NewServiceClient creates a new ServiceClient using the specified clients.
func NewServiceClient(servicesClient v2.ServicesInterface, apiClient v2.APIInterface) *ServiceClient {
	return &ServiceClient{
		servicesClient: servicesClient,
		apiClient:      apiClient,
	}
}

// GetServiceNames gets the names of the services in the specified project and stage or returns an error.
func (c *ServiceClient) GetServiceNames(ctx context.Context, project string, stage string) ([]string, error) {
	services, err := c.servicesClient.GetAllServices(ctx, project, stage, v2.ServicesGetAllServicesOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not fetch services of Keptn project %s at stage %s: %s", project, stage, err.Error())
	}

	if services == nil {
		return nil, nil
	}

	serviceNames := make([]string, len(services))
	for i, service := range services {
		serviceNames[i] = service.ServiceName
	}

	return serviceNames, nil
}

// CreateServiceInProject creates a service in all stages of the specified project or returns an error.
func (c *ServiceClient) CreateServiceInProject(ctx context.Context, project string, service string) error {
	serviceModel := apimodels.CreateService{
		ServiceName: &service,
	}

	_, err := c.apiClient.CreateService(ctx, project, serviceModel, v2.APICreateServiceOptions{})
	if err != nil {
		return fmt.Errorf("could not create service: %w", err.ToError())
	}
	return nil
}

package keptn

import (
	"fmt"

	apimodels "github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
)

// ServiceClientInterface provides access to Keptn services.
type ServiceClientInterface interface {
	// GetServiceNames gets the names of the services in the specified project and stage or returns an error.
	GetServiceNames(project string, stage string) ([]string, error)

	// CreateServiceInProject creates a service in all stages of the specified project or returns an error.
	CreateServiceInProject(project string, service string) error
}

// ServiceClient is an implementation of ServiceClientInterface using api.ServicesV1Interface and APIClientInterface.
type ServiceClient struct {
	servicesClient api.ServicesV1Interface
	apiClient      api.APIV1Interface
}

// NewServiceClient creates a new ServiceClient using the specified clients.
func NewServiceClient(servicesClient api.ServicesV1Interface, apiClient api.APIV1Interface) *ServiceClient {
	return &ServiceClient{
		servicesClient: servicesClient,
		apiClient:      apiClient,
	}
}

// GetServiceNames gets the names of the services in the specified project and stage or returns an error.
func (c *ServiceClient) GetServiceNames(project string, stage string) ([]string, error) {
	services, err := c.servicesClient.GetAllServices(project, stage)
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
func (c *ServiceClient) CreateServiceInProject(project string, service string) error {
	serviceModel := apimodels.CreateService{
		ServiceName: &service,
	}

	_, err := c.apiClient.CreateService(project, serviceModel)
	if err != nil {
		return fmt.Errorf("could not create service: %w", err.ToError())
	}
	return nil
}

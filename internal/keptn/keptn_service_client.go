package keptn

import (
	"errors"
	"fmt"

	apimodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
)

type ServiceClientInterface interface {
	GetServiceNames(project string, stage string) ([]string, error)
	CreateServiceInProject(project string, service string) error
}

type ServiceClient struct {
	servicesClient keptnapi.ServicesV1Interface
	apiClient      keptnapi.APIV1Interface
}

func NewServiceClient(client keptnapi.ServicesV1Interface, apiClient keptnapi.APIV1Interface) *ServiceClient {
	return &ServiceClient{
		servicesClient: client,
		apiClient:      apiClient,
	}
}

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

func (c *ServiceClient) CreateServiceInProject(project string, service string) error {
	_, keptnAPIErr := c.apiClient.CreateService(project, apimodels.CreateService{
		ServiceName: &service,
	})

	return errors.New(keptnAPIErr.GetMessage())
}

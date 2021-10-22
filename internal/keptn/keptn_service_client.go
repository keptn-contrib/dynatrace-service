package keptn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/rest"

	apimodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
)

type ServiceClientInterface interface {
	GetServiceNames(project string, stage string) ([]string, error)
	CreateServiceInProject(project string, service string) error
}

type ServiceClient struct {
	client    *keptnapi.ServiceHandler
	apiClient APIClientInterface
}

func NewDefaultServiceClient() *ServiceClient {
	return NewServiceClient(
		keptnapi.NewServiceHandler(common.GetShipyardControllerURL()),
		&http.Client{})
}

func NewServiceClient(client *keptnapi.ServiceHandler, httpClient *http.Client) *ServiceClient {
	return &ServiceClient{
		client: client,
		apiClient: NewAPIClient(
			rest.NewDefaultClient(
				httpClient,
				common.GetShipyardControllerURL())),
	}
}

func (c *ServiceClient) GetServiceNames(project string, stage string) ([]string, error) {
	services, err := c.client.GetAllServices(project, stage)
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
	serviceModel := &apimodels.CreateService{
		ServiceName: &service,
	}
	reqBody, err := json.Marshal(serviceModel)
	if err != nil {
		return fmt.Errorf("could not marshal service payload: %s", err.Error())
	}

	_, err = c.apiClient.Post(getServicePathFor(project), reqBody)
	return err
}

func getServicePathFor(project string) string {
	return "/v1/project/" + project + "/service"
}

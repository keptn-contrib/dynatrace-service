package keptn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	apimodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	"io/ioutil"
	"net/http"
)

type ServiceClientInterface interface {
	GetServiceNames(project string, stage string) ([]string, error)
	CreateServiceInProject(project string, service string) error
}

type ServiceClient struct {
	client     *keptnapi.ServiceHandler
	httpClient *http.Client
}

func NewDefaultServiceClient() *ServiceClient {
	return NewServiceClient(
		keptnapi.NewServiceHandler(common.GetShipyardControllerURL()),
		&http.Client{})
}

func NewServiceClient(client *keptnapi.ServiceHandler, httpClient *http.Client) *ServiceClient {
	return &ServiceClient{
		client:     client,
		httpClient: httpClient,
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

	// TODO 2021-09-08: extract the http client and maybe consolidate with dynatrace (http) client
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/project/%s/service", common.GetShipyardControllerURL(), project), bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

package keptn

import (
	"errors"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	api "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"gopkg.in/yaml.v2"
)

type ResourceClientInterface interface {
	GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error)
}

type ResourceClient struct {
	handler *api.ResourceHandler
}

func NewConfigResourceClient() *ResourceClient {
	return &ResourceClient{
		handler: api.NewResourceHandler(
			common.GetConfigurationServiceURL()),
	}
}

func (rc *ResourceClient) GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error) {
	resource, err := rc.handler.GetServiceResource(project, stage, service, "slo.yaml")
	if err != nil || resource.ResourceContent == "" {
		return nil, errors.New("No SLO file available for service " + service + " in stage " + stage)
	}

	slos := &keptn.ServiceLevelObjectives{}
	err = yaml.Unmarshal([]byte(resource.ResourceContent), slos)
	if err != nil {
		return nil, errors.New("invalid SLO file format")
	}

	return slos, nil
}

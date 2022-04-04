package keptn

import (
	"errors"
	"fmt"

	keptn "github.com/keptn/go-utils/pkg/lib"
	"gopkg.in/yaml.v2"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type SLOResourceReaderInterface interface {
	// GetSLOs gets the SLOs stored for exactly the specified project, stage and service.
	GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error)
}
type SLIAndSLOResourceWriterInterface interface {
	// UploadSLIs uploads the SLIs for the specified project, stage and service.
	UploadSLIs(project string, stage string, service string, slis *dynatrace.SLI) error

	// UploadSLOs uploads the SLOs for the specified project, stage and service.
	UploadSLOs(project string, stage string, service string, slos *keptn.ServiceLevelObjectives) error
}
type ResourceClientInterface interface {
	SLOResourceReaderInterface
	SLIAndSLOResourceWriterInterface
}

type DynatraceConfigResourceClientInterface interface {
	// GetDynatraceConfig gets the Dynatrace config for the specified project, stage and service, checking first on the service, then stage and then project level.
	GetDynatraceConfig(project string, stage string, service string) (string, error)
}

const sloFilename = "slo.yaml"
const sliFilename = "dynatrace/sli.yaml"
const configFilename = "dynatrace/dynatrace.conf.yaml"

// ResourceClient is the default implementation for the *ResourceClientInterfaces using a ConfigResourceClientInterface
type ResourceClient struct {
	client ConfigResourceClientInterface
}

// NewResourceClient creates a new ResourceClient with a Keptn resource handler for the configuration service
func NewResourceClient(client ConfigResourceClientInterface) *ResourceClient {
	return &ResourceClient{
		client: client,
	}
}

func (rc *ResourceClient) GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error) {
	resource, err := rc.client.GetServiceResource(project, stage, service, sloFilename)
	if err != nil {
		return nil, err
	}

	slos := &keptn.ServiceLevelObjectives{}
	err = yaml.Unmarshal([]byte(resource), slos)
	if err != nil {
		return nil, errors.New("invalid SLO file format")
	}

	return slos, nil
}

func (rc *ResourceClient) UploadSLOs(project string, stage string, service string, slos *keptn.ServiceLevelObjectives) error {
	// and now we save it back to Keptn
	yamlAsByteArray, err := yaml.Marshal(slos)
	if err != nil {
		return fmt.Errorf("could not convert SLOs to YAML: %s", err)
	}

	return rc.client.UploadResource(yamlAsByteArray, sloFilename, project, stage, service)
}

func (rc *ResourceClient) UploadSLIs(project string, stage string, service string, slis *dynatrace.SLI) error {
	yamlAsByteArray, err := yaml.Marshal(slis)
	if err != nil {
		return fmt.Errorf("could not convert SLIs to YAML: %s", err)
	}

	return rc.client.UploadResource(yamlAsByteArray, sliFilename, project, stage, service)
}

func (rc *ResourceClient) GetDynatraceConfig(project string, stage string, service string) (string, error) {
	return rc.client.GetResource(project, stage, service, configFilename)
}

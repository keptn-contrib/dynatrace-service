package keptn

import (
	"errors"
	"fmt"

	keptn "github.com/keptn/go-utils/pkg/lib"
	"gopkg.in/yaml.v2"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type SLOResourceReaderInterface interface {
	GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error)
}
type SLIAndSLOResourceWriterInterface interface {
	UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error
	UploadSLOs(project string, stage string, service string, dashboardSLOs *keptn.ServiceLevelObjectives) error
}
type ResourceClientInterface interface {
	SLOResourceReaderInterface
	SLIAndSLOResourceWriterInterface
}

type DynatraceConfigResourceClientInterface interface {
	GetDynatraceConfig(project string, stage string, service string) (string, error)
}

const sloFilename = "slo.yaml"
const sliFilename = "dynatrace/sli.yaml"
const dashboardFilename = "dynatrace/dashboard.json"
const configFilename = "dynatrace/dynatrace.conf.yaml"

// ResourceClient is the default implementation for the *ResourceClientInterfaces using a ConfigResourceClientInterface
type ResourceClient struct {
	client ConfigResourceClientInterface
}

// NewDefaultResourceClient creates a new ResourceClient with a default Keptn resource handler for the configuration service
func NewDefaultResourceClient() *ResourceClient {
	return NewResourceClient(
		NewDefaultConfigResourceClient())
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

func (rc *ResourceClient) UploadSLOs(project string, stage string, service string, dashboardSLOs *keptn.ServiceLevelObjectives) error {
	// and now we save it back to Keptn
	yamlAsByteArray, err := yaml.Marshal(dashboardSLOs)
	if err != nil {
		return fmt.Errorf("could not convert SLOs to YAML: %s", err)
	}

	return rc.client.UploadResource(yamlAsByteArray, sloFilename, project, stage, service)
}

func (rc *ResourceClient) GetDashboard(project string, stage string, service string) (string, error) {
	return rc.client.GetServiceResource(project, stage, service, dashboardFilename)
}

func (rc *ResourceClient) UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error {
	yamlAsByteArray, err := yaml.Marshal(sli)
	if err != nil {
		return fmt.Errorf("could not convert dashboardSLI to YAML: %s", err)
	}

	return rc.client.UploadResource(yamlAsByteArray, sliFilename, project, stage, service)
}

func (rc *ResourceClient) GetDynatraceConfig(project string, stage string, service string) (string, error) {
	return rc.client.GetResource(project, stage, service, configFilename)
}

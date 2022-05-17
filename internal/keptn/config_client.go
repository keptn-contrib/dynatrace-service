package keptn

import (
	"errors"
	"fmt"

	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnapi "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"gopkg.in/yaml.v2"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

// SLIAndSLOReaderInterface provides functionality for getting SLIs and SLOs.
type SLIAndSLOReaderInterface interface {
	// GetSLIs gets the SLIs stored for the specified project, stage and service.
	// First, the configuration of project-level is retrieved, which is then overridden by configuration on stage level, and then overridden by configuration on service level.
	GetSLIs(project string, stage string, service string) (map[string]string, error)

	// GetSLOs gets the SLOs stored for exactly the specified project, stage and service.
	GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error)
}

// SLIAndSLOWriterInterface provides functionality for uploading SLIs and SLOs.
type SLIAndSLOWriterInterface interface {
	// UploadSLIs uploads the SLIs for the specified project, stage and service.
	UploadSLIs(project string, stage string, service string, slis *dynatrace.SLI) error

	// UploadSLOs uploads the SLOs for the specified project, stage and service.
	UploadSLOs(project string, stage string, service string, slos *keptn.ServiceLevelObjectives) error
}

// SLOAndSLIClientInterface provides functionality for getting and uploading SLIs and SLOs.
type SLOAndSLIClientInterface interface {
	SLIAndSLOReaderInterface
	SLIAndSLOWriterInterface
}

// ShipyardReaderInterface provides functionality for getting a project's shipyard.
type ShipyardReaderInterface interface {
	// GetShipyard returns the shipyard definition of a project.
	GetShipyard(project string) (*keptnv2.Shipyard, error)
}

// DynatraceConfigReaderInterface provides functionality for getting a Dynatrace config.
type DynatraceConfigReaderInterface interface {
	// GetDynatraceConfig gets the Dynatrace config for the specified project, stage and service, checking first on the service, then stage and then project level.
	GetDynatraceConfig(project string, stage string, service string) (string, error)
}

const shipyardFilename = "shipyard.yaml"
const sloFilename = "slo.yaml"
const sliFilename = "dynatrace/sli.yaml"
const configFilename = "dynatrace/dynatrace.conf.yaml"

// ConfigClient is the default implementation for ResourceClientInterface using a ConfigResourceClientInterface.
type ConfigClient struct {
	client ResourceClientInterface
}

// NewConfigClient creates a new ConfigClient with a Keptn resource handler for the configuration service.
func NewConfigClient(client ResourceClientInterface) *ConfigClient {
	return &ConfigClient{
		client: client,
	}
}

// GetSLOs gets the SLOs stored for exactly the specified project, stage and service.
func (rc *ConfigClient) GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error) {
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

// UploadSLOs uploads the SLOs for the specified project, stage and service.
func (rc *ConfigClient) UploadSLOs(project string, stage string, service string, slos *keptn.ServiceLevelObjectives) error {
	yamlAsByteArray, err := yaml.Marshal(slos)
	if err != nil {
		return fmt.Errorf("could not convert SLOs to YAML: %s", err)
	}

	return rc.client.UploadResource(yamlAsByteArray, sloFilename, project, stage, service)
}

// UploadSLIs uploads the SLIs for the specified project, stage and service.
func (rc *ConfigClient) UploadSLIs(project string, stage string, service string, slis *dynatrace.SLI) error {
	yamlAsByteArray, err := yaml.Marshal(slis)
	if err != nil {
		return fmt.Errorf("could not convert SLIs to YAML: %s", err)
	}

	return rc.client.UploadResource(yamlAsByteArray, sliFilename, project, stage, service)
}

// GetDynatraceConfig gets the Dynatrace config for the specified project, stage and service, checking first on the service, then stage and then project level.
func (rc *ConfigClient) GetDynatraceConfig(project string, stage string, service string) (string, error) {
	return rc.client.GetResource(project, stage, service, configFilename)
}

// GetShipyard returns the shipyard definition of a project.
func (rc *ConfigClient) GetShipyard(project string) (*keptnv2.Shipyard, error) {
	shipyard, err := rc.getShipyard(project)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve shipyard for project %s: %w", project, err)
	}
	return shipyard, nil
}

func (rc *ConfigClient) getShipyard(project string) (*keptnv2.Shipyard, error) {
	shipyardResource, err := rc.client.GetProjectResource(project, shipyardFilename)
	if err != nil {
		return nil, err
	}

	shipyard := keptnv2.Shipyard{}
	err = yaml.Unmarshal([]byte(shipyardResource), &shipyard)
	if err != nil {
		return nil, err
	}

	return &shipyard, nil
}

type sliMap map[string]string

func (m *sliMap) insertOrUpdateMany(x map[string]string) {
	for key, value := range x {
		map[string]string(*m)[key] = value
	}
}

// GetSLIs gets the SLIs stored for the specified project, stage and service.
// First, the configuration of project-level is retrieved, which is then overridden by configuration on stage level, and then overridden by configuration on service level.
func (rc *ConfigClient) GetSLIs(project string, stage string, service string) (map[string]string, error) {
	slis := make(sliMap)

	// try to get SLI config from project
	if project != "" {
		projectSLIs, err := getSLIsFromResource(func() (string, error) { return rc.client.GetProjectResource(project, sliFilename) })
		if err != nil {
			return nil, err
		}

		slis.insertOrUpdateMany(projectSLIs)
	}

	// try to get SLI config from stage
	if project != "" && stage != "" {
		stageSLIs, err := getSLIsFromResource(func() (string, error) { return rc.client.GetStageResource(project, stage, sliFilename) })
		if err != nil {
			return nil, err
		}

		slis.insertOrUpdateMany(stageSLIs)
	}

	// try to get SLI config from service
	if project != "" && stage != "" && service != "" {
		serviceSLIs, err := getSLIsFromResource(func() (string, error) { return rc.client.GetServiceResource(project, stage, service, sliFilename) })
		if err != nil {
			return nil, err
		}

		slis.insertOrUpdateMany(serviceSLIs)
	}

	return slis, nil
}

type resourceGetterFunc func() (string, error)

// getSLIsFromResource uses the specified function to get a resource and returns the SLIs as a map.
// If is is not possible to get the resource for any other reason than it is not found, or it is not possible to unmarshal the file or it doesn't contain any indicators, an error is returned.
func getSLIsFromResource(resourceGetter resourceGetterFunc) (map[string]string, error) {
	resource, err := resourceGetter()
	if err != nil {
		var rnfErrorType *ResourceNotFoundError
		if errors.As(err, &rnfErrorType) {
			return nil, nil
		}

		return nil, err
	}

	return readSLIsFromResource(resource)
}

// readSLIsFromResource unmarshals a resource as a SLIConfig and returns the SLIs as a map.
// If it is not possible to unmarshal the file or it doesn't contain any indicators, an error is returned.
func readSLIsFromResource(resource string) (map[string]string, error) {
	sliConfig := keptnapi.SLIConfig{}
	err := yaml.Unmarshal([]byte(resource), &sliConfig)
	if err != nil {
		return nil, err
	}

	if len(sliConfig.Indicators) == 0 {
		return nil, errors.New("missing required field: indicators")
	}

	return sliConfig.Indicators, nil
}

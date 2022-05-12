package keptn

import (
	"errors"
	"fmt"

	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"gopkg.in/yaml.v2"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

// SLOReaderInterface provides functionality for getting SLOs.
type SLOReaderInterface interface {
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

// SLOAndSLIClientInterface provides functionality for getting SLOs and uploading SLIs and SLOs.
type SLOAndSLIClientInterface interface {
	SLOReaderInterface
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

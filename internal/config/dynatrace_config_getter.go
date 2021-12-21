package config

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//go:generate moq --skip-ensure -pkg adapter_mock -out ./mock/dynatrace_config_mock.go . DynatraceConfigProvider
type DynatraceConfigProvider interface {
	GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfig, error)
}

type DynatraceConfigGetter struct {
	resourceClient keptn.DynatraceConfigResourceClientInterface
}

func NewDynatraceConfigGetter(client keptn.DynatraceConfigResourceClientInterface) *DynatraceConfigGetter {
	return &DynatraceConfigGetter{
		resourceClient: client,
	}
}

// GetDynatraceConfig loads the dynatrace.conf.yaml from the GIT repo
func (d *DynatraceConfigGetter) GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfig, error) {

	fileContent, err := d.resourceClient.GetDynatraceConfig(event.GetProject(), event.GetStage(), event.GetService())
	if err != nil {
		return nil, err
	}

	if len(fileContent) > 0 {

		// replace the placeholders
		log.WithField("fileContent", fileContent).Debug("Original contents of configuration file")
		fileContent = common.ReplaceKeptnPlaceholders(fileContent, event)
		log.WithField("fileContent", fileContent).Debug("Contents of configuration file after replacements")
	}

	// unmarshal the file
	dynatraceConfig, err := parseDynatraceConfigYAML(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dynatrace config file found for service %s in stage %s in project %s: %s", event.GetService(), event.GetStage(), event.GetProject(), err.Error())
	}

	return dynatraceConfig, nil
}

func parseDynatraceConfigYAML(input string) (*DynatraceConfig, error) {
	dynatraceConfig := NewDynatraceConfigWithDefaults()
	err := yaml.Unmarshal([]byte(input), dynatraceConfig)
	if err != nil {
		return nil, err
	}

	return dynatraceConfig, nil
}

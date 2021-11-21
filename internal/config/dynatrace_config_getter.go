package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//go:generate moq --skip-ensure -pkg adapter_mock -out ./mock/dynatrace_config_mock.go . DynatraceConfigGetterInterface
type DynatraceConfigProvider interface {
	GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfigFile, error)
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
func (d *DynatraceConfigGetter) GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfigFile, error) {

	fileContent, err := d.resourceClient.GetDynatraceConfig(event.GetProject(), event.GetStage(), event.GetService())
	if err != nil {
		return nil, err
	}

	if len(fileContent) > 0 {

		// replace the placeholders
		log.WithField("fileContent", fileContent).Debug("Original contents of configuration file")
		fileContent = replaceKeptnPlaceholders(fileContent, event)
		log.WithField("fileContent", fileContent).Debug("Contents of configuration file after replacements")
	}

	// unmarshal the file
	dynatraceConfFile, err := parseDynatraceConfigFile([]byte(fileContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse dynatrace config file found for service %s in stage %s in project %s: %s", event.GetService(), event.GetStage(), event.GetProject(), err.Error())
	}

	return dynatraceConfFile, nil
}

//
// replaces $ placeholders with actual values
// $CONTEXT, $EVENT, $SOURCE
// $PROJECT, $STAGE, $SERVICE, $DEPLOYMENT
// $TESTSTRATEGY
// $LABEL.XXXX  -> will replace that with a label called XXXX
// $ENV.XXXX    -> will replace that with an env variable called XXXX
// $SECRET.YYYY -> will replace that with the k8s secret called YYYY
//
func replaceKeptnPlaceholders(input string, event adapter.EventContentAdapter) string {
	result := input

	// first we do the regular keptn values
	result = strings.Replace(result, "$CONTEXT", event.GetShKeptnContext(), -1)
	result = strings.Replace(result, "$EVENT", event.GetEvent(), -1)
	result = strings.Replace(result, "$SOURCE", event.GetSource(), -1)
	result = strings.Replace(result, "$PROJECT", event.GetProject(), -1)
	result = strings.Replace(result, "$STAGE", event.GetStage(), -1)
	result = strings.Replace(result, "$SERVICE", event.GetService(), -1)
	result = strings.Replace(result, "$DEPLOYMENT", event.GetDeployment(), -1)
	result = strings.Replace(result, "$TESTSTRATEGY", event.GetTestStrategy(), -1)

	// now we do the labels
	for key, value := range event.GetLabels() {
		result = strings.Replace(result, "$LABEL."+key, value, -1)
	}

	// now we do all environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		result = strings.Replace(result, "$ENV."+pair[0], pair[1], -1)
	}

	// TODO: iterate through k8s secrets!

	return result
}

func parseDynatraceConfigFile(input []byte) (*DynatraceConfigFile, error) {
	dynatraceConfFile := &DynatraceConfigFile{}
	err := yaml.Unmarshal(input, dynatraceConfFile)

	if err != nil {
		return nil, err
	}

	return dynatraceConfFile, nil
}

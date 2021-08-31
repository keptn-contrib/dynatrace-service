package config

import (
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//go:generate moq --skip-ensure -pkg adapter_mock -out ./mock/dynatrace_config_mock.go . DynatraceConfigGetterInterface
type DynatraceConfigGetterInterface interface {
	GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfigFile, error)
}

type DynatraceConfigGetter struct {
	resourceClient keptn.ResourceClientInterface
}

func NewDynatraceConfigGetter(client keptn.ResourceClientInterface) *DynatraceConfigGetter {
	return &DynatraceConfigGetter{
		resourceClient: client,
	}
}

// GetDynatraceConfig loads the dynatrace.conf.yaml from the GIT repo
func (d *DynatraceConfigGetter) GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfigFile, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	var fileContent string
	if common.RunLocal {
		localFileContent, err := ioutil.ReadFile(DynatraceConfigFilenameLOCAL)
		if err != nil {

			log.WithError(err).WithFields(log.Fields{
				"dynatraceConfigFilename": DynatraceConfigFilenameLOCAL,
				"service":                 event.GetService(),
				"stage":                   event.GetStage(),
				"project":                 event.GetProject(),
			}).Info("No configuration file was found LOCALLY")
			return nil, nil
		}
		log.WithField("dynatraceConfigFilename", DynatraceConfigFilenameLOCAL).Info("Loaded LOCAL configuration file")
		fileContent = string(localFileContent)
	} else {
		var err error
		fileContent, err = d.resourceClient.GetResource(event, DynatraceConfigFilename)
		if err != nil {
			return nil, err
		}
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
		errMsg := fmt.Sprintf("failed to parse %s file found for service %s in stage %s in project %s: %s",
			DynatraceConfigFilename, event.GetService(), event.GetStage(), event.GetProject(), err.Error())
		return nil, errors.New(errMsg)
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

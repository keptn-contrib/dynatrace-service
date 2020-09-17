package config

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptnutils "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"io/ioutil"
	"os"
	"strings"
)

const DynatraceConfigFilename = "dynatrace/dynatrace.conf.yaml"
const DynatraceConfigFilenameLOCAL = "dynatrace/_dynatrace.conf.yaml"

type DtTag struct {
	Context string `json:"context" yaml:"context"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value",omitempty yaml:"value",omitempty`
}

type DtTagRule struct {
	MeTypes []string `json:"meTypes" yaml:"meTypes"`
	Tags    []DtTag  `json:"tags" yaml:"tags"`
}

type DtAttachRules struct {
	TagRule []DtTagRule `json:"tagRule" yaml:"tagRule"`
}

/**
 * Defines the Dynatrace Configuration File structure!
 */
type DynatraceConfigFile struct {
	SpecVersion string         `json:"spec_version" yaml:"spec_version"`
	DtCreds     string         `json:"dtCreds",omitempty yaml:"dtCreds",omitempty`
	AttachRules *DtAttachRules `json:"attachRules",omitempty yaml:"attachRules",omitempty`
}

//
// Loads dynatrace.conf for the current service
//
func GetDynatraceConfig(event adapter.EventAdapter, logger keptn.LoggerInterface) (*DynatraceConfigFile, error) {

	logger.Info("Loading dynatrace.conf.yaml")
	// if we run in a runlocal mode we are just getting the file from the local disk
	var fileContent string
	if common.RunLocal {
		localFileContent, err := ioutil.ReadFile(DynatraceConfigFilenameLOCAL)
		if err != nil {
			logMessage := fmt.Sprintf("No %s file found LOCALLY for service %s in stage %s in project %s",
				DynatraceConfigFilenameLOCAL, event.GetService(), event.GetStage(), event.GetProject())
			logger.Info(logMessage)
			return nil, nil
		}
		logger.Info("Loaded LOCAL file " + DynatraceConfigFilenameLOCAL)
		fileContent = string(localFileContent)
	} else {
		var err error
		fileContent, err = getDynatraceConfigResource(event)
		if err != nil {
			return nil, err
		}
	}

	// replace the placeholders
	logger.Debug("Content of dynatrace.conf.yaml: " + fileContent)
	fileContent = replaceKeptnPlaceholders(fileContent, event)
	logger.Debug("After replacements: " + fileContent)

	// unmarshal the file
	dynatraceConfFile, err := parseDynatraceConfigFile([]byte(fileContent))
	if err != nil {
		errMsg := fmt.Sprintf("failed to parse %s file found for service %s in stage %s in project %s: %s",
			DynatraceConfigFilename, event.GetService(), event.GetStage(), event.GetProject(), err.Error())
		return nil, errors.New(errMsg)
	}

	return dynatraceConfFile, nil
}

func getDynatraceConfigResource(event adapter.EventAdapter) (string, error) {

	resourceHandler := keptnutils.NewResourceHandler(common.GetConfigurationServiceURL())

	// Lets search on SERVICE-LEVEL
	keptnResourceContent, err := resourceHandler.GetServiceResource(event.GetProject(), event.GetStage(), event.GetService(), DynatraceConfigFilename)
	if err == keptnutils.ResourceNotFoundError {
		// Lets search on STAGE-LEVEL
		keptnResourceContent, err = resourceHandler.GetStageResource(event.GetProject(), event.GetStage(), DynatraceConfigFilename)
		if err == keptnutils.ResourceNotFoundError {
			// Lets search on PROJECT-LEVEL
			keptnResourceContent, err = resourceHandler.GetProjectResource(event.GetProject(), DynatraceConfigFilename)
		}
	}

	if err == keptnutils.ResourceNotFoundError {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return keptnResourceContent.ResourceContent, nil
}

func parseDynatraceConfigFile(input []byte) (*DynatraceConfigFile, error) {
	dynatraceConfFile := &DynatraceConfigFile{}
	err := yaml.Unmarshal(input, dynatraceConfFile)

	if err != nil {
		return nil, err
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
func replaceKeptnPlaceholders(input string, event adapter.EventAdapter) string {
	result := input

	// first we do the regular keptn values
	result = strings.Replace(result, "$CONTEXT", event.GetContext(), -1)
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

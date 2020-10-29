package config

import (
	"errors"
	"fmt"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptnutils "github.com/keptn/go-utils/pkg/api/utils"
	"gopkg.in/yaml.v2"
)

// DynatraceConfigFilename is the resource path for the dynatrace.conf.yaml
const DynatraceConfigFilename = "dynatrace/dynatrace.conf.yaml"

// DynatraceConfigFilenameLOCAL is the local test resource path for the dynatrace.conf.yaml
const DynatraceConfigFilenameLOCAL = "dynatrace/_dynatrace.conf.yaml"

// DtTag defines a Dynatrace configuration structure
type DtTag struct {
	Context string `json:"context" yaml:"context"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
}

// DtTagRule defines a Dynatrace configuration structure
type DtTagRule struct {
	MeTypes []string `json:"meTypes" yaml:"meTypes"`
	Tags    []DtTag  `json:"tags" yaml:"tags"`
}

// DtAttachRules defines a Dynatrace configuration structure
type DtAttachRules struct {
	TagRule []DtTagRule `json:"tagRule" yaml:"tagRule"`
}

// DynatraceConfigFile defines the Dynatrace configuration structure
type DynatraceConfigFile struct {
	SpecVersion string         `json:"spec_version" yaml:"spec_version"`
	DtCreds     string         `json:"dtCreds,omitempty" yaml:"dtCreds,omitempty"`
	AttachRules *DtAttachRules `json:"attachRules,omitempty" yaml:"attachRules,omitempty"`
}

// GetDynatraceConfig loads the dynatrace.conf.yaml from the GIT repo
func GetDynatraceConfig(event adapter.EventContentAdapter, logger keptncommon.LoggerInterface) (*DynatraceConfigFile, error) {

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
		fileContent, err = getDynatraceConfigResource(event, logger)
		if err != nil {
			return nil, err
		}
	}

	if len(fileContent) > 0 {

		// replace the placeholders
		logger.Debug("Input content of dynatrace.conf.yaml: " + fileContent)
		fileContent = replaceKeptnPlaceholders(fileContent, event)
		logger.Debug("Content of dyantrace.conf.yaml after replacements: " + fileContent)
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

func getDynatraceConfigResource(event adapter.EventContentAdapter, logger keptncommon.LoggerInterface) (string, error) {

	resourceHandler := keptnutils.NewResourceHandler(common.GetConfigurationServiceURL())

	// Lets search on SERVICE-LEVEL
	if len(event.GetProject()) > 0 && len(event.GetStage()) > 0 && len(event.GetService()) > 0 {
		keptnResourceContent, err := resourceHandler.GetServiceResource(event.GetProject(), event.GetStage(), event.GetService(), DynatraceConfigFilename)
		if err == keptnutils.ResourceNotFoundError {
			logger.Info(fmt.Sprintf("No dynatrace.conf.yaml available in project %s at stage %s for service %s", event.GetProject(), event.GetStage(), event.GetService()))
		} else if err != nil {
			return "", fmt.Errorf("failed to retrieve dynatrace.conf.yaml in project %s at stage %s for service %s: %v", event.GetProject(), event.GetStage(), event.GetService(), err)
		} else {
			logger.Info(fmt.Sprintf("Found dynatrace.conf.yaml in project %s at stage %s for service %s", event.GetProject(), event.GetStage(), event.GetService()))
			return keptnResourceContent.ResourceContent, nil
		}
	}

	if len(event.GetProject()) > 0 && len(event.GetStage()) > 0 {
		keptnResourceContent, err := resourceHandler.GetStageResource(event.GetProject(), event.GetStage(), DynatraceConfigFilename)
		if err == keptnutils.ResourceNotFoundError {
			logger.Info(fmt.Sprintf("No dynatrace.conf.yaml available in project %s at stage %s", event.GetProject(), event.GetStage()))
		} else if err != nil {
			return "", fmt.Errorf("failed to retrieve dynatrace.conf.yaml in project %s at stage %s: %v", event.GetProject(), event.GetStage(), err)
		} else {
			logger.Info(fmt.Sprintf("Found dynatrace.conf.yaml in project %s at stage %s", event.GetProject(), event.GetStage()))
			return keptnResourceContent.ResourceContent, nil
		}
	}

	if len(event.GetProject()) > 0 {
		keptnResourceContent, err := resourceHandler.GetProjectResource(event.GetProject(), DynatraceConfigFilename)
		if err == keptnutils.ResourceNotFoundError {
			logger.Info(fmt.Sprintf("No dynatrace.conf.yaml available in project %s", event.GetProject()))
		} else if err != nil {
			return "", fmt.Errorf("failed to retrieve dynatrace.conf.yaml in project %s: %v", event.GetProject(), err)
		} else {
			logger.Info(fmt.Sprintf("Found dynatrace.conf.yaml in project %s", event.GetProject()))
			return keptnResourceContent.ResourceContent, nil
		}
	}

	logger.Info("No dynatrace.conf.yaml found")
	return "", nil
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

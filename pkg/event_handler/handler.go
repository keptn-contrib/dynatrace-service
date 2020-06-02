package event_handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/ghodss/yaml"
	keptn "github.com/keptn/go-utils/pkg/lib"

	utils "github.com/keptn/go-utils/pkg/api/utils"
)

const DynatraceConfigFilename = "dynatrace/dynatrace.conf.yaml"
const DynatraceConfigFilenameLOCAL = "dynatrace/_dynatrace.conf.yaml"

/**
 * Defines the Dynatrace Configuration File structure!
 */
type DynatraceConfigFile struct {
	SpecVersion string         `json:"spec_version" yaml:"spec_version"`
	DtCreds     string         `json:"dtCreds",omitempty yaml:"dtCreds",omitempty`
	AttachRules *dtAttachRules `json:"attachRules",omitempty yaml:"attachRules",omitempty`
}

type baseKeptnEvent struct {
	context string
	source  string
	event   string

	project            string
	stage              string
	service            string
	deployment         string
	testStrategy       string
	deploymentStrategy string

	image string
	tag   string

	labels map[string]string
}

type DynatraceEventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event, logger *keptn.Logger) (DynatraceEventHandler, error) {
	logger.Debug("Received event: " + event.Type())
	switch event.Type() {
	case keptn.ConfigureMonitoringEventType:
		return &ConfigureMonitoringEventHandler{Logger: logger, Event: event}, nil
	case keptn.InternalProjectCreateEventType:
		return &CreateProjectEventHandler{Logger: logger, Event: event}, nil
	case keptn.ProblemEventType:
		return &ProblemEventHandler{Logger: logger, Event: event}, nil
	case keptn.ActionTriggeredEventType:
		return &ActionHandler{Logger: logger, Event: event}, nil
	case keptn.ActionFinishedEventType:
		return &ActionHandler{Logger: logger, Event: event}, nil
	default:
		return &CDEventHandler{Logger: logger, Event: event}, nil
	}
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
func replaceKeptnPlaceholders(input string, keptnEvent *baseKeptnEvent) string {
	result := input

	// first we do the regular keptn values
	result = strings.Replace(result, "$CONTEXT", keptnEvent.context, -1)
	result = strings.Replace(result, "$EVENT", keptnEvent.event, -1)
	result = strings.Replace(result, "$SOURCE", keptnEvent.source, -1)
	result = strings.Replace(result, "$PROJECT", keptnEvent.project, -1)
	result = strings.Replace(result, "$STAGE", keptnEvent.stage, -1)
	result = strings.Replace(result, "$SERVICE", keptnEvent.service, -1)
	result = strings.Replace(result, "$DEPLOYMENT", keptnEvent.deployment, -1)
	result = strings.Replace(result, "$TESTSTRATEGY", keptnEvent.testStrategy, -1)

	// now we do the labels
	for key, value := range keptnEvent.labels {
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

//
// Loads dynatrace.conf for the current service
//
func getDynatraceConfig(keptnEvent *baseKeptnEvent, logger *keptn.Logger) (*DynatraceConfigFile, error) {

	logger.Info("Loading dynatrace.conf.yaml")
	// if we run in a runlocal mode we are just getting the file from the local disk
	var fileContent string
	if common.RunLocal {
		localFileContent, err := ioutil.ReadFile(DynatraceConfigFilenameLOCAL)
		if err != nil {
			logMessage := fmt.Sprintf("No %s file found LOCALLY for service %s in stage %s in project %s", DynatraceConfigFilenameLOCAL, keptnEvent.service, keptnEvent.stage, keptnEvent.project)
			logger.Info(logMessage)
			return nil, nil
		}
		logger.Info("Loaded LOCAL file " + DynatraceConfigFilenameLOCAL)
		fileContent = string(localFileContent)
	} else {
		resourceHandler := utils.NewResourceHandler(common.GetConfigurationServiceURL())

		// Lets search on SERVICE-LEVEL
		keptnResourceContent, err := resourceHandler.GetServiceResource(keptnEvent.project, keptnEvent.stage, keptnEvent.service, DynatraceConfigFilename)
		if err != nil || keptnResourceContent == nil || keptnResourceContent.ResourceContent == "" {
			// Lets search on STAGE-LEVEL
			keptnResourceContent, err = resourceHandler.GetStageResource(keptnEvent.project, keptnEvent.stage, DynatraceConfigFilename)
			if err != nil || keptnResourceContent == nil || keptnResourceContent.ResourceContent == "" {
				// Lets search on PROJECT-LEVEL
				keptnResourceContent, err = resourceHandler.GetProjectResource(keptnEvent.project, DynatraceConfigFilename)
				if err != nil || keptnResourceContent == nil || keptnResourceContent.ResourceContent == "" {
					logger.Debug(fmt.Sprintf("No Keptn Resource found: %s/%s/%s/%s - %s", keptnEvent.project, keptnEvent.stage, keptnEvent.service, DynatraceConfigFilename, err))
					return nil, err
				}

				logger.Debug("Found " + DynatraceConfigFilename + " on project level")
			} else {
				logger.Debug("Found " + DynatraceConfigFilename + " on stage level")
			}
		} else {
			logger.Debug("Found " + DynatraceConfigFilename + " on service level")
		}
		fileContent = keptnResourceContent.ResourceContent
	}

	// replace the placeholders
	logger.Debug("Content of dynatrace.conf.yaml: " + fileContent)
	fileContent = replaceKeptnPlaceholders(fileContent, keptnEvent)
	logger.Debug("After replacements: " + fileContent)

	// unmarshal the file
	dynatraceConfFile, err := parseDynatraceConfigFile([]byte(fileContent))

	if err != nil {
		logMessage := fmt.Sprintf("Couldn't parse %s file found for service %s in stage %s in project %s. Error: %s", DynatraceConfigFilename, keptnEvent.service, keptnEvent.stage, keptnEvent.project, err.Error())
		logger.Error(logMessage)
		return nil, errors.New(logMessage)
	}

	// logMessage := fmt.Sprintf("Loaded Config from dynatrace.conf.yaml:  %s", dynatraceConfFile)
	// logger.Info(logMessage)

	return dynatraceConfFile, nil
}

func parseDynatraceConfigFile(input []byte) (*DynatraceConfigFile, error) {
	dynatraceConfFile := &DynatraceConfigFile{}
	err := yaml.Unmarshal([]byte(input), &dynatraceConfFile)

	if err != nil {
		return nil, err
	}

	return dynatraceConfFile, nil
}

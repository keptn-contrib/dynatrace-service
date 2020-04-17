package event_handler

import (
	"errors"
	"fmt"
	"io/ioutil"

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
	AttachRules *dtAttachRules `json:"attachRules" yaml:"attachRules"`
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
	default:
		return &CDEventHandler{Logger: logger, Event: event}, nil
	}
}

//
// Loads dynatrace.conf for the current service
//
func getDynatraceConfig(project, service, stage string, logger *keptn.Logger) (*DynatraceConfigFile, error) {

	logger.Info("Loading dynatrace.conf.yaml")
	// if we run in a runlocal mode we are just getting the file from the local disk
	var fileContent []byte
	var err error
	if common.RunLocal {
		fileContent, err = ioutil.ReadFile(DynatraceConfigFilenameLOCAL)
		if err != nil {
			logMessage := fmt.Sprintf("No %s file found LOCALLY for service %s in stage %s in project %s", DynatraceConfigFilenameLOCAL, service, stage, project)
			logger.Info(logMessage)
			return nil, nil
		}
	} else {
		resourceHandler := utils.NewResourceHandler("configuration-service:8080")
		keptnResourceContent, err := resourceHandler.GetServiceResource(service, stage, project, DynatraceConfigFilename)
		if err != nil {
			logMessage := fmt.Sprintf("No %s file found for service %s in stage %s in project %s", DynatraceConfigFilename, service, stage, project)
			logger.Info(logMessage)
			return nil, nil
		}
		fileContent = []byte(keptnResourceContent.ResourceContent)
	}

	// unmarshal the file
	var dynatraceConfFile *DynatraceConfigFile
	dynatraceConfFile, err = parseDynatraceConfigFile(fileContent)

	if err != nil {
		logMessage := fmt.Sprintf("Couldn't parse %s file found for service %s in stage %s in project %s. Error: %s", DynatraceConfigFilename, service, stage, project, err.Error())
		logger.Error(logMessage)
		return nil, errors.New(logMessage)
	}

	logMessage := fmt.Sprintf("Loaded Config from dynatrace.conf.yaml:  %s", dynatraceConfFile)
	logger.Info(logMessage)

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

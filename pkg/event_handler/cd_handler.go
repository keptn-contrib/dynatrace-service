package event_handler

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type CDEventHandler struct {
	Logger *keptn.Logger
	Event  cloudevents.Event
}

func (eh CDEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	clientSet, err := common.GetKubernetesClient()
	if err != nil {
		eh.Logger.Error("could not create k8s client")
		return err
	}

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}

	dtHelper, err := lib.NewDynatraceHelper(keptnHandler)
	if err != nil {
		eh.Logger.Error("Could not create Dynatrace Helper: " + err.Error())
		return err
	}
	dtHelper.KubeApi = clientSet
	dtHelper.Logger = eh.Logger

	eh.Logger.Info("Checking if event of type " + eh.Event.Type() + " should be sent to Dynatrace...")

	if eh.Event.Type() == keptn.DeploymentFinishedEventType {
		dfData := &keptn.DeploymentFinishedEventData{}
		err := eh.Event.DataAs(dfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}
		de := createDeploymentEvent(dfData, shkeptncontext, eh.Logger)

		dtHelper.SendEvent(de)

		// TODO: an additional channel (e.g. start-tests) to correctly determine the time when the tests actually start
		ie := createInfoEvent(dfData.Project, dfData.Stage, dfData.Service, dfData.TestStrategy, dfData.Image, dfData.Tag, dfData.Labels, shkeptncontext, eh.Logger)
		if dfData.TestStrategy != "" {
			if ie.Title == "" {
				ie.Title = "Start Running Tests: " + dfData.TestStrategy
			}
			if ie.Description == "" {
				ie.Description = "Start running tests: " + dfData.TestStrategy + " against " + dfData.Service
			}
			dtHelper.SendEvent(ie)
		}
	} else if eh.Event.Type() == keptn.TestsFinishedEventType {
		tfData := &keptn.TestsFinishedEventData{}
		err := eh.Event.DataAs(tfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}
		ie := createInfoEvent(tfData.Project, tfData.Stage, tfData.Service, tfData.TestStrategy, "", "", tfData.Labels, shkeptncontext, eh.Logger)
		if ie.Title == "" {
			ie.Title = "Stop Running Tests: " + tfData.TestStrategy
		}
		if ie.Description == "" {
			ie.Description = "Stop running tests: " + tfData.TestStrategy + " against " + tfData.Service
		}
		dtHelper.SendEvent(ie)
	} else if eh.Event.Type() == keptn.EvaluationDoneEventType {
		edData := &keptn.EvaluationDoneEventData{}
		err := eh.Event.DataAs(edData)
		if err != nil {
			fmt.Println("Error while parsing JSON payload: " + err.Error())
			return err
		}
		ie := createInfoEvent(edData.Project, edData.Stage, edData.Service, edData.TestStrategy, "", "", edData.Labels, shkeptncontext, eh.Logger)
		if edData.Result == "pass" || edData.Result == "warning" {
			if edData.TestStrategy == "real-user" {
				ie.Title = "Remediation action successful"
			} else {
				ie.Title = "Promote Artifact from " + edData.Stage + " to next stage"
			}

		} else if edData.Result == "fail" && edData.DeploymentStrategy == "blue_green_service" {
			if edData.TestStrategy == "real-user" {
				ie.Title = "Remediation action not successful"
			} else {
				ie.Title = "Rollback Artifact (Switch Blue/Green) in " + edData.Stage
			}
		} else if edData.Result == "fail" && edData.DeploymentStrategy == "direct" {
			if edData.TestStrategy == "real-user" {
				ie.Title = "Remediation action not successful"
			} else {
				ie.Title = "NOT PROMOTING Artifact from " + edData.Stage + " due to failed evaluation"
			}
		} else {
			eh.Logger.Error("No valid deployment strategy defined in keptn event.")
			return nil
		}
		ie.Description = "Keptn evaluation status: " + edData.Result
		dtHelper.SendEvent(ie)
	} else {
		eh.Logger.Info("    Ignoring event.")
	}
	return nil
}

type dtTag struct {
	Context string `json:"context" yaml:"context"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value" yaml:"value"`
}

type dtTagRule struct {
	MeTypes []string `json:"meTypes" yaml:"meTypes"`
	Tags    []dtTag  `json:"tags" yaml:"tags"`
}

type dtAttachRules struct {
	TagRule []dtTagRule `json:"tagRule" yaml:"tagRule"`
}

type dtCustomProperties struct {
	Project            string `json:"Project"`
	Stage              string `json:"Stage"`
	Service            string `json:"Service"`
	TestStrategy       string `json:"Test strategy"`
	DeploymentStrategy string `json:"Deployment strategy"`
	Image              string `json:"Image"`
	Tag                string `json:"Tag"`
	KeptnContext       string `json:"Keptn context"`
}

type dtDeploymentEvent struct {
	EventType   string        `json:"eventType"`
	Source      string        `json:"source"`
	AttachRules dtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties  map[string]string `json:"customProperties"`
	DeploymentVersion string            `json:"deploymentVersion"`
	DeploymentName    string            `json:"deploymentName"`
	DeploymentProject string            `json:"deploymentProject"`
	CiBackLink        string            `json:"ciBackLink",omitempty`
	RemediationAction string            `json:"remediationAction",omitempty`
}

type dtInfoEvent struct {
	EventType   string        `json:"eventType"`
	Source      string        `json:"source"`
	AttachRules dtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Title            string            `json:"title"`
}

/**
 * Changes in #115_116: Parse Tags from dynatrace.conf.yaml and only fall back to default behavior if it doesnt exist
 */
func createAttachRules(project string, stage string, service string, logger *keptn.Logger) dtAttachRules {
	dynatraceConfig, err := getDynatraceConfig(project, stage, service, logger)

	if err != nil {
		logMessage := fmt.Sprintf("Error retrieving dynatrace.conf.yaml for %s/%s/%s: %s. Going with default", service, stage, project, err.Error())
		logger.Error(logMessage)
		dynatraceConfig = nil
	}

	if dynatraceConfig != nil && dynatraceConfig.AttachRules != nil {
		return *dynatraceConfig.AttachRules
	}

	ar := dtAttachRules{
		TagRule: []dtTagRule{
			dtTagRule{
				MeTypes: []string{"SERVICE"},
				Tags: []dtTag{
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   project,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   stage,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   service,
					},
				},
			},
		},
	}

	return ar
}

/**
 * Change with #115_116: parse labels and move them into custom properties
 */
// func createCustomProperties(project string, stage string, service string, testStrategy string, image string, tag string, labels map[string]string, keptnContext string) dtCustomProperties {
func createCustomProperties(project string, stage string, service string, testStrategy string, image string, tag string, labels map[string]string, keptnContext string, logger *keptn.Logger) map[string]string {
	// TODO: AG - parse labels and push them through

	// var customProperties dtCustomProperties
	// customProperties.Project = project
	// customProperties.Stage = stage
	// customProperties.Service = service
	// customProperties.TestStrategy = testStrategy
	// customProperties.Image = image
	// customProperties.Tag = tag
	// customProperties.KeptnContext = keptnContext
	var customProperties map[string]string
	customProperties = make(map[string]string)
	customProperties["Project"] = project
	customProperties["Stage"] = stage
	customProperties["Service"] = service
	customProperties["TestStrategy"] = testStrategy
	customProperties["Image"] = image
	customProperties["Tag"] = tag
	customProperties["KeptnContext"] = keptnContext

	// now add the rest of the labels
	for key, value := range labels {
		customProperties[key] = value
	}

	return customProperties
}

/**
 * Returns the value of the map if the value exists - otherwise returns default
 * Also removes the found value from the map if removeIfFound==true
 */
func getValueFromLabels(labels *map[string]string, valueKey string, defaultValue string, removeIfFound bool) string {
	mapValue, mapValueOk := (*labels)[valueKey]
	if mapValueOk {
		if removeIfFound {
			delete(*labels, valueKey)
		}
		return mapValue
	}

	return defaultValue
}

func createInfoEvent(project string, stage string, service string, testStrategy string, image string, tag string, labels map[string]string, keptnContext string, logger *keptn.Logger) dtInfoEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie dtInfoEvent
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "Keptn dynatrace-service"
	ie.Title = getValueFromLabels(&labels, "title", "", true)
	ie.Description = getValueFromLabels(&labels, "description", "", true)

	// now we create our attach rules
	ar := createAttachRules(project, stage, service, logger)
	ie.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(project, stage, service, testStrategy, image, tag, labels, keptnContext, logger)
	ie.CustomProperties = customProperties

	return ie
}

func createDeploymentEvent(event *keptn.DeploymentFinishedEventData, keptnContext string, logger *keptn.Logger) dtDeploymentEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de dtDeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "Keptn dynatrace-service"
	de.DeploymentName = getValueFromLabels(&event.Labels, "deploymentName", "Deploy "+event.Service+" "+event.Tag+" with strategy "+event.DeploymentStrategy, true)
	de.DeploymentProject = getValueFromLabels(&event.Labels, "deploymentProject", event.Project, true)
	de.DeploymentVersion = getValueFromLabels(&event.Labels, "deploymentVersion", event.Tag, true)
	de.CiBackLink = getValueFromLabels(&event.Labels, "ciBackLink", "", true)
	de.RemediationAction = getValueFromLabels(&event.Labels, "remediationAction", "", true)

	// now we create our attach rules
	ar := createAttachRules(event.Project, event.Stage, event.Service, logger)
	de.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext, logger)
	de.CustomProperties = customProperties

	return de
}

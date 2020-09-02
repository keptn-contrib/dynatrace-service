package event_handler

import (
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type dtTag struct {
	Context string `json:"context" yaml:"context"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value",omitempty yaml:"value",omitempty`
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

type dtConfigurationEvent struct {
	EventType   string        `json:"eventType"`
	Source      string        `json:"source"`
	AttachRules dtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Configuration    string            `json:"Configuration"`
	Original         string            `json:"Original,omitempty"`
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

type dtAnnotationEvent struct {
	EventType   string        `json:"eventType"`
	Source      string        `json:"source"`
	AttachRules dtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties      map[string]string `json:"customProperties"`
	AnnotationDescription string            `json:"annotationDescription"`
	AnnotationType        string            `json:"annotationType"`
}

/**
 * Changes in #115_116: Parse Tags from dynatrace.conf.yaml and only fall back to default behavior if it doesnt exist
 */
func createAttachRules(keptnEvent *baseKeptnEvent, dynatraceConfig *DynatraceConfigFile, logger *keptn.Logger) dtAttachRules {
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
						Value:   keptnEvent.project,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   keptnEvent.stage,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   keptnEvent.service,
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
func createCustomProperties(keptnEvent *baseKeptnEvent, logger *keptn.Logger) map[string]string {
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
	customProperties["Project"] = keptnEvent.project
	customProperties["Stage"] = keptnEvent.stage
	customProperties["Service"] = keptnEvent.service
	customProperties["TestStrategy"] = keptnEvent.testStrategy
	customProperties["Image"] = keptnEvent.image
	customProperties["Tag"] = keptnEvent.tag
	customProperties["KeptnContext"] = keptnEvent.context

	// now add the rest of the labels
	for key, value := range keptnEvent.labels {
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

// project string, stage string, service string, testStrategy string, image string, tag string, labels map[string]string, keptnContext string
func CreateInfoEvent(keptnEvent *baseKeptnEvent, dynatraceConfig *DynatraceConfigFile, logger *keptn.Logger) dtInfoEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie dtInfoEvent
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "Keptn dynatrace-service"
	ie.Title = getValueFromLabels(&keptnEvent.labels, "title", "", true)
	ie.Description = getValueFromLabels(&keptnEvent.labels, "description", "", true)

	// now we create our attach rules
	ar := createAttachRules(keptnEvent, dynatraceConfig, logger)
	ie.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(keptnEvent, logger)
	ie.CustomProperties = customProperties

	return ie
}

/**
 * Creates a Dynatrace ANNOTATION event
 */
func CreateAnnotationEvent(keptnEvent *baseKeptnEvent, dynatraceConfig *DynatraceConfigFile, logger *keptn.Logger) dtAnnotationEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie dtAnnotationEvent
	ie.EventType = "CUSTOM_ANNOTATION"
	ie.Source = "Keptn dynatrace-service"
	ie.AnnotationType = getValueFromLabels(&keptnEvent.labels, "type", "", true)
	ie.AnnotationDescription = getValueFromLabels(&keptnEvent.labels, "description", "", true)

	// now we create our attach rules
	ar := createAttachRules(keptnEvent, dynatraceConfig, logger)
	ie.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(keptnEvent, logger)
	ie.CustomProperties = customProperties

	return ie
}

func CreateDeploymentEvent(keptnEvent *baseKeptnEvent, dynatraceConfig *DynatraceConfigFile, logger *keptn.Logger) dtDeploymentEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de dtDeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "Keptn dynatrace-service"
	de.DeploymentName = getValueFromLabels(&keptnEvent.labels, "deploymentName", "Deploy "+keptnEvent.service+" "+keptnEvent.tag+" with strategy "+keptnEvent.deploymentStrategy, true)
	de.DeploymentProject = getValueFromLabels(&keptnEvent.labels, "deploymentProject", keptnEvent.project, true)
	de.DeploymentVersion = getValueFromLabels(&keptnEvent.labels, "deploymentVersion", keptnEvent.tag, true)
	de.CiBackLink = getValueFromLabels(&keptnEvent.labels, "ciBackLink", "", true)
	de.RemediationAction = getValueFromLabels(&keptnEvent.labels, "remediationAction", "", true)

	// now we create our attach rules
	ar := createAttachRules(keptnEvent, dynatraceConfig, logger)
	de.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(keptnEvent, logger)
	de.CustomProperties = customProperties

	return de
}

func CreateConfigurationEvent(keptnEvent *baseKeptnEvent, dynatraceConfig *DynatraceConfigFile, logger *keptn.Logger) dtConfigurationEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de dtConfigurationEvent
	de.EventType = "CUSTOM_CONFIGURATION"
	de.Source = "Keptn dynatrace-service"

	// now we create our attach rules
	ar := createAttachRules(keptnEvent, dynatraceConfig, logger)
	de.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(keptnEvent, logger)
	de.CustomProperties = customProperties

	return de
}

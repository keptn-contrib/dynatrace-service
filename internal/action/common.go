package action

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

/**
 * Change with #115_116: parse labels and move them into custom properties
 */
func createCustomProperties(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag) map[string]string {
	// TODO: AG - parse labels and push them through

	// var customProperties dtCustomProperties
	// customProperties.Project = Project
	// customProperties.Stage = Stage
	// customProperties.Service = Service
	// customProperties.TestStrategy = TestStrategy
	// customProperties.Image = Image
	// customProperties.Tag = Tag
	// customProperties.KeptnContext = keptnContext
	var customProperties map[string]string
	customProperties = make(map[string]string)
	customProperties["Project"] = a.GetProject()
	customProperties["Stage"] = a.GetStage()
	customProperties["Service"] = a.GetService()
	customProperties["TestStrategy"] = a.GetTestStrategy()
	customProperties["Image"] = imageAndTag.Image()
	customProperties["Tag"] = imageAndTag.Tag()
	customProperties["KeptnContext"] = a.GetShKeptnContext()
	customProperties["Keptn Service"] = a.GetSource()

	// now add the rest of the Labels
	for key, value := range a.GetLabels() {
		customProperties[key] = value
	}

	return customProperties
}

// createInfoEventDTO creates a new Dynatrace CUSTOM_INFO event
func createInfoEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.InfoEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie dynatrace.InfoEvent
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "Keptn dynatrace-service"
	ie.Title = a.GetLabels()["title"]
	ie.Description = a.GetLabels()["description"]
	ie.AttachRules = *attachRules

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(a, imageAndTag)
	ie.CustomProperties = customProperties

	return ie
}

// createAnnotationEventDTO creates a Dynatrace CUSTOM_ANNOTATION event
func createAnnotationEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.AnnotationEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie dynatrace.AnnotationEvent
	ie.EventType = "CUSTOM_ANNOTATION"
	ie.Source = "Keptn dynatrace-service"
	ie.AnnotationType = a.GetLabels()["type"]
	ie.AnnotationDescription = a.GetLabels()["description"]
	ie.AttachRules = *attachRules

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(a, imageAndTag)
	ie.CustomProperties = customProperties

	return ie
}

func getValueFromLabels(a adapter.EventContentAdapter, key string, defaultValue string) string {
	v := a.GetLabels()[key]
	if len(v) > 0 {
		return v
	}
	return defaultValue
}

// createDeploymentEventDTO creates a Dynatrace CUSTOM_DEPLOYMENT event
func createDeploymentEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.DeploymentEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de dynatrace.DeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "Keptn dynatrace-service"
	de.DeploymentName = getValueFromLabels(a, "deploymentName", "Deploy "+a.GetService()+" "+imageAndTag.Tag()+" with strategy "+a.GetDeploymentStrategy())
	de.DeploymentProject = getValueFromLabels(a, "deploymentProject", a.GetProject())
	de.DeploymentVersion = getValueFromLabels(a, "deploymentVersion", imageAndTag.Tag())
	de.CiBackLink = getValueFromLabels(a, "ciBackLink", "")
	de.RemediationAction = getValueFromLabels(a, "remediationAction", "")
	de.AttachRules = *attachRules

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(a, imageAndTag)
	de.CustomProperties = customProperties

	return de
}

// createConfigurationEventDTO creates a Dynatrace CUSTOM_CONFIGURATION event
func createConfigurationEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.ConfigurationEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de dynatrace.ConfigurationEvent
	de.EventType = "CUSTOM_CONFIGURATION"
	de.Source = "Keptn dynatrace-service"
	de.AttachRules = *attachRules

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(a, imageAndTag)
	de.CustomProperties = customProperties

	return de
}

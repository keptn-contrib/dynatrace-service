package action

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

const eventSource = "Keptn dynatrace-service"

func createAnnotationEventDTO(a adapter.EventContentAdapter, customProperties map[string]string, attachRules *dynatrace.AttachRules) dynatrace.AnnotationEvent {
	return dynatrace.AnnotationEvent{
		EventType:             dynatrace.AnnotationEventType,
		Source:                eventSource,
		AnnotationType:        a.GetLabels()["type"],
		AnnotationDescription: a.GetLabels()["description"],
		CustomProperties:      customProperties,
		AttachRules:           *attachRules,
	}
}

func createConfigurationEventDTO(a adapter.EventContentAdapter, customProperties map[string]string, attachRules *dynatrace.AttachRules) dynatrace.ConfigurationEvent {
	return dynatrace.ConfigurationEvent{
		EventType:        dynatrace.ConfigurationEventType,
		Source:           eventSource,
		CustomProperties: customProperties,
		AttachRules:      *attachRules,
	}
}

func createDeploymentEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, customProperties map[string]string, attachRules *dynatrace.AttachRules) dynatrace.DeploymentEvent {
	return dynatrace.DeploymentEvent{
		EventType:         dynatrace.DeploymentEventType,
		Source:            eventSource,
		DeploymentName:    getValueFromLabels(a, "deploymentName", "Deploy "+a.GetService()+" "+imageAndTag.Tag()+" with strategy "+a.GetDeploymentStrategy()),
		DeploymentProject: getValueFromLabels(a, "deploymentProject", a.GetProject()),
		DeploymentVersion: getValueFromLabels(a, "deploymentVersion", imageAndTag.Tag()),
		CiBackLink:        getValueFromLabels(a, "ciBackLink", ""),
		RemediationAction: getValueFromLabels(a, "remediationAction", ""),
		CustomProperties:  customProperties,
		AttachRules:       *attachRules,
	}
}

func createInfoEventDTO(a adapter.EventContentAdapter, customProperties map[string]string, attachRules *dynatrace.AttachRules) dynatrace.InfoEvent {
	return dynatrace.InfoEvent{
		EventType:        dynatrace.InfoEventType,
		Source:           eventSource,
		Title:            a.GetLabels()["title"],
		Description:      a.GetLabels()["description"],
		CustomProperties: customProperties,
		AttachRules:      *attachRules,
	}
}

func createCustomProperties(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag) map[string]string {
	customProperties := map[string]string{
		"Project":       a.GetProject(),
		"Stage":         a.GetStage(),
		"Service":       a.GetService(),
		"TestStrategy":  a.GetTestStrategy(),
		"Image":         imageAndTag.Image(),
		"Tag":           imageAndTag.Tag(),
		"KeptnContext":  a.GetShKeptnContext(),
		"Keptn Service": a.GetSource(),
	}

	// now add the rest of the labels into custom properties (changed with #115_116)
	for key, value := range a.GetLabels() {
		customProperties[key] = value
	}

	return customProperties
}

func getValueFromLabels(a adapter.EventContentAdapter, key string, defaultValue string) string {
	v := a.GetLabels()[key]
	if len(v) > 0 {
		return v
	}
	return defaultValue
}

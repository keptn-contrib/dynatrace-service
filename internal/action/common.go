package action

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

func createAnnotationEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.AnnotationEvent {
	return dynatrace.AnnotationEvent{
		EventType:             "CUSTOM_ANNOTATION",
		Source:                "Keptn dynatrace-service",
		AnnotationType:        a.GetLabels()["type"],
		AnnotationDescription: a.GetLabels()["description"],
		AttachRules:           *attachRules,
		CustomProperties:      createCustomProperties(a, imageAndTag),
	}
}

func createConfigurationEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.ConfigurationEvent {
	return dynatrace.ConfigurationEvent{
		EventType:        "CUSTOM_CONFIGURATION",
		Source:           "Keptn dynatrace-service",
		AttachRules:      *attachRules,
		CustomProperties: createCustomProperties(a, imageAndTag),
	}
}

func createDeploymentEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.DeploymentEvent {
	return dynatrace.DeploymentEvent{
		EventType:         "CUSTOM_DEPLOYMENT",
		Source:            "Keptn dynatrace-service",
		DeploymentName:    getValueFromLabels(a, "deploymentName", "Deploy "+a.GetService()+" "+imageAndTag.Tag()+" with strategy "+a.GetDeploymentStrategy()),
		DeploymentProject: getValueFromLabels(a, "deploymentProject", a.GetProject()),
		DeploymentVersion: getValueFromLabels(a, "deploymentVersion", imageAndTag.Tag()),
		CiBackLink:        getValueFromLabels(a, "ciBackLink", ""),
		RemediationAction: getValueFromLabels(a, "remediationAction", ""),
		AttachRules:       *attachRules,
		CustomProperties:  createCustomProperties(a, imageAndTag),
	}
}

func createInfoEventDTO(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, attachRules *dynatrace.AttachRules) dynatrace.InfoEvent {
	return dynatrace.InfoEvent{
		EventType:        "CUSTOM_INFO",
		Source:           "Keptn dynatrace-service",
		Title:            a.GetLabels()["title"],
		Description:      a.GetLabels()["description"],
		AttachRules:      *attachRules,
		CustomProperties: createCustomProperties(a, imageAndTag),
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

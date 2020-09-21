package event_handler

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type dtConfigurationEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Configuration    string            `json:"Configuration"`
	Original         string            `json:"Original,omitempty"`
}

type dtDeploymentEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties  map[string]string `json:"customProperties"`
	DeploymentVersion string            `json:"deploymentVersion"`
	DeploymentName    string            `json:"deploymentName"`
	DeploymentProject string            `json:"deploymentProject"`
	CiBackLink        string            `json:"ciBackLink",omitempty`
	RemediationAction string            `json:"remediationAction",omitempty`
}

type dtInfoEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Title            string            `json:"title"`
}

type dtAnnotationEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties      map[string]string `json:"customProperties"`
	AnnotationDescription string            `json:"annotationDescription"`
	AnnotationType        string            `json:"annotationType"`
}

/**
 * Changes in #115_116: Parse Tags from dynatrace.conf.yaml and only fall back to default behavior if it doesnt exist
 */
func createAttachRules(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile, logger *keptn.Logger) config.DtAttachRules {
	if dynatraceConfig != nil && dynatraceConfig.AttachRules != nil {
		return *dynatraceConfig.AttachRules
	}

	ar := config.DtAttachRules{
		TagRule: []config.DtTagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []config.DtTag{
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   a.GetProject(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   a.GetStage(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   a.GetService(),
					},
				},
			},
		},
	}

	return ar
}

/**
 * Change with #115_116: parse Labels and move them into custom properties
 */
// func createCustomProperties(Project string, Stage string, Service string, TestStrategy string, Image string, Tag string, Labels map[string]string, keptnContext string) dtCustomProperties {
func createCustomProperties(a adapter.EventContentAdapter, logger *keptn.Logger) map[string]string {
	// TODO: AG - parse Labels and push them through

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
	customProperties["Image"] = a.GetImage()
	customProperties["Tag"] = a.GetTag()
	customProperties["KeptnContext"] = a.GetShKeptnContext()

	// now add the rest of the Labels
	for key, value := range a.GetLabels() {
		customProperties[key] = value
	}

	return customProperties
}

// Project string, Stage string, Service string, TestStrategy string, Image string, Tag string, Labels map[string]string, keptnContext string
func CreateInfoEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile, logger *keptn.Logger) dtInfoEvent {

	// we fill the Dynatrace Info Event with values from the Labels or use our defaults
	var ie dtInfoEvent
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "Keptn dynatrace-Service"
	ie.Title = a.GetLabels()["title"]
	ie.Description = a.GetLabels()["description"]

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig, logger)
	ie.AttachRules = ar

	// and add the rest of the Labels and info as custom properties
	customProperties := createCustomProperties(a, logger)
	ie.CustomProperties = customProperties

	return ie
}

/**
 * Creates a Dynatrace ANNOTATION Event
 */
func CreateAnnotationEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile, logger *keptn.Logger) dtAnnotationEvent {

	// we fill the Dynatrace Info Event with values from the Labels or use our defaults
	var ie dtAnnotationEvent
	ie.EventType = "CUSTOM_ANNOTATION"
	ie.Source = "Keptn dynatrace-Service"
	ie.AnnotationType = a.GetLabels()["type"]
	ie.AnnotationDescription = a.GetLabels()["description"]

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig, logger)
	ie.AttachRules = ar

	// and add the rest of the Labels and info as custom properties
	customProperties := createCustomProperties(a, logger)
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

func CreateDeploymentEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile, logger *keptn.Logger) dtDeploymentEvent {

	// we fill the Dynatrace Deployment Event with values from the Labels or use our defaults
	var de dtDeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "Keptn dynatrace-Service"
	de.DeploymentName = getValueFromLabels(a, "deploymentName", "Deploy "+a.GetService()+" "+a.GetTag()+" with strategy "+a.GetDeploymentStrategy())
	de.DeploymentProject = getValueFromLabels(a, "deploymentProject", a.GetProject())
	de.DeploymentVersion = getValueFromLabels(a, "deploymentVersion", a.GetTag())
	de.CiBackLink = getValueFromLabels(a, "ciBackLink", "")
	de.RemediationAction = getValueFromLabels(a, "remediationAction", "")

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig, logger)
	de.AttachRules = ar

	// and add the rest of the Labels and info as custom properties
	// TODO: Event.Project, Event.Stage, Event.Service, Event.TestStrategy, Event.Image, Event.Tag, Event.Labels, keptnContext
	customProperties := createCustomProperties(a, logger)
	de.CustomProperties = customProperties

	return de
}

func CreateConfigurationEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile, logger *keptn.Logger) dtConfigurationEvent {

	// we fill the Dynatrace Deployment Event with values from the Labels or use our defaults
	var de dtConfigurationEvent
	de.EventType = "CUSTOM_CONFIGURATION"
	de.Source = "Keptn dynatrace-Service"

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig, logger)
	de.AttachRules = ar

	// and add the rest of the Labels and info as custom properties
	// TODO: Event.Project, Event.Stage, Event.Service, Event.TestStrategy, Event.Image, Event.Tag, Event.Labels, keptnContext
	customProperties := createCustomProperties(a, logger)
	de.CustomProperties = customProperties

	return de
}

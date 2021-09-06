package event

import (
	"net/url"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
)

type DTConfigurationEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Configuration    string            `json:"configuration"`
	Original         string            `json:"original,omitempty"`
}

type DTDeploymentEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties  map[string]string `json:"customProperties"`
	DeploymentVersion string            `json:"deploymentVersion"`
	DeploymentName    string            `json:"deploymentName"`
	DeploymentProject string            `json:"deploymentProject"`
	CiBackLink        string            `json:"ciBackLink,omitempty"`
	RemediationAction string            `json:"remediationAction,omitempty"`
}

type DTInfoEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Title            string            `json:"title"`
}

type DTAnnotationEvent struct {
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
func createAttachRules(a adapter.EventContentAdapter, attachRules *config.DtAttachRules) config.DtAttachRules {
	if attachRules != nil {
		return *attachRules
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
 * Change with #115_116: parse labels and move them into custom properties
 */
func createCustomProperties(a adapter.EventContentAdapter) map[string]string {
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
	customProperties["Image"] = a.GetImage()
	customProperties["Tag"] = a.GetTag()
	customProperties["KeptnContext"] = a.GetShKeptnContext()
	customProperties["Keptn Service"] = a.GetSource()

	// now add the rest of the Labels
	for key, value := range a.GetLabels() {
		customProperties[key] = value
	}

	return customProperties
}

// CreateInfoEvent creates a new Info event
func CreateInfoEvent(a adapter.EventContentAdapter, attachRules *config.DtAttachRules) DTInfoEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie DTInfoEvent
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "Keptn dynatrace-service"
	ie.Title = a.GetLabels()["title"]
	ie.Description = a.GetLabels()["description"]

	// now we create our attach rules
	ar := createAttachRules(a, attachRules)
	ie.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(a)
	ie.CustomProperties = customProperties

	return ie
}

// CreateAnnotationEvent creates a Dynatrace ANNOTATION event
func CreateAnnotationEvent(a adapter.EventContentAdapter, attachRules *config.DtAttachRules) DTAnnotationEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie DTAnnotationEvent
	ie.EventType = "CUSTOM_ANNOTATION"
	ie.Source = "Keptn dynatrace-service"
	ie.AnnotationType = a.GetLabels()["type"]
	ie.AnnotationDescription = a.GetLabels()["description"]

	// now we create our attach rules
	ar := createAttachRules(a, attachRules)
	ie.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(a)
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

func CreateDeploymentEvent(a adapter.EventContentAdapter, attachRules *config.DtAttachRules) DTDeploymentEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de DTDeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "Keptn dynatrace-service"
	de.DeploymentName = getValueFromLabels(a, "deploymentName", "Deploy "+a.GetService()+" "+a.GetTag()+" with strategy "+a.GetDeploymentStrategy())
	de.DeploymentProject = getValueFromLabels(a, "deploymentProject", a.GetProject())
	de.DeploymentVersion = getValueFromLabels(a, "deploymentVersion", a.GetTag())
	de.CiBackLink = getValueFromLabels(a, "ciBackLink", "")
	de.RemediationAction = getValueFromLabels(a, "remediationAction", "")

	// now we create our attach rules
	ar := createAttachRules(a, attachRules)
	de.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(a)
	de.CustomProperties = customProperties

	return de
}

func CreateConfigurationEvent(a adapter.EventContentAdapter, attachRules *config.DtAttachRules) DTConfigurationEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de DTConfigurationEvent
	de.EventType = "CUSTOM_CONFIGURATION"
	de.Source = "Keptn dynatrace-service"

	// now we create our attach rules
	ar := createAttachRules(a, attachRules)
	de.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(a)
	de.CustomProperties = customProperties

	return de
}

// GetEventSource gets the source to be used for CloudEvents originating from the dynatrace-service
func GetEventSource() string {
	source, _ := url.Parse("dynatrace-service")
	return source.String()
}

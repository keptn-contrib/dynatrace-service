package event

import (
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	log "github.com/sirupsen/logrus"
)

type DtConfigurationEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Configuration    string            `json:"configuration"`
	Original         string            `json:"original,omitempty"`
}

type DtDeploymentEvent struct {
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

type DtInfoEvent struct {
	EventType   string               `json:"eventType"`
	Source      string               `json:"source"`
	AttachRules config.DtAttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Title            string            `json:"title"`
}

type DtAnnotationEvent struct {
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
func createAttachRules(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile) config.DtAttachRules {
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
func CreateInfoEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile) DtInfoEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie DtInfoEvent
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "Keptn dynatrace-service"
	ie.Title = a.GetLabels()["title"]
	ie.Description = a.GetLabels()["description"]

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig)
	ie.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	customProperties := createCustomProperties(a)
	ie.CustomProperties = customProperties

	return ie
}

// CreateAnnotationEvent creates a Dynatrace ANNOTATION event
func CreateAnnotationEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile) DtAnnotationEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie DtAnnotationEvent
	ie.EventType = "CUSTOM_ANNOTATION"
	ie.Source = "Keptn dynatrace-service"
	ie.AnnotationType = a.GetLabels()["type"]
	ie.AnnotationDescription = a.GetLabels()["description"]

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig)
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

func CreateDeploymentEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile) DtDeploymentEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de DtDeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "Keptn dynatrace-service"
	de.DeploymentName = getValueFromLabels(a, "deploymentName", "Deploy "+a.GetService()+" "+a.GetTag()+" with strategy "+a.GetDeploymentStrategy())
	de.DeploymentProject = getValueFromLabels(a, "deploymentProject", a.GetProject())
	de.DeploymentVersion = getValueFromLabels(a, "deploymentVersion", a.GetTag())
	de.CiBackLink = getValueFromLabels(a, "ciBackLink", "")
	de.RemediationAction = getValueFromLabels(a, "remediationAction", "")

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig)
	de.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(a)
	de.CustomProperties = customProperties

	return de
}

func CreateConfigurationEvent(a adapter.EventContentAdapter, dynatraceConfig *config.DynatraceConfigFile) DtConfigurationEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de DtConfigurationEvent
	de.EventType = "CUSTOM_CONFIGURATION"
	de.Source = "Keptn dynatrace-service"

	// now we create our attach rules
	ar := createAttachRules(a, dynatraceConfig)
	de.AttachRules = ar

	// and add the rest of the labels and info as custom properties
	// TODO: event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, event.Labels, keptnContext
	customProperties := createCustomProperties(a)
	de.CustomProperties = customProperties

	return de
}

// GetShKeptnContext extracts the keptn context from a CloudEvent
func GetShKeptnContext(event cloudevents.Event) string {
	shkeptncontext, err := types.ToString(event.Context.GetExtensions()["shkeptncontext"])
	if err != nil {
		log.WithError(err).Debug("Event does not contain shkeptncontext")
	}
	return shkeptncontext
}

// GetEventSource gets the source to be used for CloudEvents originating from the dynatrace-service
func GetEventSource() string {
	source, _ := url.Parse("dynatrace-service")
	return source.String()
}

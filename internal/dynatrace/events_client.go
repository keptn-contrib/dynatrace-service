package dynatrace

import (
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	log "github.com/sirupsen/logrus"
)

const eventsPath = "/api/v1/events"

type ConfigurationEvent struct {
	EventType   string      `json:"eventType"`
	Source      string      `json:"source"`
	AttachRules AttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Configuration    string            `json:"configuration"`
	Original         string            `json:"original,omitempty"`
}

type DeploymentEvent struct {
	EventType   string      `json:"eventType"`
	Source      string      `json:"source"`
	AttachRules AttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties  map[string]string `json:"customProperties"`
	DeploymentVersion string            `json:"deploymentVersion"`
	DeploymentName    string            `json:"deploymentName"`
	DeploymentProject string            `json:"deploymentProject"`
	CiBackLink        string            `json:"ciBackLink,omitempty"`
	RemediationAction string            `json:"remediationAction,omitempty"`
}

type InfoEvent struct {
	EventType   string      `json:"eventType"`
	Source      string      `json:"source"`
	AttachRules AttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Title            string            `json:"title"`
}

type AnnotationEvent struct {
	EventType   string      `json:"eventType"`
	Source      string      `json:"source"`
	AttachRules AttachRules `json:"attachRules"`
	// CustomProperties  dtCustomProperties `json:"customProperties"`
	CustomProperties      map[string]string `json:"customProperties"`
	AnnotationDescription string            `json:"annotationDescription"`
	AnnotationType        string            `json:"annotationType"`
}

// TagEntry defines a Dynatrace configuration structure
type TagEntry struct {
	Context string `json:"context" yaml:"context"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
}

// TagRule defines a Dynatrace configuration structure
type TagRule struct {
	MeTypes []string   `json:"meTypes" yaml:"meTypes"`
	Tags    []TagEntry `json:"tags" yaml:"tags"`
}

// AttachRules defines a Dynatrace configuration structure
type AttachRules struct {
	TagRule []TagRule `json:"tagRule" yaml:"tagRule"`
}

/**
 * Changes in #115_116: Parse Tags from dynatrace.conf.yaml and only fall back to default behavior if it doesnt exist
 */
func createAttachRules(a adapter.EventContentAdapter, attachRules *AttachRules) AttachRules {
	if attachRules != nil {
		return *attachRules
	}

	ar := AttachRules{
		TagRule: []TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []TagEntry{
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

// CreateInfoEventDTO creates a new Dynatrace CUSTOM_INFO event
func CreateInfoEventDTO(a adapter.EventContentAdapter, attachRules *AttachRules) InfoEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie InfoEvent
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

// CreateAnnotationEventDTO creates a Dynatrace CUSTOM_ANNOTATION event
func CreateAnnotationEventDTO(a adapter.EventContentAdapter, attachRules *AttachRules) AnnotationEvent {

	// we fill the Dynatrace Info Event with values from the labels or use our defaults
	var ie AnnotationEvent
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

// CreateDeploymentEventDTO creates a Dynatrace CUSTOM_DEPLOYMENT event
func CreateDeploymentEventDTO(a adapter.EventContentAdapter, attachRules *AttachRules) DeploymentEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de DeploymentEvent
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

// CreateConfigurationEventDTO creates a Dynatrace CUSTOM_CONFIGURATION event
func CreateConfigurationEventDTO(a adapter.EventContentAdapter, attachRules *AttachRules) ConfigurationEvent {

	// we fill the Dynatrace Deployment Event with values from the labels or use our defaults
	var de ConfigurationEvent
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

type EventsClient struct {
	client ClientInterface
}

// NewEventsClient creates a new EventsClient
func NewEventsClient(client ClientInterface) *EventsClient {
	return &EventsClient{
		client: client,
	}
}

// addEvent sends an event to the Dynatrace events API
func (ec *EventsClient) addEvent(dtEvent interface{}) (string, error) {
	payload, err := json.Marshal(dtEvent)
	if err != nil {
		return "", fmt.Errorf("could not marshal event payload: %v", err)
	}

	body, err := ec.client.Post(eventsPath, payload)
	if err != nil {
		return "", fmt.Errorf("could not create event: %v", err)
	}

	return string(body), nil
}

// addEventAndLog sends an event to the Dynatrace events API and logs errors if necessary
func (ec *EventsClient) addEventAndLog(dtEvent interface{}) {
	log.Info("Sending event to Dynatrace API")
	body, err := ec.addEvent(dtEvent)
	if err != nil {
		log.WithError(err).Error("Failed sending Dynatrace events API request")
		return
	}

	log.WithField("body", body).Debug("Dynatrace API has accepted the event")
}

// AddDeploymentEvent sends a deployment event to the Dynatrace events API
func (ec *EventsClient) AddDeploymentEvent(de DeploymentEvent) {
	ec.addEventAndLog(de)
}

// AddInfoEvent sends an info event to the Dynatrace events API
func (ec *EventsClient) AddInfoEvent(ie InfoEvent) {
	ec.addEventAndLog(ie)
}

// AddAnnotationEvent sends an annotation event to the Dynatrace events API
func (ec *EventsClient) AddAnnotationEvent(ae AnnotationEvent) {
	ec.addEventAndLog(ae)
}

// AddConfigurationEvent sends a configuration event to the Dynatrace events API
func (ec *EventsClient) AddConfigurationEvent(ce ConfigurationEvent) {
	ec.addEventAndLog(ce)
}

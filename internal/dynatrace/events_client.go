package dynatrace

import (
	"context"
	"encoding/json"
	"fmt"

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

type EventsClient struct {
	client ClientInterface
}

// NewEventsClient creates a new EventsClient
func NewEventsClient(client ClientInterface) *EventsClient {
	return &EventsClient{
		client: client,
	}
}

// addEvent sends an event to the Dynatrace events API.
func (ec *EventsClient) addEvent(ctx context.Context, dtEvent interface{}) (string, error) {
	payload, err := json.Marshal(dtEvent)
	if err != nil {
		return "", fmt.Errorf("could not marshal event payload: %v", err)
	}

	body, err := ec.client.Post(ctx, eventsPath, payload)
	if err != nil {
		return "", fmt.Errorf("could not create event: %v", err)
	}

	return string(body), nil
}

// addEventAndLog sends an event to the Dynatrace events API and logs errors if necessary.
func (ec *EventsClient) addEventAndLog(ctx context.Context, dtEvent interface{}) {
	log.Info("Sending event to Dynatrace API")
	body, err := ec.addEvent(ctx, dtEvent)
	if err != nil {
		log.WithError(err).Error("Failed sending Dynatrace events API request")
		return
	}

	log.WithField("body", body).Debug("Dynatrace API has accepted the event")
}

// AddDeploymentEvent sends a deployment event to the Dynatrace events API.
func (ec *EventsClient) AddDeploymentEvent(ctx context.Context, de DeploymentEvent) {
	ec.addEventAndLog(ctx, de)
}

// AddInfoEvent sends an info event to the Dynatrace events API.
func (ec *EventsClient) AddInfoEvent(ctx context.Context, ie InfoEvent) {
	ec.addEventAndLog(ctx, ie)
}

// AddAnnotationEvent sends an annotation event to the Dynatrace events API.
func (ec *EventsClient) AddAnnotationEvent(ctx context.Context, ae AnnotationEvent) {
	ec.addEventAndLog(ctx, ae)
}

// AddConfigurationEvent sends a configuration event to the Dynatrace events API.
func (ec *EventsClient) AddConfigurationEvent(ctx context.Context, ce ConfigurationEvent) {
	ec.addEventAndLog(ctx, ce)
}

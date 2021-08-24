package dynatrace

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

const eventsPath = "/api/v1/events"

type EventsClient struct {
	client *DynatraceHelper
}

// NewEventsClient creates a new EventsClient
func NewEventsClient(client *DynatraceHelper) *EventsClient {
	return &EventsClient{
		client: client,
	}
}

// sendEvent sends an event to the Dynatrace events API
func (ec *EventsClient) sendEvent(dtEvent interface{}) (string, error) {
	payload, err := json.Marshal(dtEvent)
	if err != nil {
		return "", fmt.Errorf("could not marshal event payload: %v", err)
	}

	body, err := ec.client.Post(eventsPath, payload)
	if err != nil {
		return "", fmt.Errorf("could not create event: %v", err)
	}

	return body, nil
}

// SendEvent sends an event to the Dynatrace events API and logs errors if necessary
func (ec *EventsClient) SendEvent(dtEvent interface{}) {
	log.Info("Sending event to Dynatrace API")
	body, err := ec.sendEvent(dtEvent)
	if err != nil {
		log.WithError(err).Error("Failed sending Dynatrace events API request")
		return
	}

	log.WithField("body", body).Debug("Dynatrace API has accepted the event")
}

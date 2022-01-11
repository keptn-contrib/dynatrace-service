package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
)

const errorType = "sh.keptn.log.error"

type ErrorData struct {
	Message       string `json:"message"`
	IntegrationID string `json:"integrationid"`
	Task          string `json:"task,omitempty"`
}

type ErrorEventFactory struct {
	evt           cloudevents.Event
	err           error
	integrationID string
}

func NewErrorEventFactory(event cloudevents.Event, err error, integrationID string) *ErrorEventFactory {
	return &ErrorEventFactory{
		evt:           event,
		err:           err,
		integrationID: integrationID,
	}
}

func (f *ErrorEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	errorData := ErrorData{
		Message:       f.err.Error(),
		IntegrationID: f.integrationID,
		Task:          getTaskFromEvent(f.evt.Type()),
	}

	return adapter.NewCloudEventFactory(
		adapter.NewCloudEventAdapter(f.evt),
		errorType,
		errorData,
	).CreateCloudEvent()
}

func getTaskFromEvent(eventType string) string {
	if !keptnv2.IsTaskEventType(eventType) {
		return eventType
	}

	taskName, _, err := keptnv2.ParseTaskEventType(eventType)
	if err != nil {
		log.WithError(err).Warnf("could not extract task name from event type: %s, will set it to full type", eventType)
		return eventType
	}

	return taskName
}

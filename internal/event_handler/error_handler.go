package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/event"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

type ErrorHandler struct {
	err error
	evt cloudevents.Event
}

func NewErrorHandler(err error, event cloudevents.Event) *ErrorHandler {
	return &ErrorHandler{
		err: err,
		evt: event,
	}
}

func (eh ErrorHandler) HandleEvent() error {
	keptnClient, err := keptn.NewDefaultClient(eh.evt)
	if err != nil {
		log.WithError(err).Error("Could not instantiate Keptn client")
		return err
	}

	uniformClient := keptn.NewDefaultUniformClient()
	integrationID, err := uniformClient.GetIntegrationIDFor(event.GetEventSource())
	if err != nil {
		log.WithError(err).Error("Could not retrieve integration ID from Keptn Uniform")
	}

	log.WithError(eh.err).Debug("Sending error to Keptn Uniform")
	return keptnClient.SendCloudEvent(
		NewErrorEventFactory(eh.evt, eh.err, integrationID))
}

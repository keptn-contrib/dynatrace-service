package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/event"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"
	"github.com/keptn-contrib/dynatrace-service/internal/sli"
)

type ErrorHandler struct {
	err           error
	evt           cloudevents.Event
	uniformClient keptn.UniformClientInterface
}

func NewErrorHandler(err error, event cloudevents.Event, uniformClient keptn.UniformClientInterface) *ErrorHandler {
	return &ErrorHandler{
		err:           err,
		evt:           event,
		uniformClient: uniformClient,
	}
}

func (eh ErrorHandler) HandleEvent() error {
	keptnClient, err := keptn.NewDefaultClient(eh.evt)
	if err != nil {
		log.WithError(err).Error("Could not instantiate Keptn client")
		// no need to continue with sending, will not work anyway
		return err
	}

	switch eh.evt.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return eh.sendErroredConfigureMonitoringFinishedEvent(keptnClient)
	case keptnv2.GetTriggeredEventType(keptnv2.GetSLITaskName):
		return eh.sendErroredGetSLIFinishedEvent(keptnClient)
	default:
		return eh.sendErrorEvent(keptnClient)
	}
}

func (eh ErrorHandler) sendErroredConfigureMonitoringFinishedEvent(keptnClient *keptn.Client) error {
	adapter, err := monitoring.NewConfigureMonitoringAdapterFromEvent(eh.evt)
	if err != nil {
		return eh.sendErrorEvent(keptnClient)
	}
	return keptnClient.SendCloudEvent(monitoring.NewErroredConfigureMonitoringFinishedEventFactory(adapter, eh.err))
}

func (eh ErrorHandler) sendErroredGetSLIFinishedEvent(keptnClient *keptn.Client) error {
	adapter, err := sli.NewGetSLITriggeredAdapterFromEvent(eh.evt)
	if err != nil {
		return eh.sendErrorEvent(keptnClient)
	}
	return keptnClient.SendCloudEvent(sli.NewErroredGetSLIFinishedEventFactory(adapter, nil, eh.err))
}

func (eh ErrorHandler) sendErrorEvent(keptnClient *keptn.Client) error {
	integrationID, err := eh.uniformClient.GetIntegrationIDByName(event.GetEventSource())
	if err != nil {
		log.WithError(err).Error("Could not retrieve integration ID from Keptn Uniform")
		// no need to continue here, message will not show up in Uniform
		return err
	}

	log.WithError(eh.err).Debug("Sending error to Keptn Uniform")
	return keptnClient.SendCloudEvent(
		NewErrorEventFactory(eh.evt, eh.err, integrationID))
}

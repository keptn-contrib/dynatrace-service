package event_handler

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/monitoring"
	"github.com/keptn-contrib/dynatrace-service/internal/sli"
)

// ErrorHandler handles errors by trying to send them to Keptn Uniform.
type ErrorHandler struct {
	err               error
	event             cloudevents.Event
	eventSenderClient keptn.EventSenderClientInterface
	uniformClient     keptn.UniformClientInterface
}

// NewErrorHandler creates a new ErrorHandler for the specified error, event, ClientInterface and UniformClientInterface.
func NewErrorHandler(err error, event cloudevents.Event, eventSenderClient keptn.EventSenderClientInterface, uniformClient keptn.UniformClientInterface) *ErrorHandler {
	return &ErrorHandler{
		err:               err,
		event:             event,
		eventSenderClient: eventSenderClient,
		uniformClient:     uniformClient,
	}
}

// HandleEvent handles errors by sending an error event.
func (eh ErrorHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	switch eh.event.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return eh.sendErroredConfigureMonitoringFinishedEvent(replyCtx, eh.eventSenderClient)
	case keptnv2.GetTriggeredEventType(keptnv2.GetSLITaskName):
		return eh.sendErroredGetSLIFinishedEvent(replyCtx, eh.eventSenderClient)
	default:
		return eh.sendErrorEvent(replyCtx, eh.eventSenderClient)
	}
}

func (eh ErrorHandler) sendErroredConfigureMonitoringFinishedEvent(ctx context.Context, eventSenderClient keptn.EventSenderClientInterface) error {
	adapter, err := monitoring.NewConfigureMonitoringAdapterFromEvent(eh.event)
	if err != nil {
		return eh.sendErrorEvent(ctx, eventSenderClient)
	}
	return eventSenderClient.SendCloudEvent(monitoring.NewErroredConfigureMonitoringFinishedEventFactory(adapter, eh.err))
}

func (eh ErrorHandler) sendErroredGetSLIFinishedEvent(ctx context.Context, eventSenderClient keptn.EventSenderClientInterface) error {
	adapter, err := sli.NewGetSLITriggeredAdapterFromEvent(eh.event)
	if err != nil {
		return eh.sendErrorEvent(ctx, eventSenderClient)
	}
	return eventSenderClient.SendCloudEvent(sli.NewErroredGetSLIFinishedEventFactory(adapter, eh.err))
}

func (eh ErrorHandler) sendErrorEvent(ctx context.Context, eventSenderClient keptn.EventSenderClientInterface) error {
	integrationID, err := eh.uniformClient.GetIntegrationIDByName(ctx, adapter.GetEventSource())
	if err != nil {
		log.WithError(err).Error("Could not retrieve integration ID from Keptn Uniform")
		// no need to continue here, message will not show up in Uniform
		return err
	}

	log.WithError(eh.err).Debug("Sending error to Keptn Uniform")
	return eventSenderClient.SendCloudEvent(
		NewErrorEventFactory(eh.event, eh.err, integrationID))
}

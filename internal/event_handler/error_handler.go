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
	err           error
	event         cloudevents.Event
	keptnClient   keptn.ClientInterface
	uniformClient keptn.UniformClientInterface
}

// NewErrorHandler creates a new ErrorHandler for the specified error, event, ClientInterface and UniformClientInterface.
func NewErrorHandler(err error, event cloudevents.Event, keptnClient keptn.ClientInterface, uniformClient keptn.UniformClientInterface) *ErrorHandler {
	return &ErrorHandler{
		err:           err,
		event:         event,
		keptnClient:   keptnClient,
		uniformClient: uniformClient,
	}
}

// HandleEvent handles errors by sending an error event.
func (eh ErrorHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	switch eh.event.Type() {
	case keptnevents.ConfigureMonitoringEventType:
		return eh.sendErroredConfigureMonitoringFinishedEvent(eh.keptnClient)
	case keptnv2.GetTriggeredEventType(keptnv2.GetSLITaskName):
		return eh.sendErroredGetSLIFinishedEvent(eh.keptnClient)
	default:
		return eh.sendErrorEvent(eh.keptnClient)
	}
}

func (eh ErrorHandler) sendErroredConfigureMonitoringFinishedEvent(keptnClient keptn.ClientInterface) error {
	adapter, err := monitoring.NewConfigureMonitoringAdapterFromEvent(eh.event)
	if err != nil {
		return eh.sendErrorEvent(keptnClient)
	}
	return keptnClient.SendCloudEvent(monitoring.NewErroredConfigureMonitoringFinishedEventFactory(adapter, eh.err))
}

func (eh ErrorHandler) sendErroredGetSLIFinishedEvent(keptnClient keptn.ClientInterface) error {
	adapter, err := sli.NewGetSLITriggeredAdapterFromEvent(eh.event)
	if err != nil {
		return eh.sendErrorEvent(keptnClient)
	}
	return keptnClient.SendCloudEvent(sli.NewErroredGetSLIFinishedEventFactory(adapter, nil, eh.err))
}

func (eh ErrorHandler) sendErrorEvent(keptnClient keptn.ClientInterface) error {
	integrationID, err := eh.uniformClient.GetIntegrationIDByName(adapter.GetEventSource())
	if err != nil {
		log.WithError(err).Error("Could not retrieve integration ID from Keptn Uniform")
		// no need to continue here, message will not show up in Uniform
		return err
	}

	log.WithError(eh.err).Debug("Sending error to Keptn Uniform")
	return keptnClient.SendCloudEvent(
		NewErrorEventFactory(eh.event, eh.err, integrationID))
}

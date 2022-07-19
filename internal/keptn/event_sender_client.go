package keptn

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/go-utils/pkg/sdk/connector/controlplane"
)

// EventSenderClientInterface sends cloud events.
type EventSenderClientInterface interface {
	// SendCloudEvent sends a cloud event from specified factory or returns an error.
	SendCloudEvent(factory adapter.CloudEventFactoryInterface) error
}

// EventSenderClient is an implementation of EventSenderClientInterface.
type EventSenderClient struct {
	eventSender controlplane.EventSender
}

// NewEventSenderClient creates a new EventSenderClient using the specified event sender.
func NewEventSenderClient(eventSender controlplane.EventSender) *EventSenderClient {
	return &EventSenderClient{
		eventSender: eventSender,
	}
}

// SendCloudEvent sends a cloud event from specified factory or returns an error.
func (c *EventSenderClient) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
	ev, err := factory.CreateCloudEvent()
	if err != nil {
		return fmt.Errorf("could not create cloud event: %s", err)
	}

	keptnEvent, err := v0_2_0.ToKeptnEvent(*ev)
	if err != nil {
		return err
	}

	if err := c.eventSender(keptnEvent); err != nil {
		return fmt.Errorf("could not send %s event: %s", ev.Type(), err.Error())
	}

	return nil
}

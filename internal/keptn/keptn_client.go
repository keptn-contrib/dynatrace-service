package keptn

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/go-utils/pkg/sdk/connector/controlplane"
)

// ClientInterface sends cloud events.
type ClientInterface interface {
	// SendCloudEvent sends a cloud event from specified factory or returns an error.
	SendCloudEvent(factory adapter.CloudEventFactoryInterface) error
}

// Client is an implementation of ClientInterface.
type Client struct {
	eventSender controlplane.EventSender
}

// NewClient creates a new Client using the specified event sender and event or returns an error.
func NewClient(eventSender controlplane.EventSender) (*Client, error) {
	return &Client{
		eventSender: eventSender,
	}, nil
}

// SendCloudEvent sends a cloud event from specified factory or returns an error.
func (c *Client) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
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

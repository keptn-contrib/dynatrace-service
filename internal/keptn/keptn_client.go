package keptn

import (
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnapi "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const sliResourceURI = "dynatrace/sli.yaml"

// ClientInterface sends cloud events.
type ClientInterface interface {
	// SendCloudEvent sends a cloud event from specified factory or returns an error.
	SendCloudEvent(factory adapter.CloudEventFactoryInterface) error
}

// Client is an implementation of ClientInterface.
type Client struct {
	client *keptnv2.Keptn
}

// NewDefaultClient creates a new Client using the specified event or returns an error.
func NewDefaultClient(event event.Event) (*Client, error) {
	keptnOpts := keptnapi.KeptnOpts{
		ConfigurationServiceURL: getConfigurationServiceURL(),
		DatastoreURL:            getDatastoreURL(),
	}
	kClient, err := keptnv2.NewKeptn(&event, keptnOpts)
	if err != nil {
		return nil, fmt.Errorf("could not create default Keptn client: %v", err)
	}
	return &Client{
		client: kClient,
	}, nil
}

// SendCloudEvent sends a cloud event from specified factory or returns an error.
func (c *Client) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
	ev, err := factory.CreateCloudEvent()
	if err != nil {
		return fmt.Errorf("could not create cloud event: %s", err)
	}

	if err := c.client.SendCloudEvent(*ev); err != nil {
		return fmt.Errorf("could not send %s event: %s", ev.Type(), err.Error())
	}

	return nil
}

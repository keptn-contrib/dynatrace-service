package keptn

import (
	"errors"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnapi "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const sliResourceURI = "dynatrace/sli.yaml"

type ClientInterface interface {
	GetSLIs(project string, stage string, service string) (map[string]string, error)
	SendCloudEvent(factory adapter.CloudEventFactoryInterface) error
}

type Client struct {
	client *keptnv2.Keptn
}

func NewClient(client *keptnv2.Keptn) *Client {
	return &Client{
		client: client,
	}
}

func NewDefaultClient(event event.Event) (*Client, error) {
	keptnOpts := keptnapi.KeptnOpts{
		ConfigurationServiceURL: getConfigurationServiceURL(),
		DatastoreURL:            getDatastoreURL(),
	}
	kClient, err := keptnv2.NewKeptn(&event, keptnOpts)
	if err != nil {
		return nil, fmt.Errorf("could not create default Keptn client: %v", err)
	}
	return NewClient(kClient), nil
}

func (c *Client) GetSLIs(project string, stage string, service string) (map[string]string, error) {
	if c.client == nil {
		return nil, errors.New("could not retrieve SLI config: no Keptn client initialized")
	}

	slis, err := c.client.GetSLIConfiguration(project, stage, service, sliResourceURI)
	if err != nil {
		return nil, err
	}

	return slis, nil
}

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

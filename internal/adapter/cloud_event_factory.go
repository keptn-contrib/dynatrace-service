package adapter

import cloudevents "github.com/cloudevents/sdk-go/v2"

type CloudEventFactoryInterface interface {
	CreateCloudEvent() (*cloudevents.Event, error)
}

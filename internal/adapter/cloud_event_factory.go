package adapter

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type CloudEventFactoryInterface interface {
	CreateCloudEvent() (*cloudevents.Event, error)
}

type CloudEventFactory struct {
	// TODO 2021-09-07: fix interface below
	event     CloudEventContentAdapter
	eventType string
	payload   interface{}
}

func NewCloudEventFactory(event CloudEventContentAdapter, eventType string, payload interface{}) *CloudEventFactory {
	return &CloudEventFactory{
		event:     event,
		eventType: eventType,
		payload:   payload,
	}
}

func (f *CloudEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	ev := cloudevents.NewEvent()
	ev.SetSource(event.GetEventSource())
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetType(f.eventType)
	ev.SetExtension("shkeptncontext", f.event.GetShKeptnContext())
	ev.SetExtension("triggeredid", f.event.GetEventID())

	err := ev.SetData(cloudevents.ApplicationJSON, f.payload)
	if err != nil {
		return nil, fmt.Errorf("could not marshal cloud event payload: %v", err)
	}

	return &ev, nil
}

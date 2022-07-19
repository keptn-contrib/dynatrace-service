package event_handler

import (
	"context"
	"net/url"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/sli"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

// Test_getEventAdapterForGetSLITriggeredForDynatrace tests that getEventAdapter returns an sli.GetSLITriggeredAdapter and no error for an "sh.keptn.event.get-sli.triggered" event with SLIProvider set to "dynatrace".
func Test_getEventAdapterForGetSLITriggeredForDynatrace(t *testing.T) {
	getSLITriggeredEvent, err := createTestGetSLITriggeredCloudEvent("dynatrace")
	if !assert.NoError(t, err) {
		return
	}

	adapter, err := getEventAdapter(getSLITriggeredEvent)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, adapter) {
		return
	}

	getSLITriggeredAdapter, ok := adapter.(*sli.GetSLITriggeredAdapter)
	assert.True(t, ok)
	assert.NotNil(t, getSLITriggeredAdapter)
}

// Test_getEventAdapterForGetSLITriggeredForNotDynatrace tests that getEventAdapter returns nil and no error for an "sh.keptn.event.get-sli.triggered" event with an SLIProvider other than "dynatrace".
func Test_getEventAdapterForGetSLITriggeredNotForDynatrace(t *testing.T) {
	getSLITriggeredEvent, err := createTestGetSLITriggeredCloudEvent("other")
	if !assert.NoError(t, err) {
		return
	}

	adapter, err := getEventAdapter(getSLITriggeredEvent)
	assert.NoError(t, err)
	assert.Nil(t, adapter)
}

// TestEventHandlerIgnoresGetSLITriggeredNotForDynatrace tests that EventHandler ignores "sh.keptn.event.get-sli.triggered" events with an SLIProvider other than "dynatrace".
func TestEventHandlerIgnoresGetSLITriggeredNotForDynatrace(t *testing.T) {
	getSLITriggeredEvent, err := createTestGetSLITriggeredCloudEvent("other")
	if !assert.NoError(t, err) {
		return
	}

	clientFactory := &clientFactoryMock{
		t: t,
	}

	eventSenderClient := &eventSenderClientMock{
		t: t,
	}

	handler, err := NewEventHandler(context.Background(), clientFactory, eventSenderClient, getSLITriggeredEvent)
	if !assert.NoError(t, err) {
		return
	}

	err = handler.HandleEvent(context.Background(), context.Background())
	assert.NoError(t, err)
}

func createTestGetSLITriggeredCloudEvent(sliProvider string) (cloudevents.Event, error) {
	return createTestCloudEvent("sh.keptn.event.get-sli.triggered", keptnv2.GetSLITriggeredEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "quality-gate",
			Service: "test",
		},
		GetSLI: keptnv2.GetSLI{
			SLIProvider: sliProvider,
			Start:       "2022-07-11T09:00:00.000Z",
			End:         "2022-07-11T09:05:00.000Z",
			Indicators:  []string{"srt"},
		},
	})
}

func createTestCloudEvent(eventType string, payload any) (cloudevents.Event, error) {
	ev := cloudevents.NewEvent()

	source, _ := url.Parse("dynatrace-service")
	ev.SetSource(source.String())

	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetType(eventType)
	ev.SetExtension("shkeptncontext", "")

	err := ev.SetData(cloudevents.ApplicationJSON, payload)

	return ev, err
}

// eventSenderClientMock is an implementation of keptn.EventSenderClientInterface whose methods panic if called.
type eventSenderClientMock struct {
	t *testing.T
}

func (m *eventSenderClientMock) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
	m.t.Fatalf("SendCloudEvent() should not be needed in this mock!")
	return nil
}

// clientFactoryMock is an implementation of keptn.ClientFactoryInterface whose methods panic if called.
type clientFactoryMock struct {
	t *testing.T
}

func (m *clientFactoryMock) CreateEventClient() keptn.EventClientInterface {
	m.t.Fatalf("SendCloudEvent() should not be needed in this mock!")
	return nil
}

func (m *clientFactoryMock) CreateResourceClient() keptn.ResourceClientInterface {
	m.t.Fatalf("SendCloudEvent() should not be needed in this mock!")
	return nil
}

func (m *clientFactoryMock) CreateServiceClient() keptn.ServiceClientInterface {
	m.t.Fatalf("SendCloudEvent() should not be needed in this mock!")
	return nil
}

func (m *clientFactoryMock) CreateUniformClient() keptn.UniformClientInterface {
	m.t.Fatalf("SendCloudEvent() should not be needed in this mock!")
	return nil
}

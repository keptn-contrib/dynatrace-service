package event_handler

import (
	"net/url"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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

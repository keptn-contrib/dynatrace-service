package problem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

func TestProblemEventHandler_HandleEvent(t *testing.T) {

	tests := []struct {
		name                 string
		receivedEvent        *cloudevents.Event
		expectedEmittedEvent *cloudevents.Event
	}{
		{
			name:                 "open problem event",
			receivedEvent:        readCloudEventFromFile("./testdata/open_problem/received_ce.json"),
			expectedEmittedEvent: readCloudEventFromFile("./testdata/open_problem/expected_emitted_ce.json"),
		},
		{
			name:                 "open problem event with tags",
			receivedEvent:        readCloudEventFromFile("./testdata/open_problem_with_tags/received_ce.json"),
			expectedEmittedEvent: readCloudEventFromFile("./testdata/open_problem_with_tags/expected_emitted_ce.json"),
		},
		{
			name:                 "closed problem event",
			receivedEvent:        readCloudEventFromFile("./testdata/closed_problem/received_ce.json"),
			expectedEmittedEvent: readCloudEventFromFile("./testdata/closed_problem/expected_emitted_ce.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			adapter, err := NewProblemAdapterFromEvent(*tt.receivedEvent)
			if !assert.NoError(t, err) {
				return
			}

			kClient := &keptnClientMock{}
			ph := NewProblemEventHandler(adapter, kClient)

			err = ph.HandleEvent()

			assert.NoError(t, err)
			if assert.EqualValues(t, 1, len(kClient.eventSink)) {
				assert.EqualValues(t, tt.expectedEmittedEvent, kClient.eventSink[0])
			}
		})
	}
}

type keptnClientMock struct {
	eventSink []*cloudevents.Event
}

func (m *keptnClientMock) GetCustomQueries(project string, stage string, service string) (*keptn.CustomQueries, error) {
	panic("GetCustomQueries() should not be needed in this mock!")
}

func (m *keptnClientMock) GetShipyard() (*keptnv2.Shipyard, error) {
	panic("GetShipyard() should not be needed in this mock!")
}

func (m *keptnClientMock) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
	// simulate errors while creating cloud event
	if factory == nil {
		return fmt.Errorf("missing factory")
	}

	ce, err := factory.CreateCloudEvent()
	if err != nil {
		return err
	}

	m.eventSink = append(m.eventSink, ce)
	return nil
}

func readCloudEventFromFile(fileName string) *cloudevents.Event {
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic("could not load local test file: " + fileName)
	}

	ce := cloudevents.Event{}
	err = json.Unmarshal(fileContent, &ce)
	if err != nil {
		panic("Cannot make cloud event: " + err.Error())
	}
	return &ce
}

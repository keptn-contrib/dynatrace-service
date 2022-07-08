package action

import (
	"fmt"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/http"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// multiple PGIs found will be returned, when no custom rules are defined
func TestReleaseTriggeredEventHandler_HandleEvent_MultipleEntities(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"multiple_entities.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	expectedAttachRules := getPGIOnlyAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

// multiple PGIs found will be returned, when no custom rules are defined and version information is based on event label
func TestReleaseTriggeredEventHandler_HandleEvent_MultipleEntitiesBasedOnLabel(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"multiple_entities.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	labels := map[string]string{
		"releasesVersion": "1.2.3",
	}

	expectedAttachRules := getPGIOnlyAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              labels,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

// single PGI found will be combined with the custom attach rules
func TestReleaseTriggeredEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"single_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	customAttachRules := getCustomAttachRules()

	expectedAttachRules := getCustomAttachRulesWithEntityIds("PROCESS_GROUP_INSTANCE-D23E64F62FDC200A")

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

// single PGI found will be combined with the custom attach rules that include version information from event label
func TestReleaseTriggeredEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRulesBasedOnLabel(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"single_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	labels := map[string]string{
		"releasesVersion": "1.2.3",
	}

	customAttachRules := getCustomAttachRules()

	expectedAttachRules := getCustomAttachRulesWithEntityIds("PROCESS_GROUP_INSTANCE-D23E64F62FDC200A")

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: expectedAttachRules,
		labels:              labels,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

// no PGIs found and no custom attach rules will result in default attach rules
func TestReleaseTriggeredEventHandler_HandleEvent_NoEntitiesAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

// no deployment.triggered resp. deployment.finished events found and no custom attach rules will result in default attach rules
func TestReleaseTriggeredEventHandler_HandleEvent_NoEventsFoundAndNoCustomAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := getDefaultAttachRules()

	testConfigs := []struct {
		name    string
		eClient *eventClientFake
	}{
		{
			name: "deployment.triggered event not found",
			eClient: &eventClientFake{
				t:           t,
				imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
				eventTimestamps: timestampsForType{
					"sh.keptn.event.deployment.triggered": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.triggered event"),
					},
					"sh.keptn.event.deployment.finished": {
						time: time.Unix(1654000313, 0),
						err:  nil,
					},
				},
			},
		},
		{
			name: "deployment.finished event not found",
			eClient: &eventClientFake{
				t:           t,
				imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
				eventTimestamps: timestampsForType{
					"sh.keptn.event.deployment.triggered": {
						time: time.Unix(1654000240, 0),
						err:  nil,
					},
					"sh.keptn.event.deployment.finished": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.finished event"),
					},
				},
			},
		},
		{
			name: "both deployment events not found",
			eClient: &eventClientFake{
				t:           t,
				imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
				eventTimestamps: timestampsForType{
					"sh.keptn.event.deployment.triggered": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.triggered event"),
					},
					"sh.keptn.event.deployment.finished": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.finished event"),
					},
				},
			},
		},
	}

	for _, testConfig := range testConfigs {
		t.Run(testConfig.name, func(t *testing.T) {

			setup := releaseTriggeredTestSetup{
				t:                   t,
				handler:             handler,
				eClient:             testConfig.eClient,
				customAttachRules:   nil,
				expectedAttachRules: expectedAttachRules,
				labels:              nil,
			}

			assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
		})
	}
}

// no PGIs found but custom attach rules will result custom attach rules only
func TestReleaseTriggeredEventHandler_HandleEvent_NoEntitiesAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: *customAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

// no entities will be queried, because there is no version information. Default attach rules will be returned if there are no custom rules
func TestReleaseTriggeredEventHandler_HandleEvent_NoVersionInformationAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

// no entities will be queried, because there is no version information. Custom attach rules will be returned if they are present
func TestReleaseTriggeredEventHandler_HandleEvent_NoVersionInformationAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForReleaseTriggered(),
	}

	setup := releaseTriggeredTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: *customAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, setup)
}

type releaseTriggeredTestSetup struct {
	t                   *testing.T
	handler             http.Handler
	eClient             *eventClientFake
	customAttachRules   *dynatrace.AttachRules
	expectedAttachRules dynatrace.AttachRules
	labels              map[string]string
}

func (s releaseTriggeredTestSetup) createHandlerAndTeardown() (eventHandler, func()) {
	event := releaseTriggeredEventData{
		baseEventData: baseEventData{
			context: "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
			source:  "jmeter-service",
			event:   "sh.keptn.event.test.finished",
			project: "pod-tato-head",
			stage:   "hardening",
			service: "helloservice",
			labels:  s.labels,
		},
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewReleaseTriggeredEventHandler(&event, client, s.eClient, s.customAttachRules), teardown
}

func (s releaseTriggeredTestSetup) createExpectedDynatraceEvent() dynatrace.InfoEvent {
	tag := s.eClient.imageAndTag.Tag()
	properties := customProperties{
		"Image":         s.eClient.imageAndTag.Image(),
		"Keptn Service": "jmeter-service",
		"KeptnContext":  "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
		"Project":       "pod-tato-head",
		"Service":       "helloservice",
		"Stage":         "hardening",
		"Tag":           tag,
		"TestStrategy":  "",
	}

	addLabelsToProperties(s.t, properties, s.labels)

	return dynatrace.InfoEvent{
		EventType:        "CUSTOM_INFO",
		Source:           "Keptn dynatrace-service",
		Description:      "Release triggered for pod-tato-head, hardening and helloservice",
		Title:            "Release triggered for pod-tato-head, hardening and helloservice",
		CustomProperties: properties,
		AttachRules:      s.expectedAttachRules,
	}
}

func (s releaseTriggeredTestSetup) createEventPayloadContainer() dynatrace.InfoEvent {
	return dynatrace.InfoEvent{}
}

type releaseTriggeredEventData struct {
	baseEventData

	result keptnv2.ResultType
}

func (e releaseTriggeredEventData) GetResult() keptnv2.ResultType {
	if e.result == "" {
		return keptnv2.ResultPass
	}

	return e.result
}

func setupCorrectTimestampResultsForReleaseTriggered() timestampsForType {
	return timestampsForType{
		"sh.keptn.event.deployment.triggered": {
			time: time.Unix(1654000240, 0),
			err:  nil,
		},
		"sh.keptn.event.deployment.finished": {
			time: time.Unix(1654000313, 0),
			err:  nil,
		},
	}
}

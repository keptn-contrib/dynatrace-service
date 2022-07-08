package action

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// multiple PGIs found will be returned, when no custom rules are defined
func TestTestFinishedEventHandler_HandleEvent_MultipleEntities(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"multiple_entities.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	expectedAttachRules := getPGIOnlyAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

// multiple PGIs found will be returned, when no custom rules are defined and version information is based on event label
func TestTestFinishedEventHandler_HandleEvent_MultipleEntitiesBasedOnLabel(t *testing.T) {
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
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              labels,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

// single PGI found will be combined with the custom attach rules
func TestTestFinishedEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"single_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	customAttachRules := getCustomAttachRules()

	expectedAttachRules := getCustomAttachRulesWithEntityIds("PROCESS_GROUP_INSTANCE-D23E64F62FDC200A")

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

// single PGI found will be combined with the custom attach rules that include version information from event label
func TestTestFinishedEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRulesBasedOnLabel(t *testing.T) {
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
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: expectedAttachRules,
		labels:              labels,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

// no PGIs found and no custom attach rules will result in default attach rules
func TestTestFinishedEventHandler_HandleEvent_NoEntitiesAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

// no deployment.triggered resp. deployment.finished events found and no custom attach rules will result in default attach rules
func TestTestFinishedEventHandler_HandleEvent_NoEventsFoundAndNoCustomAttachRules(t *testing.T) {
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

			setup := testFinishedTestSetup{
				t:                   t,
				handler:             handler,
				eClient:             testConfig.eClient,
				customAttachRules:   nil,
				expectedAttachRules: expectedAttachRules,
				labels:              nil,
			}

			assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
		})
	}
}

// no PGIs found but custom attach rules will result custom attach rules only
func TestTestFinishedEventHandler_HandleEvent_NoEntitiesAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: *customAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

// no entities will be queried, because there is no version information. Default attach rules will be returned if there are no custom rules
func TestTestFinishedEventHandler_HandleEvent_NoVersionInformationAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

// no entities will be queried, because there is no version information. Custom attach rules will be returned if they are present
func TestTestFinishedEventHandler_HandleEvent_NoVersionInformationAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForTestFinished(),
	}

	setup := testFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: *customAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, setup)
}

type testFinishedTestSetup struct {
	t                   *testing.T
	handler             http.Handler
	eClient             *eventClientFake
	customAttachRules   *dynatrace.AttachRules
	expectedAttachRules dynatrace.AttachRules
	labels              map[string]string
}

func (s testFinishedTestSetup) createHandlerAndTeardown() (eventHandler, func()) {
	event := baseEventData{
		context: "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
		source:  "jmeter-service",
		event:   "sh.keptn.event.test.finished",
		project: "pod-tato-head",
		stage:   "hardening",
		service: "helloservice",
		labels:  s.labels,
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewTestFinishedEventHandler(&event, client, s.eClient, s.customAttachRules), teardown
}

func (s testFinishedTestSetup) createExpectedDynatraceEvent() dynatrace.AnnotationEvent {
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

	return dynatrace.AnnotationEvent{
		EventType:             "CUSTOM_ANNOTATION",
		Source:                "Keptn dynatrace-service",
		AnnotationType:        "Stop Tests",
		AnnotationDescription: "Stop running tests: against helloservice",
		CustomProperties:      properties,
		AttachRules:           s.expectedAttachRules,
	}
}

func (s testFinishedTestSetup) createEventPayloadContainer() dynatrace.AnnotationEvent {
	return dynatrace.AnnotationEvent{}
}

func setupCorrectTimestampResultsForTestFinished() timestampsForType {
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

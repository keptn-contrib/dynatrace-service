package action

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"net/http"
	"testing"
)

type testTriggeredTestSetup struct {
	t                   *testing.T
	handler             http.Handler
	eClient             *eventClientFake
	customAttachRules   *dynatrace.AttachRules
	expectedAttachRules dynatrace.AttachRules
	labels              map[string]string
}

func (s testTriggeredTestSetup) createHandlerAndTeardown() (eventHandler, func()) {
	event := baseEventData{
		context:      "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
		source:       "jmeter-service",
		event:        "sh.keptn.event.test.finished",
		project:      "pod-tato-head",
		stage:        "hardening",
		service:      "helloservice",
		testStrategy: "performance",
		labels:       s.labels,
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewTestTriggeredEventHandler(&event, client, s.eClient, s.customAttachRules), teardown
}

func (s testTriggeredTestSetup) createExpectedDynatraceEvent() dynatrace.AnnotationEvent {
	tag := s.eClient.imageAndTag.Tag()
	properties := customProperties{
		"Image":         s.eClient.imageAndTag.Image(),
		"Keptn Service": "jmeter-service",
		"KeptnContext":  "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
		"Project":       "pod-tato-head",
		"Service":       "helloservice",
		"Stage":         "hardening",
		"Tag":           tag,
		"TestStrategy":  "performance",
	}

	addLabelsToProperties(s.t, properties, s.labels)

	return dynatrace.AnnotationEvent{
		EventType:             "CUSTOM_ANNOTATION",
		Source:                "Keptn dynatrace-service",
		AnnotationType:        "Start Tests: performance",
		AnnotationDescription: "Start running tests: performance against helloservice",
		CustomProperties:      properties,
		AttachRules:           s.expectedAttachRules,
	}
}

func (s testTriggeredTestSetup) createEventPayloadContainer() dynatrace.AnnotationEvent {
	return dynatrace.AnnotationEvent{}
}

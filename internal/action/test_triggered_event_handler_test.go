package action

import (
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
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
		context:      testKeptnShContext,
		source:       "jmeter-service",
		event:        "sh.keptn.event.test.finished",
		project:      testProject,
		stage:        testStage,
		service:      testService,
		testStrategy: "performance",
		labels:       s.labels,
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewTestTriggeredEventHandler(&event, client, s.eClient, keptn.NewBridgeURLCreator(newKeptnCredentialsProviderMock()), s.customAttachRules), teardown
}

func (s testTriggeredTestSetup) createExpectedDynatraceEvent() dynatrace.AnnotationEvent {
	tag := s.eClient.imageAndTag.Tag()
	properties := customProperties{
		"Image":         s.eClient.imageAndTag.Image(),
		"Keptn Service": "jmeter-service",
		"KeptnContext":  testKeptnShContext,
		"Keptns Bridge": testKeptnsBridge,
		"Project":       testProject,
		"Service":       testService,
		"Stage":         testStage,
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

package action

import (
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

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
			context: testKeptnShContext,
			source:  "jmeter-service",
			event:   "sh.keptn.event.test.finished",
			project: testProject,
			stage:   testStage,
			service: testService,
			labels:  s.labels,
		},
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewReleaseTriggeredEventHandler(&event, client, s.eClient, keptn.NewBridgeURLCreator(newKeptnCredentialsProviderMock()), s.customAttachRules), teardown
}

func (s releaseTriggeredTestSetup) createExpectedDynatraceEvent() dynatrace.InfoEvent {
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

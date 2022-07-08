package action

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/http"
	"testing"
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

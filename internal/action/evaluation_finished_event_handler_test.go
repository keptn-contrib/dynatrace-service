package action

import (
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type evaluationFinishedTestSetup struct {
	t                   *testing.T
	handler             http.Handler
	eClient             *eventClientFake
	customAttachRules   *dynatrace.AttachRules
	expectedAttachRules dynatrace.AttachRules
	labels              map[string]string
}

func (s evaluationFinishedTestSetup) createHandlerAndTeardown() (eventHandler, func()) {
	event := evaluationFinishedEventData{
		baseEventData: baseEventData{
			context: testKeptnShContext,
			source:  "lighthouse-service",
			event:   "sh.keptn.event.evaluation.finished",
			project: testProject,
			stage:   testStage,
			service: testService,
			labels:  s.labels,
		},
		score:     100,
		result:    keptnv2.ResultPass,
		startTime: "2022-05-31T12:30:40.739Z",
		endTime:   "2022-05-31T12:31:53.278Z",
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewEvaluationFinishedEventHandler(&event, client, s.eClient, keptn.NewBridgeURLCreator(newKeptnCredentialsProviderMock()), s.customAttachRules), teardown
}

func (s evaluationFinishedTestSetup) createExpectedDynatraceEvent() dynatrace.InfoEvent {
	properties := customProperties{
		"evaluationHeatmapURL": testEvaluationHeatmapURL,
		"Image":                s.eClient.imageAndTag.Image(),
		"Keptn Service":        "lighthouse-service",
		"KeptnContext":         testKeptnShContext,
		"Keptns Bridge":        testKeptnsBridge,
		"Project":              testProject,
		"Service":              testService,
		"Stage":                testStage,
		"Tag":                  s.eClient.imageAndTag.Tag(),
		"TestStrategy":         "",
	}

	addLabelsToProperties(s.t, properties, s.labels)

	return dynatrace.InfoEvent{
		EventType:        "CUSTOM_INFO",
		Description:      "Quality Gate Result in stage hardening: pass (100.00/100)",
		Title:            "Evaluation result: pass",
		Source:           "Keptn dynatrace-service",
		CustomProperties: properties,
		AttachRules:      s.expectedAttachRules,
	}
}

func (s evaluationFinishedTestSetup) createEventPayloadContainer() dynatrace.InfoEvent {
	return dynatrace.InfoEvent{}
}

type evaluationFinishedEventData struct {
	baseEventData

	score     float64
	result    keptnv2.ResultType
	startTime string
	endTime   string
}

func (e *evaluationFinishedEventData) GetEvaluationScore() float64 {
	return e.score
}
func (e *evaluationFinishedEventData) GetResult() keptnv2.ResultType {
	return e.result
}
func (e *evaluationFinishedEventData) GetStartTime() string {
	return e.startTime
}
func (e *evaluationFinishedEventData) GetEndTime() string {
	return e.endTime
}

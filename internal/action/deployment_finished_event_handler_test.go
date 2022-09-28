package action

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type deploymentFinishedTestSetup struct {
	t                   *testing.T
	handler             http.Handler
	eClient             *eventClientFake
	customAttachRules   *dynatrace.AttachRules
	expectedAttachRules dynatrace.AttachRules
	labels              map[string]string
}

// no deployment.started event was found, so time is reset, but no PGIs found and no custom attach rules will result in default attach rules
func TestDeploymentFinishedEventHandler_HandleEvent_NoEventFoundAndNoCustomAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewNotAvailableImageAndTag(),
		eventTimestamps: timestampsForType{
			"sh.keptn.event.deployment.started": {
				time: time.Time{},
				err:  fmt.Errorf("could not find deployment.started event"),
			},
		},
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: getDefaultAttachRules(),
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

func (s deploymentFinishedTestSetup) createHandlerAndTeardown() (eventHandler, func()) {
	event := deploymentFinishedEventData{
		baseEventData: baseEventData{

			context: testKeptnShContext,
			source:  "helm-service",
			event:   "sh.keptn.event.deployment.finished",
			project: testProject,
			stage:   testStage,
			service: testService,
			labels:  s.labels,
		},
		time: time.Unix(1654000313, 0),
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewDeploymentFinishedEventHandler(&event, client, s.eClient, keptn.NewBridgeURLCreator(newKeptnCredentialsProviderMock()), s.customAttachRules), teardown
}

func (s deploymentFinishedTestSetup) createExpectedDynatraceEvent() dynatrace.DeploymentEvent {
	tag := s.eClient.imageAndTag.Tag()
	properties := customProperties{
		"Image":         s.eClient.imageAndTag.Image(),
		"Keptn Service": "helm-service",
		"KeptnContext":  testKeptnShContext,
		"Keptns Bridge": testKeptnsBridge,
		"Project":       testProject,
		"Service":       testService,
		"Stage":         testStage,
		"Tag":           tag,
		"TestStrategy":  "",
	}

	addLabelsToProperties(s.t, properties, s.labels)

	return dynatrace.DeploymentEvent{
		EventType:         "CUSTOM_DEPLOYMENT",
		Source:            "Keptn dynatrace-service",
		DeploymentName:    "Deploy helloservice " + tag + " with strategy ",
		DeploymentVersion: tag,
		DeploymentProject: testProject,
		CustomProperties:  properties,
		AttachRules:       s.expectedAttachRules,
	}
}

func (s deploymentFinishedTestSetup) createEventPayloadContainer() dynatrace.DeploymentEvent {
	return dynatrace.DeploymentEvent{}
}

type deploymentFinishedEventData struct {
	baseEventData

	time time.Time
}

func (e *deploymentFinishedEventData) GetTime() time.Time {
	return e.time
}

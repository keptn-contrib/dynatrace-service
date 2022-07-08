package action

import (
	"net/http"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// multiple PGIs found will be returned, when no custom rules are defined
func TestDeploymentFinishedEventHandler_HandleEvent_MultipleEntities(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"multiple_entities.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	expectedAttachRules := getPGIOnlyAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

// multiple PGIs found will be returned, when no custom rules are defined and version information is based on event label
func TestDeploymentFinishedEventHandler_HandleEvent_MultipleEntitiesBasedOnLabel(t *testing.T) {
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
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              labels,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

// single PGI found will be combined with the custom attach rules
func TestDeploymentFinishedEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"single_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	customAttachRules := getCustomAttachRules()

	expectedAttachRules := getCustomAttachRulesWithEntityIds("PROCESS_GROUP_INSTANCE-D23E64F62FDC200A")

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

// single PGI found will be combined with the custom attach rules that include version information from event label
func TestDeploymentFinishedEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRulesBasedOnLabel(t *testing.T) {
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
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: expectedAttachRules,
		labels:              labels,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

// no PGIs found and no custom attach rules will result in default attach rules
func TestDeploymentFinishedEventHandler_HandleEvent_NoEntitiesAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

// no PGIs found but custom attach rules will result custom attach rules only
func TestDeploymentFinishedEventHandler_HandleEvent_NoEntitiesAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: *customAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

// no entities will be queried, because there is no version information. Default attach rules will be returned if there are no custom rules
func TestDeploymentFinishedEventHandler_HandleEvent_NoVersionInformationAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   nil,
		expectedAttachRules: expectedAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

// no entities will be queried, because there is no version information. Custom attach rules will be returned if they are present
func TestDeploymentFinishedEventHandler_HandleEvent_NoVersionInformationAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentFinished(),
	}

	setup := deploymentFinishedTestSetup{
		t:                   t,
		handler:             handler,
		eClient:             eClient,
		customAttachRules:   customAttachRules,
		expectedAttachRules: *customAttachRules,
		labels:              nil,
	}

	assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, setup)
}

type deploymentFinishedTestSetup struct {
	t                   *testing.T
	handler             http.Handler
	eClient             *eventClientFake
	customAttachRules   *dynatrace.AttachRules
	expectedAttachRules dynatrace.AttachRules
	labels              map[string]string
}

func (s deploymentFinishedTestSetup) createHandlerAndTeardown() (eventHandler, func()) {
	event := deploymentFinishedEventData{
		baseEventData: baseEventData{

			context: "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
			source:  "helm-service",
			event:   "sh.keptn.event.deployment.finished",
			project: "pod-tato-head",
			stage:   "hardening",
			service: "helloservice",
			labels:  s.labels,
		},
		time: time.Unix(1654000313, 0),
	}

	client, _, teardown := createDynatraceClient(s.t, s.handler)

	return NewDeploymentFinishedEventHandler(&event, client, s.eClient, s.customAttachRules), teardown
}

func (s deploymentFinishedTestSetup) createExpectedDynatraceEvent() dynatrace.DeploymentEvent {
	tag := s.eClient.imageAndTag.Tag()
	properties := customProperties{
		"Image":         s.eClient.imageAndTag.Image(),
		"Keptn Service": "helm-service",
		"KeptnContext":  "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
		"Project":       "pod-tato-head",
		"Service":       "helloservice",
		"Stage":         "hardening",
		"Tag":           tag,
		"TestStrategy":  "",
	}

	addLabelsToProperties(s.t, properties, s.labels)

	return dynatrace.DeploymentEvent{
		EventType:         "CUSTOM_DEPLOYMENT",
		Source:            "Keptn dynatrace-service",
		DeploymentName:    "Deploy helloservice " + tag + " with strategy ",
		DeploymentVersion: tag,
		DeploymentProject: "pod-tato-head",
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

func setupCorrectTimestampResultsForDeploymentFinished() timestampsForType {
	return timestampsForType{
		"sh.keptn.event.deployment.triggered": {
			time: time.Unix(1654000240, 0),
			err:  nil,
		},
	}
}

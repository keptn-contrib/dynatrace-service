package action

import (
	"context"
	"net/http"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const testdataFolder = "./testdata/evaluation_finished/"

// multiple PGIs found will be returned, when no custom rules are defined
func TestEvaluationFinishedEventHandler_HandleEvent_MultipleEntities(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(
		"/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29%2CreleasesVersion%28%221.2.3%22%29&from=1654000240000&to=1654000313000",
		testdataFolder+"multiple_entities.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	expectedAttachRules := dynatrace.AttachRules{
		EntityIds: []string{
			"PROCESS_GROUP_INSTANCE-95C5FBF859599282",
			"PROCESS_GROUP_INSTANCE-D23E64F62FDC200A",
			"PROCESS_GROUP_INSTANCE-DE323A8B8449D009",
			"PROCESS_GROUP_INSTANCE-F59D42FEA235E5F9",
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, nil, expectedAttachRules, nil)
}

// multiple PGIs found will be returned, when no custom rules are defined and version information is based on event label
func TestEvaluationFinishedEventHandler_HandleEvent_MultipleEntitiesBasedOnLabel(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(
		"/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29%2CreleasesVersion%28%221.2.4%22%29&from=1654000240000&to=1654000313000",
		testdataFolder+"multiple_entities.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	labels := map[string]string{
		"releasesVersion": "1.2.4",
	}

	expectedAttachRules := dynatrace.AttachRules{
		EntityIds: []string{
			"PROCESS_GROUP_INSTANCE-95C5FBF859599282",
			"PROCESS_GROUP_INSTANCE-D23E64F62FDC200A",
			"PROCESS_GROUP_INSTANCE-DE323A8B8449D009",
			"PROCESS_GROUP_INSTANCE-F59D42FEA235E5F9",
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewNotAvailableImageAndTag(),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, nil, expectedAttachRules, labels)
}

// single PGI found will be combined with the custom attach rules
func TestEvaluationFinishedEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(
		"/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29%2CreleasesVersion%28%221.2.3%22%29&from=1654000240000&to=1654000313000",
		testdataFolder+"single_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_multiple_200.json")

	customAttachRules := &dynatrace.AttachRules{
		EntityIds: []string{"PROCESS_GROUP-XXXXXXXXXXXXXXXXX"},
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "my-tag",
						Value:   "my-value",
					},
				},
			},
		},
	}

	expectedAttachRules := dynatrace.AttachRules{
		EntityIds: []string{
			"PROCESS_GROUP-XXXXXXXXXXXXXXXXX",
			"PROCESS_GROUP_INSTANCE-D23E64F62FDC200A",
		},
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "my-tag",
						Value:   "my-value",
					},
				},
			},
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, customAttachRules, expectedAttachRules, nil)
}

// single PGI found will be combined with the custom attach rules that include version information from event label
func TestEvaluationFinishedEventHandler_HandleEvent_SingleEntityAndUserSpecifiedAttachRulesBasedOnLabel(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(
		"/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29%2CreleasesVersion%28%221.2.4%22%29&from=1654000240000&to=1654000313000",
		testdataFolder+"single_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	labels := map[string]string{
		"releasesVersion": "1.2.4",
	}

	customAttachRules := &dynatrace.AttachRules{
		EntityIds: []string{
			"PROCESS_GROUP-XXXXXXXXXXXXXXXXX",
		},
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "my-tag",
						Value:   "my-value",
					},
				},
			},
		},
	}

	expectedAttachRules := dynatrace.AttachRules{
		EntityIds: []string{
			"PROCESS_GROUP-XXXXXXXXXXXXXXXXX",
			"PROCESS_GROUP_INSTANCE-D23E64F62FDC200A",
		},
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "my-tag",
						Value:   "my-value",
					},
				},
			},
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewNotAvailableImageAndTag(),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, customAttachRules, expectedAttachRules, labels)
}

// no PGIs found and no custom attach rules will result in default attach rules
func TestEvaluationFinishedEventHandler_HandleEvent_NoEntitiesAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(
		"/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29%2CreleasesVersion%28%221.2.3%22%29&from=1654000240000&to=1654000313000",
		testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := dynatrace.AttachRules{
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   "pod-tato-head",
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   "hardening",
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   "helloservice",
					},
				},
			},
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, nil, expectedAttachRules, nil)
}

// no PGIs found but custom attach rules will result custom attach rules only
func TestEvaluationFinishedEventHandler_HandleEvent_NoEntitiesAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(
		"/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29%2CreleasesVersion%28%221.2.3%22%29&from=1654000240000&to=1654000313000",
		testdataFolder+"no_entity.json")
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := &dynatrace.AttachRules{
		EntityIds: []string{"PROCESS_GROUP-XXXXXXXXXXXXXXXXX"},
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "my-tag",
						Value:   "my-value",
					},
				},
			},
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, customAttachRules, *customAttachRules, nil)
}

// no entities will be queried, because there is no version information. Default attach rules will be returned if there are no custom rules
func TestEvaluationFinishedEventHandler_HandleEvent_NoVersionInformationAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	expectedAttachRules := dynatrace.AttachRules{
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   "pod-tato-head",
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   "hardening",
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   "helloservice",
					},
				},
			},
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewNotAvailableImageAndTag(),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, nil, expectedAttachRules, nil)
}

// no entities will be queried, because there is no version information. Custom attach rules will be returned if they are present
func TestEvaluationFinishedEventHandler_HandleEvent_NoVersionInformationAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", testdataFolder+"events_response_single_200.json")

	customAttachRules := &dynatrace.AttachRules{
		EntityIds: []string{"PROCESS_GROUP-XXXXXXXXXXXXXXXXX"},
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "my-tag",
						Value:   "my-value",
					},
				},
			},
		},
	}

	eClient := &eventClientFake{
		t:           t,
		imageAndTag: common.NewNotAvailableImageAndTag(),
	}

	assertThatCorrectEventWasSent(t, handler, eClient, customAttachRules, *customAttachRules, nil)
}

func assertThatCorrectEventWasSent(t *testing.T, handler *test.FileBasedURLHandlerWithSink, eClient *eventClientFake, customAttachRules *dynatrace.AttachRules, expectedAttachRules dynatrace.AttachRules, labels map[string]string) {
	eventHandler, teardown := createEvaluationFinishedEventHandler(t, handler, eClient, customAttachRules, labels)
	defer teardown()

	err := eventHandler.HandleEvent(context.Background(), context.Background())
	assert.NoError(t, err)

	infoEvent := dynatrace.InfoEvent{}
	handler.GetStoredPayloadForURL("/api/v1/events", &infoEvent)

	assert.EqualValues(t, createExpectedInfoEvent(t, expectedAttachRules, eClient.imageAndTag, labels), infoEvent)
}

func createEvaluationFinishedEventHandler(t *testing.T, handler http.Handler, eClient keptn.EventClientInterface, attachRules *dynatrace.AttachRules, labels map[string]string) (*EvaluationFinishedEventHandler, func()) {
	event := evaluationFinishedEventData{
		context:   "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
		source:    "lighthouse-service",
		event:     "sh.keptn.event.evaluation.finished",
		project:   "pod-tato-head",
		stage:     "hardening",
		service:   "helloservice",
		score:     100,
		result:    keptnv2.ResultPass,
		startTime: "2022-05-31T12:30:40.739Z",
		endTime:   "2022-05-31T12:31:53.278Z",
		labels:    labels,
	}

	client, _, teardown := createDynatraceClient(t, handler)

	return NewEvaluationFinishedEventHandler(&event, client, eClient, attachRules), teardown
}

func createExpectedInfoEvent(t *testing.T, attachRules dynatrace.AttachRules, imageAndTag common.ImageAndTag, labels map[string]string) dynatrace.InfoEvent {
	properties := map[string]string{
		"Image":         imageAndTag.Image(),
		"Keptn Service": "lighthouse-service",
		"KeptnContext":  "7c2c890f-b3ac-4caa-8922-f44d2aa54ec9",
		"Project":       "pod-tato-head",
		"Service":       "helloservice",
		"Stage":         "hardening",
		"Tag":           imageAndTag.Tag(),
		"TestStrategy":  "",
	}

	for key, value := range labels {
		if old, ok := properties[key]; ok {
			t.Errorf("Overwriting old value '%s' for key '%s' in properties map with new value '%s'", old, key, value)
		}

		properties[key] = value
	}

	return dynatrace.InfoEvent{
		EventType:        "CUSTOM_INFO",
		Description:      "Quality Gate Result in stage hardening: pass (100.00/100)",
		Title:            "Evaluation result: pass",
		Source:           "Keptn dynatrace-service",
		CustomProperties: properties,
		AttachRules:      attachRules,
	}
}

type eventClientFake struct {
	t                         *testing.T
	isPartOfRemediation       bool
	remediationDetectionError error
	problemID                 string
	problemIDError            error
	imageAndTag               common.ImageAndTag
}

func (e *eventClientFake) IsPartOfRemediation(_ adapter.EventContentAdapter) (bool, error) {
	return e.isPartOfRemediation, e.remediationDetectionError
}

func (e *eventClientFake) FindProblemID(_ adapter.EventContentAdapter) (string, error) {
	return e.problemID, e.problemIDError
}

func (e *eventClientFake) GetImageAndTag(_ adapter.EventContentAdapter) common.ImageAndTag {
	return e.imageAndTag
}

type evaluationFinishedEventData struct {
	context string
	source  string
	event   string

	project            string
	stage              string
	service            string
	deployment         string
	testStrategy       string
	deploymentStrategy string

	labels map[string]string

	score     float64
	result    keptnv2.ResultType
	startTime string
	endTime   string
}

func (e *evaluationFinishedEventData) GetShKeptnContext() string {
	return e.context
}

func (e *evaluationFinishedEventData) GetEvent() string {
	return e.event
}

func (e *evaluationFinishedEventData) GetSource() string {
	return e.source
}

func (e *evaluationFinishedEventData) GetProject() string {
	return e.project
}

func (e *evaluationFinishedEventData) GetStage() string {
	return e.stage
}

func (e *evaluationFinishedEventData) GetService() string {
	return e.service
}

func (e *evaluationFinishedEventData) GetDeployment() string {
	return e.deployment
}

func (e *evaluationFinishedEventData) GetTestStrategy() string {
	return e.testStrategy
}

func (e *evaluationFinishedEventData) GetDeploymentStrategy() string {
	return e.deploymentStrategy
}

func (e *evaluationFinishedEventData) GetLabels() map[string]string {
	return e.labels
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

package action

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const testDynatraceAPIToken = "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"

const testdataFolder = "./testdata/attach_rules/"

func getPGIQuery(version string) string {
	return "/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29%2CreleasesVersion%28%22" + version + "%22%29&from=1654000240000&to=1654000313000"
}

func getDefaultPGIQuery() string {
	return getPGIQuery("1.2.3")
}

func getDefaultAttachRules() dynatrace.AttachRules {
	return dynatrace.AttachRules{
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
}

func getCustomAttachRules() *dynatrace.AttachRules {
	return &dynatrace.AttachRules{
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
}

func getCustomAttachRulesWithEntityIds(entityIds ...string) dynatrace.AttachRules {
	attachRules := getCustomAttachRules()
	attachRules.EntityIds = append(attachRules.EntityIds, entityIds...)
	return *attachRules
}

func getPGIOnlyAttachRules() dynatrace.AttachRules {
	return dynatrace.AttachRules{
		EntityIds: []string{
			"PROCESS_GROUP_INSTANCE-95C5FBF859599282",
			"PROCESS_GROUP_INSTANCE-D23E64F62FDC200A",
			"PROCESS_GROUP_INSTANCE-DE323A8B8449D009",
			"PROCESS_GROUP_INSTANCE-F59D42FEA235E5F9",
		},
	}
}

// mimic the event_handler.DynatraceEventHandler interface to avoid circular dependencies
type eventHandler interface {
	HandleEvent(workCtx context.Context, replyCtx context.Context) error
}

type dynatraceEvent interface {
	dynatrace.InfoEvent | dynatrace.AnnotationEvent | dynatrace.DeploymentEvent
}

type testSetup[E dynatraceEvent] interface {
	createHandlerAndTeardown() (eventHandler, func())
	createExpectedDynatraceEvent() E
	createEventPayloadContainer() E
}

func assertThatCorrectEventWasSent[E dynatraceEvent](t *testing.T, handler *test.FileBasedURLHandlerWithSink, setup testSetup[E]) {
	eventHandler, teardown := setup.createHandlerAndTeardown()
	defer teardown()

	err := eventHandler.HandleEvent(context.Background(), context.Background())
	assert.NoError(t, err)

	dynatraceEvent := setup.createEventPayloadContainer()
	handler.GetStoredPayloadForURL("/api/v1/events", &dynatraceEvent)

	assert.EqualValues(t, setup.createExpectedDynatraceEvent(), dynatraceEvent)
}

func createDynatraceClient(t *testing.T, handler http.Handler) (dynatrace.ClientInterface, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dh := dynatrace.NewClientWithHTTP(createDynatraceCredentials(t, url), httpClient)

	return dh, url, teardown
}

func createDynatraceCredentials(t *testing.T, url string) *credentials.DynatraceCredentials {
	dynatraceCredentials, err := credentials.NewDynatraceCredentials(url, testDynatraceAPIToken)
	assert.NoError(t, err)
	return dynatraceCredentials
}

func addLabelsToProperties(t *testing.T, properties customProperties, labels map[string]string) {
	for key, value := range labels {
		if old, ok := properties[key]; ok {
			t.Errorf("Overwriting old value '%s' for key '%s' in properties map with new value '%s'", old, key, value)
		}

		properties[key] = value
	}
}

type eventClientFake struct {
	t                         *testing.T
	isPartOfRemediation       bool
	remediationDetectionError error
	problemID                 string
	problemIDError            error
	imageAndTag               common.ImageAndTag
	eventTimestamps           timestampsForType
}

type timestampsForType map[string]struct {
	time time.Time
	err  error
}

func (e *eventClientFake) IsPartOfRemediation(_ context.Context, _ adapter.EventContentAdapter) (bool, error) {
	return e.isPartOfRemediation, e.remediationDetectionError
}

func (e *eventClientFake) FindProblemID(_ context.Context, _ adapter.EventContentAdapter) (string, error) {
	return e.problemID, e.problemIDError
}

func (e *eventClientFake) GetImageAndTag(_ context.Context, _ adapter.EventContentAdapter) common.ImageAndTag {
	return e.imageAndTag
}

func (e *eventClientFake) GetEventTimeStampForType(_ context.Context, _ adapter.EventContentAdapter, eventType string) (*time.Time, error) {
	result, found := e.eventTimestamps[eventType]
	if !found {
		e.t.Errorf("Could find entry for event type: %s - fix test setup!", eventType)
	}

	if result.err == nil {
		return &result.time, nil
	}
	return nil, result.err
}

type baseEventData struct {
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
}

func (e *baseEventData) GetShKeptnContext() string {
	return e.context
}

func (e *baseEventData) GetEvent() string {
	return e.event
}

func (e *baseEventData) GetSource() string {
	return e.source
}

func (e *baseEventData) GetProject() string {
	return e.project
}

func (e *baseEventData) GetStage() string {
	return e.stage
}

func (e *baseEventData) GetService() string {
	return e.service
}

func (e *baseEventData) GetDeployment() string {
	return e.deployment
}

func (e *baseEventData) GetTestStrategy() string {
	return e.testStrategy
}

func (e *baseEventData) GetDeploymentStrategy() string {
	return e.deploymentStrategy
}

func (e *baseEventData) GetLabels() map[string]string {
	return e.labels
}

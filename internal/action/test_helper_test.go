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

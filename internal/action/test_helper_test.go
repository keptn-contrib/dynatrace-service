package action

import (
	"context"
	"net/http"
	"testing"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const testDynatraceAPIToken = "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"

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

type eventClientFake struct {
	t                         *testing.T
	isPartOfRemediation       bool
	remediationDetectionError error
	problemID                 string
	problemIDError            error
	imageAndTag               common.ImageAndTag
	eventTimestamp            *time.Time
	eventTimestampError       error
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

func (e *eventClientFake) GetEventTimeStampForType(_ context.Context, _ adapter.EventContentAdapter, _ string) (*time.Time, error) {
	return e.eventTimestamp, e.eventTimestampError
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

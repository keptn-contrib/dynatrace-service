package sli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const indicator = "response_time_p95"
const testDynatraceAPIToken = "dtOc01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"
const testDashboardID = "12345678-1111-4444-8888-123456789012"

var testGetSLIEventDataWithDefaultStartAndEnd = createTestGetSLIEventDataWithStartAndEnd("", "")

var getSLIFinishedEventSuccessAssertionsFunc = func(t *testing.T, data *keptnv2.GetSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultPass, data.Result)
	assert.Empty(t, data.Message)
}

var getSLIFinishedEventWarningAssertionsFunc = func(t *testing.T, data *keptnv2.GetSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultWarning, data.Result)
	assert.NotEmpty(t, data.Message)
}

var getSLIFinishedEventFailureAssertionsFunc = func(t *testing.T, data *keptnv2.GetSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
	assert.NotEmpty(t, data.Message)
}

func createTestGetSLIEventDataWithStartAndEnd(sliStart string, sliEnd string) *getSLIEventData {
	return &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
		sliStart:   sliStart,
		sliEnd:     sliEnd,
	}
}

func runTestAndAssertNoError(t *testing.T, ev *getSLIEventData, handler http.Handler, kClient *keptnClientMock, rClient keptn.SLOAndSLIClientInterface, dashboard string) {
	eh, _, teardown := createGetSLIEventHandler(t, ev, handler, kClient, rClient, dashboard)
	defer teardown()

	assert.NoError(t, eh.HandleEvent(context.Background(), context.Background()))
}

func assertCorrectGetSLIEvents(t *testing.T, events []*cloudevents.Event, getSLIFinishedEventAssertionsFunc func(*testing.T, *keptnv2.GetSLIFinishedEventData), sliResultAssertionsFuncs ...func(*testing.T, *keptnv2.SLIResult)) {
	assert.EqualValues(t, 2, len(events))

	assert.EqualValues(t, keptnv2.GetStartedEventType(keptnv2.GetSLITaskName), events[0].Type())
	assert.EqualValues(t, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), events[1].Type())

	var data keptnv2.GetSLIFinishedEventData
	err := json.Unmarshal(events[1].Data(), &data)
	if err != nil {
		t.Fatalf("could not parse event payload correctly: %s", err)
	}

	getSLIFinishedEventAssertionsFunc(t, &data)

	assert.EqualValues(t, keptnv2.StatusSucceeded, data.Status)

	assertCorrectSLIResults(t, &data, sliResultAssertionsFuncs...)
}

func assertCorrectSLIResults(t *testing.T, getSLIFinishedEventData *keptnv2.GetSLIFinishedEventData, sliResultAssertionsFuncs ...func(t *testing.T, actual *keptnv2.SLIResult)) {
	if !assert.EqualValues(t, len(sliResultAssertionsFuncs), len(getSLIFinishedEventData.GetSLI.IndicatorValues), "number of assertions should match number of SLI indicator values") {
		return
	}
	for i, assertionsFunction := range sliResultAssertionsFuncs {
		assertionsFunction(t, getSLIFinishedEventData.GetSLI.IndicatorValues[i])
	}
}

func createSLIAssertionsFunc(expectedMetric string, expectedDefintion string) func(t *testing.T, actualMetric string, actualDefinition string) {
	return func(t *testing.T, actualMetric string, actualDefinition string) {
		assert.EqualValues(t, expectedMetric, actualMetric)
		assert.EqualValues(t, expectedDefintion, actualDefinition)
	}
}

func createSuccessfulSLIResultAssertionsFunc(expectedMetric string, expectedValue float64) func(t *testing.T, actual *keptnv2.SLIResult) {
	return func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, expectedMetric, actual.Metric, "Indicator metric should match")
		assert.EqualValues(t, expectedValue, actual.Value, "Indicator values should match")
		assert.True(t, actual.Success, "Indicator success should be true")
	}
}

func createFailedSLIResultAssertionsFunc(expectedMetric string, expectedMessageSubstrings ...string) func(*testing.T, *keptnv2.SLIResult) {
	return func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.False(t, actual.Success, "Indicator success should be false")
		assert.EqualValues(t, expectedMetric, actual.Metric, "Indicator metric should match")
		assert.Zero(t, actual.Value, "Indicator value should be zero")

		for _, expectedSubstring := range expectedMessageSubstrings {
			assert.Contains(t, actual.Message, expectedSubstring, "all substrings should be contained in message")
		}
	}
}

func createGetSLIEventHandler(t *testing.T, keptnEvent GetSLITriggeredAdapterInterface, handler http.Handler, eventSenderClient keptn.EventSenderClientInterface, rClient keptn.SLOAndSLIClientInterface, dashboard string) (*GetSLIEventHandler, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials, err := credentials.NewDynatraceCredentials(url, testDynatraceAPIToken)
	assert.NoError(t, err)

	eh := &GetSLIEventHandler{
		event:             keptnEvent,
		dtClient:          dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		eventSenderClient: eventSenderClient,
		resourceClient:    rClient,
		dashboard:         dashboard,
		secretName:        "dynatrace", // we do not need this string
	}

	return eh, url, teardown
}

type getSLIEventData struct {
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

	indicators      []string
	customFilters   []*keptnv2.SLIFilter
	notForDynatrace bool
	sliStart        string
	sliEnd          string
}

func (e *getSLIEventData) GetShKeptnContext() string {
	return e.context
}

func (e *getSLIEventData) GetEvent() string {
	return e.event
}

func (e *getSLIEventData) GetSource() string {
	return e.source
}

func (e *getSLIEventData) GetProject() string {
	return e.project
}

func (e *getSLIEventData) GetStage() string {
	return e.stage
}

func (e *getSLIEventData) GetService() string {
	return e.service
}

func (e *getSLIEventData) GetDeployment() string {
	return e.deployment
}

func (e *getSLIEventData) GetTestStrategy() string {
	return e.testStrategy
}

func (e *getSLIEventData) GetDeploymentStrategy() string {
	return e.deploymentStrategy
}

func (e *getSLIEventData) GetLabels() map[string]string {
	return e.labels
}

func (e *getSLIEventData) GetEventID() string {
	return "some-event-id"
}

func (e *getSLIEventData) IsNotForDynatrace() bool {
	return e.notForDynatrace
}

func (e *getSLIEventData) GetSLIStart() string {
	if e.sliStart == "" {
		return "2021-09-28T13:16:39.000Z"
	}

	return e.sliStart
}

func (e *getSLIEventData) GetSLIEnd() string {
	if e.sliEnd == "" {
		return "2021-09-28T13:21:39.000Z"
	}

	return e.sliEnd
}

func (e *getSLIEventData) GetIndicators() []string {
	return e.indicators
}

func (e *getSLIEventData) GetCustomSLIFilters() []*keptnv2.SLIFilter {
	return e.customFilters
}

func (e *getSLIEventData) AddLabel(name string, value string) {
	if e.labels == nil {
		e.labels = make(map[string]string)
	}

	e.labels[name] = value
}

type resourceClientMock struct {
	t            *testing.T
	slis         map[string]string
	getSLIsError error
}

func newResourceClientMock(t *testing.T) *resourceClientMock {
	return &resourceClientMock{
		t: t,
	}
}

func newResourceClientMockWithSLIs(t *testing.T, slis map[string]string) *resourceClientMock {
	return &resourceClientMock{
		t:    t,
		slis: slis,
	}
}

func newResourceClientMockWithGetSLIsError(t *testing.T, getSLIsError error) *resourceClientMock {
	return &resourceClientMock{
		t:            t,
		getSLIsError: getSLIsError,
	}
}

func (m *resourceClientMock) GetSLIs(_ context.Context, _ string, _ string, _ string) (map[string]string, error) {
	if m.getSLIsError != nil {
		return nil, m.getSLIsError
	}

	return m.slis, nil
}

func (m *resourceClientMock) GetSLOs(_ context.Context, _ string, _ string, _ string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")
	return nil, nil
}

func (m *resourceClientMock) UploadSLIs(_ context.Context, _ string, _ string, _ string, _ *dynatrace.SLI) error {
	m.t.Fatalf("UploadSLIs() should not be needed in this mock!")
	return nil
}

func (m *resourceClientMock) UploadSLOs(_ context.Context, _ string, _ string, _ string, _ *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLOs() should not be needed in this mock!")
	return nil
}

type keptnClientMock struct {
	eventSink []*cloudevents.Event
}

func (m *keptnClientMock) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
	// simulate errors while creating cloud event
	if factory == nil {
		return fmt.Errorf("could not send create cloud event")
	}

	ce, err := factory.CreateCloudEvent()
	if err != nil {
		panic("could not create cloud event: " + err.Error())
	}

	m.eventSink = append(m.eventSink, ce)

	return nil
}

package sli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn/go-utils/pkg/common/timeutils"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const testIndicatorResponseTimeP95 = "response_time_p95"
const testDynatraceAPIToken = "dtOc01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"
const testDashboardID = "12345678-1111-4444-8888-123456789012"
const testSLIStart = "2022-09-28T00:00:00.000Z"
const testSLIEnd = "2022-09-29T00:00:00.000Z"

const resolutionInf = "Inf"
const resolutionIsNullKeyValuePair = "resolution=null&"
const singleValueVisualConfigType = "SINGLE_VALUE"
const graphChartVisualConfigType = "GRAPH_CHART"

var testGetSLIEventData = createTestGetSLIEventDataWithIndicators([]string{testIndicatorResponseTimeP95})

var getSLIFinishedEventSuccessAssertionsFunc = func(t *testing.T, data *getSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultPass, data.Result)
	assert.Empty(t, data.Message)
}

var getSLIFinishedEventWarningAssertionsFunc = func(t *testing.T, data *getSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultWarning, data.Result)
	assert.NotEmpty(t, data.Message)
}

var getSLIFinishedEventFailureAssertionsFunc = func(t *testing.T, data *getSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
	assert.NotEmpty(t, data.Message)
}

func createTestGetSLIEventDataWithIndicators(indicators []string) *getSLIEventData {
	return &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: indicators, // we need this to check later on in the custom queries
		sliStart:   testSLIStart,
		sliEnd:     testSLIEnd,
	}
}

// convertTimeStringToUnixMillisecondsString converts a ISO8601 (or fallback format RFC3339) time string to a string with the Unix timestamp, or "0" if this cannot be done.
func convertTimeStringToUnixMillisecondsString(timeString string) string {
	time, err := timeutils.ParseTimestamp(timeString)
	if err != nil {
		return "0"
	}

	return common.TimestampToUnixMillisecondsString(*time)
}

// buildMetricsV2DefinitionRequestString builds a Metrics v2 definition request string with the specified metric ID for use in testing.
func buildMetricsV2DefinitionRequestString(metricID string) string {
	return fmt.Sprintf("%s/%s", dynatrace.MetricsPath, url.PathEscape(metricID))
}

// buildMetricsV2QueryRequestStringWithResolutionInf builds a Metrics v2 request string with the specified metric selector and resolution inf for use in testing.
func buildMetricsV2QueryRequestStringWithResolutionInf(metricSelector string) string {
	return buildMetricsV2QueryRequestStringWithResolution(metricSelector, metrics.ResolutionInf)
}

// buildMetricsV2QueryRequestStringWithResolution builds a Metrics v2 request string with the specified metric selector and resolution for use in testing.
func buildMetricsV2QueryRequestStringWithResolution(metricSelector string, resolution string) string {
	return fmt.Sprintf("%s?from=%s&metricSelector=%s&resolution=%s&to=%s", dynatrace.MetricsQueryPath, convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(metricSelector), resolution, convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildMetricsV2QueryRequestString builds a Metrics v2 request string with the specified metric selector for use in testing.
func buildMetricsV2QueryRequestString(metricSelector string) string {
	return fmt.Sprintf("%s?from=%s&metricSelector=%s&to=%s", dynatrace.MetricsQueryPath, convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(metricSelector), convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildMetricsV2QueryRequestStringWithEntitySelectorAndResolutionInf builds a Metrics v2 request string with the specified entity and metric selectors and resolution inf for use in testing.
func buildMetricsV2QueryRequestStringWithEntitySelectorAndResolutionInf(entitySelector string, metricSelector string) string {
	return buildMetricsV2QueryRequestStringWithEntitySelectorAndResolution(entitySelector, metricSelector, metrics.ResolutionInf)
}

// buildMetricsV2QueryRequestStringWithEntitySelectorAndResolution builds a Metrics v2 request string with the specified entity and metric selectors and resolution for use in testing.
func buildMetricsV2QueryRequestStringWithEntitySelectorAndResolution(entitySelector string, metricSelector string, resolution string) string {
	return fmt.Sprintf("%s?entitySelector=%s&from=%s&metricSelector=%s&resolution=%s&to=%s", dynatrace.MetricsQueryPath, url.QueryEscape(entitySelector), convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(metricSelector), resolution, convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildMetricsV2QueryRequestStringWithEntitySelector builds a Metrics v2 request string with the specified entity and metric selectors for use in testing.
func buildMetricsV2QueryRequestStringWithEntitySelector(entitySelector string, metricSelector string) string {
	return fmt.Sprintf("%s?entitySelector=%s&from=%s&metricSelector=%s&to=%s", dynatrace.MetricsQueryPath, url.QueryEscape(entitySelector), convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(metricSelector), convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildMetricsV2QueryRequestStringWithMZSelectorAndResolutionInf builds a Metrics v2 request string with the specified metric and management zone selectors and resolution inf for use in testing.
func buildMetricsV2QueryRequestStringWithMZSelectorAndResolutionInf(metricSelector string, mzSelector string) string {
	return buildMetricsV2QueryRequestStringWithMZSelectorAndResolution(metricSelector, mzSelector, metrics.ResolutionInf)
}

// buildMetricsV2QueryRequestStringWithMZSelectorAndResolution builds a Metrics v2 request string with the specified metric and management zone selectors and resolution for use in testing.
func buildMetricsV2QueryRequestStringWithMZSelectorAndResolution(metricSelector string, mzSelector string, resolution string) string {
	return fmt.Sprintf("%s?from=%s&metricSelector=%s&mzSelector=%s&resolution=%s&to=%s", dynatrace.MetricsQueryPath, convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(metricSelector), url.QueryEscape(mzSelector), resolution, convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildMetricsV2QueryRequestStringWithMZSelector builds a Metrics v2 request string with the specified metric and management zone selectors for use in testing.
func buildMetricsV2QueryRequestStringWithMZSelector(metricSelector string, mzSelector string) string {
	return fmt.Sprintf("%s?from=%s&metricSelector=%s&mzSelector=%s&to=%s", dynatrace.MetricsQueryPath, convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(metricSelector), url.QueryEscape(mzSelector), convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildProblemsV2Request builds a Problems V2 request string with the specified problem selector for use in testing.
func buildProblemsV2Request(problemSelector string) string {
	return fmt.Sprintf("%s?from=%s&problemSelector=%s&to=%s", dynatrace.ProblemsV2Path, convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(problemSelector), convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildSecurityProblemsRequest builds a Security Problems request string with the specified security problem selector for use in testing.
func buildSecurityProblemsRequest(securityProblemSelector string) string {
	return fmt.Sprintf("%s?from=%s&securityProblemSelector=%s&to=%s", dynatrace.SecurityProblemsPath, convertTimeStringToUnixMillisecondsString(testSLIStart), url.QueryEscape(securityProblemSelector), convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildSLORequest builds a SLO request string with the specified SLO ID for use in testing.
func buildSLORequest(sloID string) string {
	return fmt.Sprintf("%s/%s?from=%s&timeFrame=GTF&to=%s", dynatrace.SLOPath, url.PathEscape(sloID), convertTimeStringToUnixMillisecondsString(testSLIStart), convertTimeStringToUnixMillisecondsString(testSLIEnd))
}

// buildUSQLRequest builds a USQL request string with the specified query for use in testing.
func buildUSQLRequest(query string) string {
	return fmt.Sprintf("%s?addDeepLinkFields=false&endTimestamp=%s&explain=false&query=%s&startTimestamp=%s", dynatrace.USQLPath, convertTimeStringToUnixMillisecondsString(testSLIEnd), url.QueryEscape(query), convertTimeStringToUnixMillisecondsString(testSLIStart))
}

func runGetSLIsFromDashboardTestWithDashboardParameterAndCheckSLIs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, dashboard string, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	runGetSLIsFromDashboardTestWithConfigClientAndDashboardParameterAndCheckSLIs(t, handler, &uploadSLOsConfigClientMock{t: t}, getSLIEventData, dashboard, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
}

func runGetSLIsFromDashboardTestWithDashboardParameterAndCheckSLIsAndSLOs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, dashboard string, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData), uploadedSLOsAssertionsFunc func(t *testing.T, actual *keptnapi.ServiceLevelObjectives), sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	configClient := &uploadSLOsConfigClientMock{t: t}
	runGetSLIsFromDashboardTestWithConfigClientAndDashboardParameterAndCheckSLIs(t, handler, configClient, getSLIEventData, dashboard, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
	uploadedSLOsAssertionsFunc(t, configClient.uploadedSLOs)
}

func runGetSLIsFromDashboardTestWithConfigClientAndDashboardParameterAndCheckSLIs(t *testing.T, handler http.Handler, configClient configClientInterface, getSLIEventData *getSLIEventData, dashboard string, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	eventSenderClient := &eventSenderClientMock{}
	runTestAndAssertNoError(t, getSLIEventData, handler, eventSenderClient, configClient, dashboard)
	assertCorrectGetSLIEvents(t, eventSenderClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
}

func runGetSLIsFromDashboardTestWithConfigClientAndCheckSLIs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, configClient configClientInterface, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	runGetSLIsFromDashboardTestWithConfigClientAndDashboardParameterAndCheckSLIs(t, handler, configClient, getSLIEventData, testDashboardID, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
}

func runGetSLIsFromDashboardTestAndCheckSLIs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData), sliResultsAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	runGetSLIsFromDashboardTestWithConfigClientAndDashboardParameterAndCheckSLIs(t, handler, &uploadSLOsConfigClientMock{t: t}, getSLIEventData, testDashboardID, getSLIFinishedEventAssertionsFunc, sliResultsAssertionsFuncs...)
}

func runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t *testing.T, handler http.Handler, getSLIEventData *getSLIEventData, getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData), uploadedSLOsAssertionsFunc func(t *testing.T, actual *keptnapi.ServiceLevelObjectives), sliResultsAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	configClient := &uploadSLOsConfigClientMock{t: t}
	runGetSLIsFromDashboardTestWithConfigClientAndDashboardParameterAndCheckSLIs(t, handler, configClient, getSLIEventData, testDashboardID, getSLIFinishedEventAssertionsFunc, sliResultsAssertionsFuncs...)
	uploadedSLOsAssertionsFunc(t, configClient.uploadedSLOs)
}

func runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t *testing.T, handler http.Handler, configClient *getSLIsConfigClientMock, requestedIndicator string, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFunc func(t *testing.T, actual sliResult)) {
	runGetSLIsFromFilesTestAndCheckSLIs(t, handler, configClient, []string{requestedIndicator}, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
}

func runGetSLIsFromFilesTestWithNoIndicatorsRequestedAndCheckSLIs(t *testing.T, handler http.Handler, configClient *getSLIsConfigClientMock, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFunc func(t *testing.T, actual sliResult)) {
	runGetSLIsFromFilesTestAndCheckSLIs(t, handler, configClient, []string{}, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
}

func runGetSLIsFromFilesTestAndCheckSLIs(t *testing.T, handler http.Handler, configClient *getSLIsConfigClientMock, requestedIndicators []string, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFunc func(t *testing.T, actual sliResult)) {
	runGetSLIsFromFilesTestWithEventAndCheckSLIs(t, handler, configClient, createTestGetSLIEventDataWithIndicators(requestedIndicators), getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
}

func runGetSLIsFromFilesTestWithEventAndCheckSLIs(t *testing.T, handler http.Handler, configClient *getSLIsConfigClientMock, ev *getSLIEventData, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFunc func(t *testing.T, actual sliResult)) {
	eventSenderClient := &eventSenderClientMock{}

	// we do not want to query a dashboard, so we leave it empty
	runTestAndAssertNoError(t, ev, handler, eventSenderClient, configClient, "")

	assertCorrectGetSLIEvents(t, eventSenderClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
}

func runTestAndAssertNoError(t *testing.T, ev *getSLIEventData, handler http.Handler, eventSenderClient *eventSenderClientMock, configClient configClientInterface, dashboard string) {
	eh, _, teardown := createGetSLIEventHandler(t, ev, handler, eventSenderClient, configClient, dashboard)
	defer teardown()

	assert.NoError(t, eh.HandleEvent(context.Background(), context.Background()))
}

func assertCorrectGetSLIEvents(t *testing.T, events []*cloudevents.Event, getSLIFinishedEventAssertionsFunc func(*testing.T, *getSLIFinishedEventData), sliResultAssertionsFuncs ...func(*testing.T, sliResult)) {
	assert.EqualValues(t, 2, len(events))

	assert.EqualValues(t, keptnv2.GetStartedEventType(keptnv2.GetSLITaskName), events[0].Type())
	assert.EqualValues(t, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), events[1].Type())

	var data getSLIFinishedEventData
	err := json.Unmarshal(events[1].Data(), &data)
	if err != nil {
		t.Fatalf("could not parse event payload correctly: %s", err)
	}

	getSLIFinishedEventAssertionsFunc(t, &data)

	assert.EqualValues(t, keptnv2.StatusSucceeded, data.Status)

	assertCorrectSLIResults(t, &data, sliResultAssertionsFuncs...)
}

func assertCorrectSLIResults(t *testing.T, getSLIFinishedEventData *getSLIFinishedEventData, sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	if !assert.EqualValues(t, len(sliResultAssertionsFuncs), len(getSLIFinishedEventData.GetSLI.IndicatorValues), "number of assertions should match number of SLI indicator values") {
		return
	}
	for i, assertionsFunction := range sliResultAssertionsFuncs {
		assertionsFunction(t, getSLIFinishedEventData.GetSLI.IndicatorValues[i])
	}
}

func createSLIAssertionsFunc(expectedMetric string, expectedDefinition string) func(t *testing.T, actualMetric string, actualDefinition string) {
	return func(t *testing.T, actualMetric string, actualDefinition string) {
		assert.EqualValues(t, expectedMetric, actualMetric)
		assert.EqualValues(t, expectedDefinition, actualDefinition)
	}
}

func createSuccessfulSLIResultAssertionsFunc(expectedMetric string, expectedValue float64, expectedQuery string) func(t *testing.T, actual sliResult) {
	return func(t *testing.T, actual sliResult) {
		assert.EqualValues(t, expectedMetric, actual.Metric, "Indicator metric should match")
		assert.EqualValues(t, expectedValue, actual.Value, "Indicator values should match")
		assert.EqualValues(t, expectedQuery, actual.Query, "Indicator query should match")
		assert.True(t, actual.Success, "Indicator success should be true")
	}
}

func createFailedSLIResultAssertionsFunc(expectedMetric string, expectedMessageSubstrings ...string) func(*testing.T, sliResult) {
	return func(t *testing.T, actual sliResult) {
		assert.False(t, actual.Success, "Indicator success should be false")
		assert.EqualValues(t, expectedMetric, actual.Metric, "Indicator metric should match")
		assert.Zero(t, actual.Value, "Indicator value should be zero")
		assert.Empty(t, actual.Query, "Indicator query should be empty")

		for _, expectedSubstring := range expectedMessageSubstrings {
			assert.Contains(t, actual.Message, expectedSubstring, "all substrings should be contained in message")
		}
	}
}

func createFailedSLIResultWithQueryAssertionsFunc(expectedMetric string, expectedQuery string, expectedMessageSubstrings ...string) func(*testing.T, sliResult) {
	return func(t *testing.T, actual sliResult) {
		assert.False(t, actual.Success, "Indicator success should be false")
		assert.EqualValues(t, expectedMetric, actual.Metric, "Indicator metric should match")
		assert.Zero(t, actual.Value, "Indicator value should be zero")
		assert.EqualValues(t, expectedQuery, actual.Query, "Indicator query should match")

		for _, expectedSubstring := range expectedMessageSubstrings {
			assert.Contains(t, actual.Message, expectedSubstring, "all substrings should be contained in message")
		}
	}
}

func createGetSLIEventHandler(t *testing.T, keptnEvent GetSLITriggeredAdapterInterface, handler http.Handler, eventSenderClient keptn.EventSenderClientInterface, configClient configClientInterface, dashboard string) (*GetSLIEventHandler, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials, err := credentials.NewDynatraceCredentials(url, testDynatraceAPIToken)
	assert.NoError(t, err)

	eh := &GetSLIEventHandler{
		event:             keptnEvent,
		dtClient:          dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		eventSenderClient: eventSenderClient,
		configClient:      configClient,
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

type getSLIsConfigClientMock struct {
	t            *testing.T
	slis         map[string]string
	getSLIsError error
}

func newConfigClientMockWithNoSLIsOrError(t *testing.T) *getSLIsConfigClientMock {
	return &getSLIsConfigClientMock{
		t: t,
	}
}

func newConfigClientMockWithSLIs(t *testing.T, slis map[string]string) *getSLIsConfigClientMock {
	return &getSLIsConfigClientMock{
		t:    t,
		slis: slis,
	}
}

func newConfigClientMockThatErrorsGetSLIs(t *testing.T, getSLIsError error) *getSLIsConfigClientMock {
	return &getSLIsConfigClientMock{
		t:            t,
		getSLIsError: getSLIsError,
	}
}

func (m *getSLIsConfigClientMock) GetSLIs(_ context.Context, _ string, _ string, _ string) (map[string]string, error) {
	if m.getSLIsError != nil {
		return nil, m.getSLIsError
	}

	return m.slis, nil
}

func (m *getSLIsConfigClientMock) GetSLOs(_ context.Context, _ string, _ string, _ string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")
	return nil, nil
}

func (m *getSLIsConfigClientMock) UploadSLOs(_ context.Context, _ string, _ string, _ string, _ *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLOs() should not be needed in this mock!")
	return nil
}

type eventSenderClientMock struct {
	eventSink []*cloudevents.Event
}

func (m *eventSenderClientMock) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
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

// uploadSLOsConfigClientMock is a mock implementation of configClientInterface which provides a mock implementation of UploadSLOs which optionally returns an error.
type uploadSLOsConfigClientMock struct {
	t               *testing.T
	uploadSLOsError error
	slosUploaded    bool
	uploadedSLOs    *keptnapi.ServiceLevelObjectives
}

func newConfigClientMockThatAllowsUploadSLOs(t *testing.T) *uploadSLOsConfigClientMock {
	return &uploadSLOsConfigClientMock{t: t}
}

func newConfigClientMockThatErrorsUploadSLOs(t *testing.T, uploadSLOsError error) *uploadSLOsConfigClientMock {
	return &uploadSLOsConfigClientMock{
		t:               t,
		uploadSLOsError: uploadSLOsError,
	}
}

func (m *uploadSLOsConfigClientMock) GetSLIs(_ context.Context, _ string, _ string, _ string) (map[string]string, error) {
	m.t.Fatalf("GetSLIs() should not be needed in this mock!")
	return nil, nil
}

func (m *uploadSLOsConfigClientMock) GetSLOs(_ context.Context, _ string, _ string, _ string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")
	return nil, nil
}

func (m *uploadSLOsConfigClientMock) UploadSLOs(_ context.Context, _ string, _ string, _ string, slos *keptnapi.ServiceLevelObjectives) error {
	if m.uploadSLOsError != nil {
		return m.uploadSLOsError
	}

	m.uploadedSLOs = slos
	m.slosUploaded = true
	return nil
}

type metricsV2QueryRequestBuilder struct {
	values url.Values
}

func newMetricsV2QueryRequestBuilder(metricSelector string) *metricsV2QueryRequestBuilder {
	values := url.Values{}
	values.Add("metricSelector", metricSelector)
	values.Add("from", convertTimeStringToUnixMillisecondsString(testSLIStart))
	values.Add("to", convertTimeStringToUnixMillisecondsString(testSLIEnd))
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) withEntitySelector(entitySelector string) *metricsV2QueryRequestBuilder {
	values := cloneURLValues(b.values)
	values.Add("entitySelector", entitySelector)
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) withFold() *metricsV2QueryRequestBuilder {
	existingMetricSelector := b.metricSelector()
	values := cloneURLValues(b.values)
	values.Set("metricSelector", "("+existingMetricSelector+"):fold()")
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) withResolution(resolution string) *metricsV2QueryRequestBuilder {
	values := cloneURLValues(b.values)
	values.Add("resolution", resolution)
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) withMZSelector(mzSelector string) *metricsV2QueryRequestBuilder {
	values := cloneURLValues(b.values)
	values.Add("mzSelector", mzSelector)
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) encode() string {
	return fmt.Sprintf("%s?%s", dynatrace.MetricsQueryPath, b.values.Encode())
}

func (b *metricsV2QueryRequestBuilder) metricSelector() string {
	return b.values.Get("metricSelector")
}

func cloneURLValues(values url.Values) url.Values {
	clone := url.Values{}
	for k, v := range values {
		for _, vv := range v {
			clone.Add(k, vv)
		}
	}
	return clone
}

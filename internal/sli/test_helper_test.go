package sli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn/go-utils/pkg/common/timeutils"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const testProject = "sockshop"
const testStage = "staging"
const testService = "carts"

const testIndicatorResponseTimeP95 = "response_time_p95"
const testIndicatorStaticSLOPass = "static_slo_-_pass"
const testIndicatorNoMetric = "no_metric"
const testDynatraceAPIToken = "dtOc01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"
const testDashboardID = "12345678-1111-4444-8888-123456789012"
const testDashboardQuery = "query"
const testSLIStart = "2022-09-28T00:00:00.000Z"
const testSLIEnd = "2022-09-29T00:00:00.000Z"

const resolutionInf = "Inf"
const resolutionIsNullKeyValuePair = "resolution=null&"
const singleValueVisualConfigType = "SINGLE_VALUE"
const graphChartVisualConfigType = "GRAPH_CHART"

const (
	testErrorSubStringZeroMetricSeriesCollections = "Metrics API v2 returned zero metric series collections"
	testErrorSubStringZeroMetricSeries            = "Metrics API v2 returned zero metric series"
	testErrorSubStringZeroValues                  = "Metrics API v2 returned zero values"
	testErrorSubStringNullAsValue                 = "Metrics API v2 returned 'null' as value"
	testErrorSubStringTwoMetricSeriesCollections  = "Metrics API v2 returned 2 metric series collections"
	testErrorSubStringTwoMetricSeries             = "Metrics API v2 returned 2 metric series"
	testErrorSubStringTwoValues                   = "Metrics API v2 returned 2 values"
)

var testSLOsWithResponseTimeP95 = createTestSLOs(createTestSLOWithPassCriterion(testIndicatorResponseTimeP95, "<=200"))

var testGetSLIEventData = createTestGetSLIEventDataWithIndicators([]string{testIndicatorResponseTimeP95})

var getSLIFinishedEventSuccessAssertionsFunc = func(t *testing.T, data *getSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultPass, data.Result)
	assert.Empty(t, data.Message)
}

func createGetSLIFinishedEventSuccessAssertionsFuncWithMessageSubstrings(expectedMessageSubstrings ...string) func(*testing.T, *getSLIFinishedEventData) {
	return func(t *testing.T, data *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultPass, data.Result)
		assertMessageContainsSubstrings(t, data.Message, expectedMessageSubstrings...)
	}
}

var getSLIFinishedEventWarningAssertionsFunc = func(t *testing.T, data *getSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultWarning, data.Result)
	assert.NotEmpty(t, data.Message)
}

func createGetSLIFinishedEventFailureAssertionsFuncWithMessageSubstrings(expectedMessageSubstrings ...string) func(*testing.T, *getSLIFinishedEventData) {
	return func(t *testing.T, data *getSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
		assertMessageContainsSubstrings(t, data.Message, expectedMessageSubstrings...)
	}
}

var getSLIFinishedEventFailureAssertionsFunc = func(t *testing.T, data *getSLIFinishedEventData) {
	assert.EqualValues(t, keptnv2.ResultFailed, data.Result)
	assert.NotEmpty(t, data.Message)
}

func createTestGetSLIEventDataWithIndicators(indicators []string) *getSLIEventData {
	return &getSLIEventData{
		project:    testProject,
		stage:      testStage,
		service:    testService,
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

func assertMessageContainsSubstrings(t *testing.T, message string, expectedSubstrings ...string) bool {
	for _, expectedSubstring := range expectedSubstrings {
		if !assert.Contains(t, message, expectedSubstring, "all substrings should be contained in message") {
			return false
		}
	}
	return true
}

// buildMetricsV2DefinitionRequestString builds a Metrics v2 definition request string with the specified metric ID for use in testing.
func buildMetricsV2DefinitionRequestString(metricID string) string {
	return fmt.Sprintf("%s/%s", dynatrace.MetricsPath, url.PathEscape(metricID))
}

// buildMetricsUnitsConvertRequest builds a Metrics Units convert request string with the specified source unit ID, value and target unit ID for use in testing.
func buildMetricsUnitsConvertRequest(sourceUnitID string, value float64, targetUnitID string) string {
	vs := strconv.FormatFloat(value, 'f', -1, 64)
	return fmt.Sprintf("%s/%s/convert?targetUnit=%s&value=%s", dynatrace.MetricsUnitsPath, url.PathEscape(sourceUnitID), targetUnitID, vs)
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

func runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t *testing.T, handler http.Handler, configClient configClientInterface, requestedIndicator string, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFunc func(t *testing.T, actual sliResult)) {
	runGetSLIsFromFilesTestAndCheckSLIs(t, handler, configClient, []string{requestedIndicator}, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
}

func runGetSLIsFromFilesTestWithNoIndicatorsRequestedAndCheckSLIs(t *testing.T, handler http.Handler, configClient configClientInterface, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	runGetSLIsFromFilesTestAndCheckSLIs(t, handler, configClient, []string{}, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
}

func runGetSLIsFromFilesTestAndCheckSLIs(t *testing.T, handler http.Handler, configClient configClientInterface, requestedIndicators []string, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	runGetSLIsFromFilesTestWithEventAndCheckSLIs(t, handler, configClient, createTestGetSLIEventDataWithIndicators(requestedIndicators), getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
}

func runGetSLIsFromFilesTestWithEventAndCheckSLIs(t *testing.T, handler http.Handler, configClient configClientInterface, ev *getSLIEventData, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData), sliResultAssertionsFuncs ...func(t *testing.T, actual sliResult)) {
	eventSenderClient := &eventSenderClientMock{}

	// we do not want to query a dashboard, so we leave it empty
	runTestAndAssertNoError(t, ev, handler, eventSenderClient, configClient, "")

	assertCorrectGetSLIEvents(t, eventSenderClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFuncs...)
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
		assert.Empty(t, actual.Message)
		assert.True(t, actual.Success, "Indicator success should be true")
	}
}

func createFailedSLIResultAssertionsFunc(expectedMetric string, expectedMessageSubstrings ...string) func(*testing.T, sliResult) {
	return func(t *testing.T, actual sliResult) {
		assert.False(t, actual.Success, "Indicator success should be false")
		assert.EqualValues(t, expectedMetric, actual.Metric, "Indicator metric should match")
		assert.Zero(t, actual.Value, "Indicator value should be zero")
		assert.Empty(t, actual.Query, "Indicator query should be empty")
		assertMessageContainsSubstrings(t, actual.Message, expectedMessageSubstrings...)
	}
}

func createFailedSLIResultWithQueryAssertionsFunc(expectedMetric string, expectedQuery string, expectedMessageSubstrings ...string) func(*testing.T, sliResult) {
	return func(t *testing.T, actual sliResult) {
		assert.False(t, actual.Success, "Indicator success should be false")
		assert.EqualValues(t, expectedMetric, actual.Metric, "Indicator metric should match")
		assert.Zero(t, actual.Value, "Indicator value should be zero")
		assert.EqualValues(t, expectedQuery, actual.Query, "Indicator query should match")
		assertMessageContainsSubstrings(t, actual.Message, expectedMessageSubstrings...)
	}
}

func createGetSLIEventHandler(t *testing.T, keptnEvent GetSLITriggeredAdapterInterface, handler http.Handler, eventSenderClient keptn.EventSenderClientInterface, configClient configClientInterface, dashboardProperty string) (*GetSLIEventHandler, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials, err := credentials.NewDynatraceCredentials(url, testDynatraceAPIToken)
	assert.NoError(t, err)

	eh := &GetSLIEventHandler{
		event:             keptnEvent,
		dtClient:          dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		eventSenderClient: eventSenderClient,
		configClient:      configClient,
		dashboardProperty: dashboardProperty,
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

type getSLIsAndGetSLOsConfigClientMock struct {
	t            *testing.T
	slis         map[string]string
	getSLIsError error
	slos         *keptnapi.ServiceLevelObjectives
	getSLOsError error
}

func newConfigClientMockWithSLIsAndSLOs(t *testing.T, slis map[string]string, slos *keptnapi.ServiceLevelObjectives) *getSLIsAndGetSLOsConfigClientMock {
	return &getSLIsAndGetSLOsConfigClientMock{
		t:    t,
		slis: slis,
		slos: slos,
	}
}

func newConfigClientMockThatErrorsGetSLIs(t *testing.T, getSLIsError error) *getSLIsAndGetSLOsConfigClientMock {
	return &getSLIsAndGetSLOsConfigClientMock{
		t:            t,
		getSLIsError: getSLIsError,
	}
}

func newConfigClientMockWithSLIsThatErrorsGetSLOs(t *testing.T, slis map[string]string, getSLOsError error) *getSLIsAndGetSLOsConfigClientMock {
	return &getSLIsAndGetSLOsConfigClientMock{
		t:            t,
		slis:         slis,
		getSLOsError: getSLOsError,
	}
}

func (m *getSLIsAndGetSLOsConfigClientMock) GetSLIs(_ context.Context, _ string, _ string, _ string) (map[string]string, error) {
	if m.getSLIsError != nil {
		return nil, m.getSLIsError
	}

	return m.slis, nil
}

func (m *getSLIsAndGetSLOsConfigClientMock) GetSLOs(_ context.Context, _ string, _ string, _ string) (*keptnapi.ServiceLevelObjectives, error) {
	if m.getSLOsError != nil {
		return nil, m.getSLOsError
	}
	return m.slos, nil
}

func (m *getSLIsAndGetSLOsConfigClientMock) UploadSLOs(_ context.Context, _ string, _ string, _ string, _ *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLOs() should not be needed in this mock!")
	return nil
}

func createTestSLOs(objectives ...*keptncommon.SLO) *keptncommon.ServiceLevelObjectives {
	totalScore := common.CreateDefaultSLOScore()
	comparison := common.CreateDefaultSLOComparison()
	return &keptncommon.ServiceLevelObjectives{
		Objectives: objectives,
		TotalScore: &totalScore,
		Comparison: &comparison,
	}
}

func createTestSLOWithPassCriterion(name string, passCriterion string) *keptncommon.SLO {
	return &keptncommon.SLO{
		SLI:    name,
		Pass:   []*keptncommon.SLOCriteria{{Criteria: []string{passCriterion}}},
		Weight: 1,
	}
}

func createTestInformationalSLO(name string) *keptncommon.SLO {
	return &keptncommon.SLO{
		SLI:    name,
		Weight: 1,
	}
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

func (b *metricsV2QueryRequestBuilder) copyWithEntitySelector(entitySelector string) *metricsV2QueryRequestBuilder {
	values := cloneURLValues(b.values)
	values.Add("entitySelector", entitySelector)
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) copyWithFold() *metricsV2QueryRequestBuilder {
	existingMetricSelector := b.metricSelector()
	values := cloneURLValues(b.values)
	values.Set("metricSelector", "("+existingMetricSelector+"):fold()")
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) copyWithResolution(resolution string) *metricsV2QueryRequestBuilder {
	values := cloneURLValues(b.values)
	values.Add("resolution", resolution)
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) copyWithMZSelector(mzSelector string) *metricsV2QueryRequestBuilder {
	values := cloneURLValues(b.values)
	values.Add("mzSelector", mzSelector)
	return &metricsV2QueryRequestBuilder{values: values}
}

func (b *metricsV2QueryRequestBuilder) build() string {
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

func createHandlerWithDashboard(t *testing.T, testDataFolder string) *test.CombinedURLHandler {
	handler := test.NewCombinedURLHandler(t)
	handler.AddExactFile(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	return handler
}

func createHandlerWithTemplatedDashboard(t *testing.T, templateFilename string, templatingData interface{}) *test.CombinedURLHandler {
	handler := test.NewCombinedURLHandler(t)
	handler.AddExactTemplate(dynatrace.DashboardsPath+"/"+testDashboardID, templateFilename, templatingData)
	return handler
}

func addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler *test.CombinedURLHandler, testDataFolder string, requestBuilder *metricsV2QueryRequestBuilder) string {
	expectedMetricsRequest1 := requestBuilder.build()
	expectedMetricsRequest2 := requestBuilder.copyWithResolution(resolutionInf).build()

	handler.AddExactFile(buildMetricsV2DefinitionRequestString(requestBuilder.metricSelector()), filepath.Join(testDataFolder, "metrics_get_by_id.json"))
	handler.AddExactFile(expectedMetricsRequest1, filepath.Join(testDataFolder, "metrics_get_by_query1.json"))
	handler.AddExactFile(expectedMetricsRequest2, filepath.Join(testDataFolder, "metrics_get_by_query2.json"))

	return expectedMetricsRequest2
}

func addRequestsToHandlerForSuccessfulMetricsQueryWithFold(handler *test.CombinedURLHandler, testDataFolder string, requestBuilder *metricsV2QueryRequestBuilder) string {
	expectedMetricsRequest1 := requestBuilder.build()
	expectedMetricsRequest2 := requestBuilder.copyWithFold().build()

	handler.AddExactFile(buildMetricsV2DefinitionRequestString(requestBuilder.metricSelector()), filepath.Join(testDataFolder, "metrics_get_by_id.json"))
	handler.AddExactFile(expectedMetricsRequest1, filepath.Join(testDataFolder, "metrics_get_by_query1.json"))
	handler.AddExactFile(expectedMetricsRequest2, filepath.Join(testDataFolder, "metrics_get_by_query2.json"))

	return expectedMetricsRequest2
}

func addRequestToHandlerForBaseMetricDefinition(handler *test.CombinedURLHandler, testDataFolder string, baseMetricSelector string) {
	handler.AddExactFile(buildMetricsV2DefinitionRequestString(baseMetricSelector), filepath.Join(testDataFolder, "metrics_get_by_id_base.json"))
}

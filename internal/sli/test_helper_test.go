package sli

import (
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

func setupTestAndAssertNoError(t *testing.T, handler http.Handler, kClient *keptnClientMock, rClient keptn.ResourceClientInterface, dashboard string) {
	ev := &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
	}

	eh, _, teardown := createGetSLIEventHandler(ev, handler, kClient, rClient, dashboard)
	defer teardown()

	err := eh.retrieveMetrics()

	assert.NoError(t, err)
}

func assertThatEventHasExpectedPayloadWithMatchingFunc(t *testing.T, assertionsFunc func(*testing.T, *keptnv2.SLIResult), events []*cloudevents.Event, eventAssertionsFunc func(data *keptnv2.GetSLIFinishedEventData)) {
	data := assertThatEventsAreThere(t, events, eventAssertionsFunc)

	assert.EqualValues(t, 1, len(data.GetSLI.IndicatorValues))
	assertionsFunc(t, data.GetSLI.IndicatorValues[0])
}

func assertThatEventsAreThere(t *testing.T, events []*cloudevents.Event, eventAssertionsFunc func(data *keptnv2.GetSLIFinishedEventData)) *keptnv2.GetSLIFinishedEventData {
	assert.EqualValues(t, 2, len(events))

	assert.EqualValues(t, keptnv2.GetStartedEventType(keptnv2.GetSLITaskName), events[0].Type())
	assert.EqualValues(t, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), events[1].Type())

	var data keptnv2.GetSLIFinishedEventData
	err := json.Unmarshal(events[1].Data(), &data)
	if err != nil {
		t.Fatalf("could not parse event payload correctly: %s", err)
	}

	eventAssertionsFunc(&data)

	assert.EqualValues(t, keptnv2.StatusSucceeded, data.Status)

	return &data
}

func createGetSLIEventHandler(keptnEvent GetSLITriggeredAdapterInterface, handler http.Handler, kClient keptn.ClientInterface, rClient keptn.ResourceClientInterface, dashboard string) (*GetSLIEventHandler, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	eh := &GetSLIEventHandler{
		event:          keptnEvent,
		dtClient:       dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		kClient:        kClient,
		resourceClient: rClient,
		dashboard:      dashboard,
		secretName:     "dynatrace", // we do not need this string
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
	t *testing.T
}

func (m *resourceClientMock) GetSLOs(project string, stage string, service string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")
	return nil, nil
}

func (m *resourceClientMock) UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error {
	m.t.Fatalf("UploadSLI() should not be needed in this mock!")
	return nil
}

func (m *resourceClientMock) UploadSLOs(project string, stage string, service string, dashboardSLOs *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLOs() should not be needed in this mock!")
	return nil
}

func (m *resourceClientMock) GetDashboard(project string, stage string, service string) (string, error) {
	// we do not want to have any dashboard stored, so return empty string
	return "", nil
}

func (m *resourceClientMock) UploadDashboard(project string, stage string, service string, dashboard *dynatrace.Dashboard) error {
	m.t.Fatalf("UploadDashboard() should not be needed in this mock!")
	return nil
}

type keptnClientMock struct {
	eventSink          []*cloudevents.Event
	customQueries      map[string]string
	customQueriesError error
}

func (m *keptnClientMock) GetCustomQueries(project string, stage string, service string) (*keptn.CustomQueries, error) {
	if m.customQueriesError != nil {
		return nil, m.customQueriesError
	}

	if m.customQueries == nil {
		return keptn.NewEmptyCustomQueries(), nil
	}

	return keptn.NewCustomQueries(m.customQueries), nil
}

func (m *keptnClientMock) GetShipyard() (*keptnv2.Shipyard, error) {
	panic("GetShipyard() should not be needed in this mock!")
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

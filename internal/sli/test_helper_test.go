package sli

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/http"
	"sync"
)

func createGetSLIEventHandler(keptnEvent GetSLITriggeredAdapterInterface, handler http.Handler, kClient keptn.ClientInterface) (*GetSLIEventHandler, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	eh := &GetSLIEventHandler{
		event:          keptnEvent,
		dtClient:       dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		kClient:        kClient,
		resourceClient: &resourceClientMock{},
		dashboard:      "",          // we do not want to query a dashboard, so we leave it empty (and have no dashboard stored)
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

type resourceClientMock struct{}

func (m *resourceClientMock) GetSLOs(project string, stage string, service string) (*keptnapi.ServiceLevelObjectives, error) {
	panic("GetSLOs() should not be needed in this mock!")
}

func (m *resourceClientMock) UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error {
	panic("UploadSLI() should not be needed in this mock!")
}

func (m *resourceClientMock) UploadSLOs(project string, stage string, service string, dashboardSLOs *keptnapi.ServiceLevelObjectives) error {
	panic("UploadSLOs() should not be needed in this mock!")
}

func (m *resourceClientMock) GetDashboard(project string, stage string, service string) (string, error) {
	// we do not want to have any dashboard stored, so return empty string
	return "", nil
}

func (m *resourceClientMock) UploadDashboard(project string, stage string, service string, dashboard *dynatrace.Dashboard) error {
	panic("UploadDashboard() should not be needed in this mock!")
}

type keptnClientMock struct {
	eventSink          []*cloudevents.Event
	customQueries      map[string]string
	customQueriesError error
	mutex              sync.Mutex
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
	m.mutex.Lock()
	defer m.mutex.Unlock()

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

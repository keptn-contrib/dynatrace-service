package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type GetSLITriggeredEvent struct {
	Context string
	Source  string
	Event   string

	Project            string
	Stage              string
	Service            string
	Deployment         string
	TestStrategy       string
	DeploymentStrategy string

	Labels map[string]string

	IsForDynatrace bool
	SLIStart       string
	SLIEnd         string
	Indicators     []string
	ID             string
	SLIFilters     []*keptnv2.SLIFilter
}

// GetShKeptnContext returns the shkeptncontext
func (e *GetSLITriggeredEvent) GetShKeptnContext() string {
	return e.Context
}

// GetSource returns the source specified in the CloudEvent context
func (e *GetSLITriggeredEvent) GetSource() string {
	return e.Source
}

// GetEvent returns the event type
func (e *GetSLITriggeredEvent) GetEvent() string {
	return e.Event
}

// GetProject returns the project
func (e *GetSLITriggeredEvent) GetProject() string {
	return e.Project
}

// GetStage returns the stage
func (e *GetSLITriggeredEvent) GetStage() string {
	return e.Stage
}

// GetService returns the service
func (e *GetSLITriggeredEvent) GetService() string {
	return e.Service
}

// GetDeployment returns the name of the deployment
func (e *GetSLITriggeredEvent) GetDeployment() string {
	return e.Deployment
}

// GetTestStrategy returns the used test strategy
func (e *GetSLITriggeredEvent) GetTestStrategy() string {
	return e.TestStrategy
}

// GetDeploymentStrategy returns the used deployment strategy
func (e GetSLITriggeredEvent) GetDeploymentStrategy() string {
	return e.DeploymentStrategy
}

// GetLabels returns a map of labels
func (e *GetSLITriggeredEvent) GetLabels() map[string]string {
	return e.Labels
}

func (e *GetSLITriggeredEvent) IsNotForDynatrace() bool {
	return e.IsForDynatrace
}

func (e *GetSLITriggeredEvent) GetSLIStart() string {
	return e.SLIStart
}

func (e *GetSLITriggeredEvent) GetSLIEnd() string {
	return e.SLIEnd
}

func (e *GetSLITriggeredEvent) GetIndicators() []string {
	return e.Indicators
}

func (e *GetSLITriggeredEvent) GetCustomSLIFilters() []*keptnv2.SLIFilter {
	return e.SLIFilters
}

func (e *GetSLITriggeredEvent) GetEventID() string {
	return e.ID
}

func (e *GetSLITriggeredEvent) AddLabel(key string, value string) {
	if e.Labels == nil {
		e.Labels = make(map[string]string)
	}

	e.Labels[key] = value
}

// createKeptnEvent creates a new Keptn Event for project, stage and service
func createKeptnEvent(project string, stage string, service string) GetSLITriggeredAdapterInterface {
	return &GetSLITriggeredEvent{
		Project: project,
		Stage:   stage,
		Service: service,
	}
}

func createRetrievalWithHandler(keptnEvent GetSLITriggeredAdapterInterface, handler http.Handler) (*Retrieval, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	dh := NewRetrieval(
		keptnEvent,
		dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		KeptnClientMock{},
		DashboardReaderMock{})

	return dh, url, teardown
}

func TestCreateRetrievalWithHandler(t *testing.T) {
	keptnEvent := createKeptnEvent("sockshop", "dev", "carts")
	dh, url, teardown := createRetrievalWithHandler(keptnEvent, nil)
	defer teardown()

	c := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	assert.EqualValues(t, c, dh.dtClient.Credentials())
	assert.EqualValues(t, keptnEvent, dh.KeptnEvent)
	assert.EqualValues(t, KeptnClientMock{}, dh.kClient)
	assert.EqualValues(t, DashboardReaderMock{}, dh.dashboardReader)
}

type KeptnClientMock struct{}

func (KeptnClientMock) GetCustomQueries(project string, stage string, service string) (*keptn.CustomQueries, error) {
	return keptn.NewEmptyCustomQueries(), nil
}

func (KeptnClientMock) GetShipyard() (*keptnv2.Shipyard, error) {
	return &keptnv2.Shipyard{}, nil
}

func (KeptnClientMock) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
	return nil
}

type DashboardReaderMock struct{}

func (DashboardReaderMock) GetDashboard(project string, stage string, service string) (string, error) {
	return "", nil
}

package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
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

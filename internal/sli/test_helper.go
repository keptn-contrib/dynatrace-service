package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type BaseKeptnEvent struct {
	Context string
	Source  string
	Event   string

	Project            string
	Stage              string
	Service            string
	Deployment         string
	TestStrategy       string
	DeploymentStrategy string

	Image string
	Tag   string

	Labels map[string]string

	IsForDynatrace bool
	SLIStart       string
	SLIEnd         string
	Indicators     []string
	ID             string
	SLIFilters     []*keptnv2.SLIFilter
}

// GetShKeptnContext returns the shkeptncontext
func (e *BaseKeptnEvent) GetShKeptnContext() string {
	return e.Context
}

// GetSource returns the source specified in the CloudEvent context
func (e *BaseKeptnEvent) GetSource() string {
	return e.Source
}

// GetEvent returns the event type
func (e *BaseKeptnEvent) GetEvent() string {
	return e.Event
}

// GetProject returns the project
func (e *BaseKeptnEvent) GetProject() string {
	return e.Project
}

// GetStage returns the stage
func (e *BaseKeptnEvent) GetStage() string {
	return e.Stage
}

// GetService returns the service
func (e *BaseKeptnEvent) GetService() string {
	return e.Service
}

// GetDeployment returns the name of the deployment
func (e *BaseKeptnEvent) GetDeployment() string {
	return e.Deployment
}

// GetTestStrategy returns the used test strategy
func (e *BaseKeptnEvent) GetTestStrategy() string {
	return e.TestStrategy
}

// GetDeploymentStrategy returns the used deployment strategy
func (e BaseKeptnEvent) GetDeploymentStrategy() string {
	return e.DeploymentStrategy
}

// GetImage returns the deployed image
func (e *BaseKeptnEvent) GetImage() string {
	return e.Image
}

// GetTag returns the deployed tag
func (e *BaseKeptnEvent) GetTag() string {
	return e.Tag
}

// GetLabels returns a map of labels
func (e *BaseKeptnEvent) GetLabels() map[string]string {
	return e.Labels
}

func (e *BaseKeptnEvent) IsNotForDynatrace() bool {
	return e.IsForDynatrace
}

func (e *BaseKeptnEvent) GetSLIStart() string {
	return e.SLIStart
}

func (e *BaseKeptnEvent) GetSLIEnd() string {
	return e.SLIEnd
}

func (e *BaseKeptnEvent) GetIndicators() []string {
	return e.Indicators
}

func (e *BaseKeptnEvent) GetCustomSLIFilters() []*keptnv2.SLIFilter {
	return e.SLIFilters
}

func (e *BaseKeptnEvent) GetEventID() string {
	return e.ID
}

func (e *BaseKeptnEvent) AddLabel(key string, value string) {
	if e.Labels == nil {
		e.Labels = make(map[string]string)
	}

	e.Labels[key] = value
}

type KeptnClientMock struct{}

func (KeptnClientMock) GetCustomQueries(project string, stage string, service string) (*keptn.CustomQueries, error) {
	return keptn.NewEmptyCustomQueries(), nil
}

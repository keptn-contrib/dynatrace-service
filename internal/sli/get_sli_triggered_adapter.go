package sli

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type GetSLITriggeredAdapterInterface interface {
	adapter.EventContentAdapter

	IsNotForDynatrace() bool
	GetSLIStart() string
	GetSLIEnd() string
	GetIndicators() []string
	GetCustomSLIFilters() []*keptnv2.SLIFilter
	GetEventID() string
}

// GetSLITriggeredAdapter is a content adaptor for events of type sh.keptn.event.action.started
type GetSLITriggeredAdapter struct {
	event      keptnv2.GetSLITriggeredEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewGetSLITriggeredAdapterFromEvent creates a new GetSLITriggeredAdapter from a cloudevents Event
func NewGetSLITriggeredAdapterFromEvent(e cloudevents.Event) (*GetSLITriggeredAdapter, error) {
	stData := &keptnv2.GetSLITriggeredEventData{}
	err := e.DataAs(stData)
	if err != nil {
		return nil, fmt.Errorf("could not parse action started event payload: %v", err)
	}

	return &GetSLITriggeredAdapter{
		event:      *stData,
		cloudEvent: adapter.NewCloudEventAdapter(e),
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a GetSLITriggeredAdapter) GetShKeptnContext() string {
	return a.cloudEvent.Context()
}

// GetSource returns the source specified in the CloudEvent context
func (a GetSLITriggeredAdapter) GetSource() string {
	return a.cloudEvent.Source()
}

// GetEvent returns the event type
func (a GetSLITriggeredAdapter) GetEvent() string {
	return ""
}

// GetProject returns the project
func (a GetSLITriggeredAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a GetSLITriggeredAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a GetSLITriggeredAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a GetSLITriggeredAdapter) GetDeployment() string {
	return a.event.Deployment
}

// GetTestStrategy returns the used test strategy
func (a GetSLITriggeredAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a GetSLITriggeredAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetImage returns the deployed image
func (a GetSLITriggeredAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a GetSLITriggeredAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a GetSLITriggeredAdapter) GetLabels() map[string]string {
	return a.event.Labels
}

func (a GetSLITriggeredAdapter) IsNotForDynatrace() bool {
	return a.event.GetSLI.SLIProvider != "dynatrace"
}

func (a GetSLITriggeredAdapter) GetSLIStart() string {
	return a.event.GetSLI.Start
}

func (a GetSLITriggeredAdapter) GetSLIEnd() string {
	return a.event.GetSLI.End
}

func (a GetSLITriggeredAdapter) GetIndicators() []string {
	return a.event.GetSLI.Indicators
}

func (a GetSLITriggeredAdapter) GetCustomSLIFilters() []*keptnv2.SLIFilter {
	return a.event.GetSLI.CustomFilters
}

func (a GetSLITriggeredAdapter) GetEventID() string {
	return a.cloudEvent.ID()
}

func (a *GetSLITriggeredAdapter) addLabel(name string, value string) {
	if a.event.Labels == nil {
		a.event.Labels = make(map[string]string)
	}

	a.event.Labels[name] = value
}
package action

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
)

type DeploymentFinishedAdapterInterface interface {
	adapter.EventContentAdapter

	GetTime() time.Time
}

// DeploymentFinishedAdapter godoc
type DeploymentFinishedAdapter struct {
	event      keptnv2.DeploymentFinishedEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewDeploymentFinishedAdapterFromEvent creates a new DeploymentFinishedAdapter from a cloudevents Event
func NewDeploymentFinishedAdapterFromEvent(e cloudevents.Event) (*DeploymentFinishedAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	dfData := &keptnv2.DeploymentFinishedEventData{}
	err := ceAdapter.PayloadAs(dfData)
	if err != nil {
		return nil, err
	}

	return &DeploymentFinishedAdapter{
		event:      *dfData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a DeploymentFinishedAdapter) GetShKeptnContext() string {
	return a.cloudEvent.GetShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a DeploymentFinishedAdapter) GetSource() string {
	return a.cloudEvent.GetSource()
}

// GetEvent returns the event type
func (a DeploymentFinishedAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName)
}

// GetProject returns the project
func (a DeploymentFinishedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a DeploymentFinishedAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a DeploymentFinishedAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a DeploymentFinishedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a DeploymentFinishedAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a DeploymentFinishedAdapter) GetDeploymentStrategy() string {
	return a.event.Deployment.DeploymentStrategy
}

// GetLabels returns a map of labels
func (a DeploymentFinishedAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	if len(a.event.Deployment.DeploymentURIsLocal) > 0 {
		labels["deploymentURILocal"] = a.event.Deployment.DeploymentURIsLocal[0]
	}
	if len(a.event.Deployment.DeploymentURIsPublic) > 0 {
		labels["deploymentURIPublic"] = a.event.Deployment.DeploymentURIsPublic[0]
	}
	return labels
}

// GetTime returns the time stamp of the event
func (a DeploymentFinishedAdapter) GetTime() time.Time {
	return a.cloudEvent.GetTime()
}

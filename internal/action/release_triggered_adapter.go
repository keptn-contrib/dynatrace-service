package action

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type ReleaseTriggeredAdapterInterface interface {
	adapter.EventContentAdapter

	GetResult() keptnv2.ResultType
}

// ReleaseTriggeredAdapter is a content adaptor for events of type sh.keptn.event.release.triggered
type ReleaseTriggeredAdapter struct {
	event      keptnv2.ReleaseTriggeredEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewReleaseTriggeredAdapterFromEvent creates a new ReleaseTriggeredAdapter from a cloudevents Event
func NewReleaseTriggeredAdapterFromEvent(e cloudevents.Event) (*ReleaseTriggeredAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	rtData := &keptnv2.ReleaseTriggeredEventData{}
	err := ceAdapter.PayloadAs(rtData)
	if err != nil {
		return nil, err
	}

	return &ReleaseTriggeredAdapter{
		event:      *rtData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ReleaseTriggeredAdapter) GetShKeptnContext() string {
	return a.cloudEvent.GetShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a ReleaseTriggeredAdapter) GetSource() string {
	return a.cloudEvent.GetSource()
}

// GetEvent returns the event type
func (a ReleaseTriggeredAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName)
}

// GetProject returns the project
func (a ReleaseTriggeredAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a ReleaseTriggeredAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a ReleaseTriggeredAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a ReleaseTriggeredAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used Test strategy
func (a ReleaseTriggeredAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used Deployment strategy
func (a ReleaseTriggeredAdapter) GetDeploymentStrategy() string {
	return a.event.Deployment.DeploymentStrategy
}

// GetLabels returns a map of labels
func (a ReleaseTriggeredAdapter) GetLabels() map[string]string {
	return a.event.Labels
}

func (a ReleaseTriggeredAdapter) GetResult() keptnv2.ResultType {
	return a.event.Result
}

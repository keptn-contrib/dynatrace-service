package action

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type ActionStartedAdapterInterface interface {
	adapter.EventContentAdapter
}

// ActionStartedAdapter is a content adaptor for events of type sh.keptn.event.action.started
type ActionStartedAdapter struct {
	event      keptnv2.ActionStartedEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewActionStartedAdapterFromEvent creates a new ActionStartedAdapter from a cloudevents Event
func NewActionStartedAdapterFromEvent(e cloudevents.Event) (*ActionStartedAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	asData := &keptnv2.ActionStartedEventData{}
	err := ceAdapter.PayloadAs(asData)
	if err != nil {
		return nil, err
	}

	return &ActionStartedAdapter{
		event:      *asData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ActionStartedAdapter) GetShKeptnContext() string {
	return a.cloudEvent.GetShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a ActionStartedAdapter) GetSource() string {
	return a.cloudEvent.GetSource()
}

// GetEvent returns the event type
func (a ActionStartedAdapter) GetEvent() string {
	return keptnv2.GetStartedEventType(keptnv2.ActionTaskName)
}

// GetProject returns the project
func (a ActionStartedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a ActionStartedAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a ActionStartedAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a ActionStartedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a ActionStartedAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a ActionStartedAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetLabels returns a map of labels
func (a ActionStartedAdapter) GetLabels() map[string]string {
	return keptn.AddOptionalKeptnBridgeUrlToLabels(a.event.Labels, a.GetShKeptnContext())
}

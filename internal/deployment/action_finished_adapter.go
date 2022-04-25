package deployment

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type ActionFinishedAdapterInterface interface {
	adapter.EventContentAdapter

	GetResult() keptnv2.ResultType
	GetStatus() keptnv2.StatusType
}

// ActionFinishedAdapter is a content adaptor for events of type sh.keptn.event.action.finished
type ActionFinishedAdapter struct {
	event      keptnv2.ActionFinishedEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewActionFinishedAdapterFromEvent creates a new ActionFinishedAdapter from a cloudevents Event
func NewActionFinishedAdapterFromEvent(e cloudevents.Event) (*ActionFinishedAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	afData := &keptnv2.ActionFinishedEventData{}
	err := ceAdapter.PayloadAs(afData)
	if err != nil {
		return nil, err
	}

	return &ActionFinishedAdapter{
		event:      *afData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ActionFinishedAdapter) GetShKeptnContext() string {
	return a.cloudEvent.GetShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a ActionFinishedAdapter) GetSource() string {
	return a.cloudEvent.GetSource()
}

// GetEvent returns the event type
func (a ActionFinishedAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.ActionTaskName)
}

// GetProject returns the project
func (a ActionFinishedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a ActionFinishedAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a ActionFinishedAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a ActionFinishedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a ActionFinishedAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a ActionFinishedAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetLabels returns a map of labels
func (a ActionFinishedAdapter) GetLabels() map[string]string {
	return keptn.AddOptionalKeptnBridgeUrlToLabels(a.event.Labels, a.GetShKeptnContext())
}

func (a ActionFinishedAdapter) GetResult() keptnv2.ResultType {
	return a.event.Result
}

func (a ActionFinishedAdapter) GetStatus() keptnv2.StatusType {
	return a.event.Status
}

package problem

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type ActionTriggeredAdapterInterface interface {
	adapter.EventContentAdapter

	GetAction() string
	GetActionDescription() string
}

// ActionTriggeredAdapter encapsulates a cloud event and its parsed payload
type ActionTriggeredAdapter struct {
	event      keptnv2.ActionTriggeredEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewActionTriggeredAdapterFromEvent creates a new ActionTriggeredAdapter from a cloudevents Event
func NewActionTriggeredAdapterFromEvent(e cloudevents.Event) (*ActionTriggeredAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	atData := &keptnv2.ActionTriggeredEventData{}
	err := ceAdapter.PayloadAs(atData)
	if err != nil {
		return nil, err
	}

	return &ActionTriggeredAdapter{
		event:      *atData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ActionTriggeredAdapter) GetShKeptnContext() string {
	return a.cloudEvent.ShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a ActionTriggeredAdapter) GetSource() string {
	return a.cloudEvent.Source()
}

// GetEvent returns the event type
func (a ActionTriggeredAdapter) GetEvent() string {
	return keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName)
}

// GetProject returns the project
func (a ActionTriggeredAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a ActionTriggeredAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a ActionTriggeredAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a ActionTriggeredAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a ActionTriggeredAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a ActionTriggeredAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetLabels returns a map of labels
func (a ActionTriggeredAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := credentials.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels[common.KEPTNSBRIDGE_LABEL] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	return labels
}

func (a ActionTriggeredAdapter) GetAction() string {
	return a.event.Action.Action
}

func (a ActionTriggeredAdapter) GetActionDescription() string {
	return a.event.Action.Description
}

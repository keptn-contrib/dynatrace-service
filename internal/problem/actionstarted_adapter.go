package problem

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// ActionStartedAdapter is a content adaptor for events of type sh.keptn.event.action.started
type ActionStartedAdapter struct {
	event   keptnv2.ActionStartedEventData
	context string
	source  string
}

// NewActionStartedAdapter creates a new ActionStartedAdapter
func NewActionStartedAdapter(event keptnv2.ActionStartedEventData, shkeptncontext, source string) ActionStartedAdapter {
	return ActionStartedAdapter{event: event, context: shkeptncontext, source: source}
}

// NewActionStartedAdapterFromEvent creates a new ActionStartedAdapter from a cloudevents Event
func NewActionStartedAdapterFromEvent(e cloudevents.Event) (*ActionStartedAdapter, error) {
	asData := &keptnv2.ActionStartedEventData{}
	err := e.DataAs(asData)
	if err != nil {
		return nil, fmt.Errorf("could not parse action started event payload: %v", err)
	}

	adapter := NewActionStartedAdapter(*asData, event.GetShKeptnContext(e), e.Source())
	return &adapter, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ActionStartedAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a ActionStartedAdapter) GetSource() string {
	return a.source
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

// GetImage returns the deployed image
func (a ActionStartedAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a ActionStartedAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a ActionStartedAdapter) GetLabels() map[string]string {
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

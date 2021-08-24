package problem

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// ActionTriggeredAdapter godoc
type ActionTriggeredAdapter struct {
	event   keptnv2.ActionTriggeredEventData
	context string
	source  string
}

// NewActionTriggeredAdapter creates a new ActionTriggeredAdapter
func NewActionTriggeredAdapter(event keptnv2.ActionTriggeredEventData, shkeptncontext, source string) ActionTriggeredAdapter {
	return ActionTriggeredAdapter{event: event, context: shkeptncontext, source: source}
}

// NewActionTriggeredAdapterFromEvent creates a new ActionTriggeredAdapter from a cloudevents Event
func NewActionTriggeredAdapterFromEvent(e cloudevents.Event) (*ActionTriggeredAdapter, error) {
	atData := &keptnv2.ActionTriggeredEventData{}
	err := e.DataAs(atData)
	if err != nil {
		return nil, fmt.Errorf("could not parse action triggered event payload: %v", err)
	}

	adapter := NewActionTriggeredAdapter(*atData, event.GetShKeptnContext(e), e.Source())
	return &adapter, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ActionTriggeredAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a ActionTriggeredAdapter) GetSource() string {
	return a.source
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

// GetImage returns the deployed image
func (a ActionTriggeredAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a ActionTriggeredAdapter) GetTag() string {
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

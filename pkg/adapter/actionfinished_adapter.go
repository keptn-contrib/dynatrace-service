package adapter

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

// ActionFinishedAdapter is a content adaptor for events of type sh.keptn.event.action.finished
type ActionFinishedAdapter struct {
	event   keptn.ActionFinishedEventData
	context string
	source  string
}

// NewActionFinishedAdapter creates a new ActionFinishedAdapter
func NewActionFinishedAdapter(event keptn.ActionFinishedEventData, shkeptncontext, source string) ActionFinishedAdapter {
	return ActionFinishedAdapter{event: event, context: shkeptncontext}
}

// GetShKeptnContext returns the shkeptncontext
func (a ActionFinishedAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a ActionFinishedAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a ActionFinishedAdapter) GetEvent() string {
	return keptn.ActionFinishedEventType
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

// GetImage returns the deployed image
func (a ActionFinishedAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a ActionFinishedAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a ActionFinishedAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels[common.KEPTNSBRIDGE_LABEL] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	return labels
}

// ActionStartedAdapter is a content adaptor for events of type sh.keptn.event.action.started
type ActionStartedAdapter struct {
	event   keptn.ActionStartedEventData
	context string
	source  string
}

// NewActionStartedAdapter creates a new ActionStartedAdapter
func NewActionStartedAdapter(event keptn.ActionStartedEventData, shkeptncontext, source string) ActionStartedAdapter {
	return ActionStartedAdapter{event: event, context: shkeptncontext}
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
	return keptn.ActionStartedEventType
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
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels[common.KEPTNSBRIDGE_LABEL] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	return labels
}

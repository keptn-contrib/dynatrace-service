package problem

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// ActionFinishedAdapter is a content adaptor for events of type sh.keptn.event.action.finished
type ActionFinishedAdapter struct {
	event   keptnv2.ActionFinishedEventData
	context string
	source  string
}

// NewActionFinishedAdapter creates a new ActionFinishedAdapter
func NewActionFinishedAdapter(event keptnv2.ActionFinishedEventData, shkeptncontext, source string) ActionFinishedAdapter {
	return ActionFinishedAdapter{event: event, context: shkeptncontext, source: source}
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
	keptnBridgeURL, err := credentials.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels[common.KEPTNSBRIDGE_LABEL] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	return labels
}

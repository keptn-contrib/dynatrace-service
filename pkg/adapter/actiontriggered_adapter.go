package adapter

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ActionTriggeredAdapter struct {
	event   keptn.ActionTriggeredEventData
	context string
	source  string
}

func NewActionTriggeredAdapter(event keptn.ActionTriggeredEventData, shkeptncontext, source string) ActionTriggeredAdapter {
	return ActionTriggeredAdapter{event: event, context: shkeptncontext}
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
	return keptn.ActionTriggeredEventType
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
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	return labels
}

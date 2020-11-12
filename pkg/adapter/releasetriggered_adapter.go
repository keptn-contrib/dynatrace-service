package adapter

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// ReleaseTriggeredAdapter godoc
type ReleaseTriggeredAdapter struct {
	event   keptnv2.ReleaseTriggeredEventData
	context string
	source  string
}

// NewReleaseTriggeredAdapter godoc
func NewReleaseTriggeredAdapter(event keptnv2.ReleaseTriggeredEventData, shkeptncontext, source string) ReleaseTriggeredAdapter {
	return ReleaseTriggeredAdapter{event: event, context: shkeptncontext}
}

// GetShKeptnContext returns the shkeptncontext
func (a ReleaseTriggeredAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a ReleaseTriggeredAdapter) GetSource() string {
	return a.source
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

// GetImage returns the deployed image
func (a ReleaseTriggeredAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a ReleaseTriggeredAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a ReleaseTriggeredAdapter) GetLabels() map[string]string {
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

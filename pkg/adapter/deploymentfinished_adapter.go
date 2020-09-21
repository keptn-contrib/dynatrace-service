package adapter

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type DeploymentFinishedAdapter struct {
	event   keptn.DeploymentFinishedEventData
	context string
	source  string
}

func NewDeploymentFinishedAdapter(event keptn.DeploymentFinishedEventData, shkeptncontext, source string) DeploymentFinishedAdapter {
	return DeploymentFinishedAdapter{event: event, context: shkeptncontext}
}

// GetShKeptnContext returns the shkeptncontext
func (a DeploymentFinishedAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a DeploymentFinishedAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a DeploymentFinishedAdapter) GetEvent() string {
	return keptn.DeploymentFinishedEventType
}

// GetProject returns the project
func (a DeploymentFinishedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a DeploymentFinishedAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a DeploymentFinishedAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a DeploymentFinishedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a DeploymentFinishedAdapter) GetTestStrategy() string {
	return a.event.TestStrategy
}

// GetDeploymentStrategy returns the used deployment strategy
func (a DeploymentFinishedAdapter) GetDeploymentStrategy() string {
	return a.event.DeploymentStrategy
}

// GetImage returns the deployed image
func (a DeploymentFinishedAdapter) GetImage() string {
	return a.event.Image
}

// GetTag returns the deployed tag
func (a DeploymentFinishedAdapter) GetTag() string {
	return a.event.Tag
}

// GetLabels returns a map of labels
func (a DeploymentFinishedAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	if a.event.DeploymentURILocal != "" {
		labels["deploymentURILocal"] = a.event.DeploymentURILocal
	}
	if a.event.DeploymentURIPublic != "" {
		labels["deploymentURIPublic"] = a.event.DeploymentURIPublic
	}
	return labels
}

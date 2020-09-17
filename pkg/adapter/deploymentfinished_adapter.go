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

func (a DeploymentFinishedAdapter) GetContext() string {
	return a.context
}

func (a DeploymentFinishedAdapter) GetSource() string {
	return a.source
}

func (a DeploymentFinishedAdapter) GetEvent() string {
	return keptn.DeploymentFinishedEventType
}

func (a DeploymentFinishedAdapter) GetProject() string {
	return a.event.Project
}

func (a DeploymentFinishedAdapter) GetStage() string {
	return a.event.Stage
}

func (a DeploymentFinishedAdapter) GetService() string {
	return a.event.Service
}

func (a DeploymentFinishedAdapter) GetDeployment() string {
	return ""
}

func (a DeploymentFinishedAdapter) GetTestStrategy() string {
	return a.event.TestStrategy
}

func (a DeploymentFinishedAdapter) GetDeploymentStrategy() string {
	return a.event.DeploymentStrategy
}

func (a DeploymentFinishedAdapter) GetImage() string {
	return a.event.Image
}

func (a DeploymentFinishedAdapter) GetTag() string {
	return a.event.Tag
}

func (a DeploymentFinishedAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + a.GetContext()
	}
	if a.event.DeploymentURILocal != "" {
		labels["deploymentURILocal"] = a.event.DeploymentURILocal
	}
	if a.event.DeploymentURIPublic != "" {
		labels["deploymentURIPublic"] = a.event.DeploymentURIPublic
	}
	return labels
}

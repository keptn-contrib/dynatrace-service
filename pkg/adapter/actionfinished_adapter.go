package adapter

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ActionFinishedAdapter struct {
	event   keptn.ActionFinishedEventData
	context string
	source  string
}

func NewActionFinishedAdapter(event keptn.ActionFinishedEventData, shkeptncontext, source string) ActionFinishedAdapter {
	return ActionFinishedAdapter{event: event, context: shkeptncontext}
}

func (a ActionFinishedAdapter) GetContext() string {
	return a.context
}

func (a ActionFinishedAdapter) GetSource() string {
	return a.source
}

func (a ActionFinishedAdapter) GetEvent() string {
	return keptn.ActionFinishedEventType
}

func (a ActionFinishedAdapter) GetProject() string {
	return a.event.Project
}

func (a ActionFinishedAdapter) GetStage() string {
	return a.event.Stage
}

func (a ActionFinishedAdapter) GetService() string {
	return a.event.Service
}

func (a ActionFinishedAdapter) GetDeployment() string {
	return ""
}

func (a ActionFinishedAdapter) GetTestStrategy() string {
	return ""
}

func (a ActionFinishedAdapter) GetDeploymentStrategy() string {
	return ""
}

func (a ActionFinishedAdapter) GetImage() string {
	return ""
}

func (a ActionFinishedAdapter) GetTag() string {
	return ""
}

func (a ActionFinishedAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + a.GetContext()
	}
	return labels
}

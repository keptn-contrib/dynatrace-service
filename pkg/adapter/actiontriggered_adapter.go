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

func (a ActionTriggeredAdapter) GetContext() string {
	return a.context
}

func (a ActionTriggeredAdapter) GetSource() string {
	return a.source
}

func (a ActionTriggeredAdapter) GetEvent() string {
	return keptn.ActionTriggeredEventType
}

func (a ActionTriggeredAdapter) GetProject() string {
	return a.event.Project
}

func (a ActionTriggeredAdapter) GetStage() string {
	return a.event.Stage
}

func (a ActionTriggeredAdapter) GetService() string {
	return a.event.Service
}

func (a ActionTriggeredAdapter) GetDeployment() string {
	return ""
}

func (a ActionTriggeredAdapter) GetTestStrategy() string {
	return ""
}

func (a ActionTriggeredAdapter) GetDeploymentStrategy() string {
	return ""
}

func (a ActionTriggeredAdapter) GetImage() string {
	return ""
}

func (a ActionTriggeredAdapter) GetTag() string {
	return ""
}

func (a ActionTriggeredAdapter) GetLabels() map[string]string {
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

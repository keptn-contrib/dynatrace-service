package adapter

import (
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ProjectCreateAdapter struct {
	event   keptn.ProjectCreateEventData
	context string
	source  string
}

func NewProjectCreateAdapter(event keptn.ProjectCreateEventData, shkeptncontext, source string) ProjectCreateAdapter {
	return ProjectCreateAdapter{event: event, context: shkeptncontext}
}

func (a ProjectCreateAdapter) GetContext() string {
	return a.context
}

func (a ProjectCreateAdapter) GetSource() string {
	return a.source
}

func (a ProjectCreateAdapter) GetEvent() string {
	return keptn.InternalProjectCreateEventType
}

func (a ProjectCreateAdapter) GetProject() string {
	return a.event.Project
}

func (a ProjectCreateAdapter) GetStage() string {
	return ""
}

func (a ProjectCreateAdapter) GetService() string {
	return ""
}

func (a ProjectCreateAdapter) GetDeployment() string {
	return ""
}

func (a ProjectCreateAdapter) GetTestStrategy() string {
	return ""
}

func (a ProjectCreateAdapter) GetDeploymentStrategy() string {
	return ""
}

func (a ProjectCreateAdapter) GetImage() string {
	return ""
}

func (a ProjectCreateAdapter) GetTag() string {
	return ""
}

func (a ProjectCreateAdapter) GetLabels() map[string]string {
	return nil
}

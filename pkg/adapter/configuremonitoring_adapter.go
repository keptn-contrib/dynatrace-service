package adapter

import (
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type ConfigureMonitoringAdapter struct {
	event   keptn.ConfigureMonitoringEventData
	context string
	source  string
}

func NewConfigureMonitoringAdapter(event keptn.ConfigureMonitoringEventData, shkeptncontext, source string) ConfigureMonitoringAdapter {
	return ConfigureMonitoringAdapter{event: event, context: shkeptncontext}
}

func (a ConfigureMonitoringAdapter) GetContext() string {
	return a.context
}

func (a ConfigureMonitoringAdapter) GetSource() string {
	return a.source
}

func (a ConfigureMonitoringAdapter) GetEvent() string {
	return keptn.ConfigureMonitoringEventType
}

func (a ConfigureMonitoringAdapter) GetProject() string {
	return a.event.Project
}

func (a ConfigureMonitoringAdapter) GetStage() string {
	return ""
}

func (a ConfigureMonitoringAdapter) GetService() string {
	return a.event.Service
}

func (a ConfigureMonitoringAdapter) GetDeployment() string {
	return ""
}

func (a ConfigureMonitoringAdapter) GetTestStrategy() string {
	return ""
}

func (a ConfigureMonitoringAdapter) GetDeploymentStrategy() string {
	return ""
}

func (a ConfigureMonitoringAdapter) GetImage() string {
	return ""
}

func (a ConfigureMonitoringAdapter) GetTag() string {
	return ""
}

func (a ConfigureMonitoringAdapter) GetLabels() map[string]string {
	return nil
}

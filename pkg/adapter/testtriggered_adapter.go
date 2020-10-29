package adapter

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type TestTriggeredAdapter struct {
	event   keptnv2.TestTriggeredEventData
	context string
	source  string
}

func NewTestTriggeredAdapter(event keptnv2.TestTriggeredEventData, shkeptncontext, source string) TestTriggeredAdapter {
	return TestTriggeredAdapter{event: event, context: shkeptncontext}
}

// GetShKeptnContext returns the shkeptncontext
func (a TestTriggeredAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a TestTriggeredAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a TestTriggeredAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.TestTaskName)
}

// GetProject returns the project
func (a TestTriggeredAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a TestTriggeredAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a TestTriggeredAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a TestTriggeredAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a TestTriggeredAdapter) GetTestStrategy() string {
	return a.event.Test.TestStrategy
}

// GetDeploymentStrategy returns the used deployment strategy
func (a TestTriggeredAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetImage returns the deployed image
func (a TestTriggeredAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a TestTriggeredAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a TestTriggeredAdapter) GetLabels() map[string]string {
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

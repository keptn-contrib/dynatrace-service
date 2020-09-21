package adapter

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type TestFinishedAdapter struct {
	event   keptn.TestsFinishedEventData
	context string
	source  string
}

func NewTestFinishedAdapter(event keptn.TestsFinishedEventData, shkeptncontext, source string) TestFinishedAdapter {
	return TestFinishedAdapter{event: event, context: shkeptncontext}
}

// GetShKeptnContext returns the shkeptncontext
func (a TestFinishedAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a TestFinishedAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a TestFinishedAdapter) GetEvent() string {
	return keptn.TestsFinishedEventType
}

// GetProject returns the project
func (a TestFinishedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a TestFinishedAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a TestFinishedAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a TestFinishedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a TestFinishedAdapter) GetTestStrategy() string {
	return a.event.TestStrategy
}

// GetDeploymentStrategy returns the used deployment strategy
func (a TestFinishedAdapter) GetDeploymentStrategy() string {
	return a.event.DeploymentStrategy
}

// GetImage returns the deployed image
func (a TestFinishedAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a TestFinishedAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a TestFinishedAdapter) GetLabels() map[string]string {
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

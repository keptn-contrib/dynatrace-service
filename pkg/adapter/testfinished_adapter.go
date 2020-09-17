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

func (a TestFinishedAdapter) GetContext() string {
	return a.context
}

func (a TestFinishedAdapter) GetSource() string {
	return a.source
}

func (a TestFinishedAdapter) GetEvent() string {
	return keptn.TestsFinishedEventType
}

func (a TestFinishedAdapter) GetProject() string {
	return a.event.Project
}

func (a TestFinishedAdapter) GetStage() string {
	return a.event.Stage
}

func (a TestFinishedAdapter) GetService() string {
	return a.event.Service
}

func (a TestFinishedAdapter) GetDeployment() string {
	return ""
}

func (a TestFinishedAdapter) GetTestStrategy() string {
	return a.event.TestStrategy
}

func (a TestFinishedAdapter) GetDeploymentStrategy() string {
	return a.event.DeploymentStrategy
}

func (a TestFinishedAdapter) GetImage() string {
	return ""
}

func (a TestFinishedAdapter) GetTag() string {
	return ""
}

func (a TestFinishedAdapter) GetLabels() map[string]string {
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

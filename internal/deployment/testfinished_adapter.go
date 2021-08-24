package deployment

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// TestFinishedAdapter godoc
type TestFinishedAdapter struct {
	event   keptnv2.TestFinishedEventData
	context string
	source  string
}

// NewTestFinishedAdapter creates a new TestFinishedAdapter
func NewTestFinishedAdapter(event keptnv2.TestFinishedEventData, shkeptncontext, source string) TestFinishedAdapter {
	return TestFinishedAdapter{event: event, context: shkeptncontext, source: source}
}

// NewTestFinishedAdapterFromEvent creates a new TestFinishedAdapter from a cloudevents Event
func NewTestFinishedAdapterFromEvent(e cloudevents.Event) (*TestFinishedAdapter, error) {
	tfData := &keptnv2.TestFinishedEventData{}
	err := e.DataAs(tfData)
	if err != nil {
		return nil, fmt.Errorf("could not parse test finished event payload: %v", err)
	}

	adapter := NewTestFinishedAdapter(*tfData, event.GetShKeptnContext(e), e.Source())
	return &adapter, nil
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
	return keptnv2.GetFinishedEventType(keptnv2.TestTaskName)
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
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a TestFinishedAdapter) GetDeploymentStrategy() string {
	return ""
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
	keptnBridgeURL, err := credentials.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels[common.KEPTNSBRIDGE_LABEL] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	return labels
}

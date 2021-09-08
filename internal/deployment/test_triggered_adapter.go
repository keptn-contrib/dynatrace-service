package deployment

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type TestTriggeredAdapterInterface interface {
	adapter.EventContentAdapter
}

// TestTriggeredAdapter is a content adaptor for events of type sh.keptn.event.test.triggered
type TestTriggeredAdapter struct {
	event      keptnv2.TestTriggeredEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewTestTriggeredAdapterFromEvent creates a new TestTriggeredAdapter from a cloudevents Event
func NewTestTriggeredAdapterFromEvent(e cloudevents.Event) (*TestTriggeredAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	ttData := &keptnv2.TestTriggeredEventData{}
	err := ceAdapter.PayloadAs(ttData)
	if err != nil {
		return nil, err
	}

	return &TestTriggeredAdapter{
		event:      *ttData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a TestTriggeredAdapter) GetShKeptnContext() string {
	return a.cloudEvent.ShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a TestTriggeredAdapter) GetSource() string {
	return a.cloudEvent.Source()
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

// GetLabels returns a map of labels
func (a TestTriggeredAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := credentials.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	return labels
}

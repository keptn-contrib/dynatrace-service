package monitoring

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

// ConfigureMonitoringAdapter godoc
type ConfigureMonitoringAdapter struct {
	event   keptn.ConfigureMonitoringEventData
	context string
	source  string
}

// NewConfigureMonitoringAdapter creates a new ConfigureMonitoringAdapter
func NewConfigureMonitoringAdapter(event keptn.ConfigureMonitoringEventData, shkeptncontext, source string) ConfigureMonitoringAdapter {
	return ConfigureMonitoringAdapter{event: event, context: shkeptncontext, source: source}
}

// NewConfigureMonitoringAdapterFromEvent creates a new ConfigureMonitoringAdapter from a cloudevents Event
func NewConfigureMonitoringAdapterFromEvent(e cloudevents.Event) (*ConfigureMonitoringAdapter, error) {
	cmData := &keptn.ConfigureMonitoringEventData{}
	err := e.DataAs(cmData)
	if err != nil {
		return nil, fmt.Errorf("could not parse deployment finished event payload: %v", err)
	}

	adapter := NewConfigureMonitoringAdapter(*cmData, event.GetShKeptnContext(e), e.Source())
	return &adapter, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ConfigureMonitoringAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a ConfigureMonitoringAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a ConfigureMonitoringAdapter) GetEvent() string {
	return keptn.ConfigureMonitoringEventType
}

// GetProject returns the project
func (a ConfigureMonitoringAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a ConfigureMonitoringAdapter) GetStage() string {
	return ""
}

// GetService returns the service
func (a ConfigureMonitoringAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a ConfigureMonitoringAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a ConfigureMonitoringAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a ConfigureMonitoringAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetImage returns the deployed image
func (a ConfigureMonitoringAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a ConfigureMonitoringAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a ConfigureMonitoringAdapter) GetLabels() map[string]string {
	return nil
}

func (a ConfigureMonitoringAdapter) IsNotForDynatrace() bool {
	return a.event.Type != "dynatrace"
}

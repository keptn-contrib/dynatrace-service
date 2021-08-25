package monitoring

import (
	"encoding/base64"
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ProjectCreateAdapter godoc
type ProjectCreateAdapter struct {
	event   keptnv2.ProjectCreateFinishedEventData
	context string
	source  string
}

// NewProjectCreateAdapter godoc
func NewProjectCreateAdapter(event keptnv2.ProjectCreateFinishedEventData, shkeptncontext, source string) ProjectCreateAdapter {
	return ProjectCreateAdapter{event: event, context: shkeptncontext, source: source}
}

// NewProjectCreateAdapterFromEvent creates a new ProjectCreateAdapter from a cloudevents Event
func NewProjectCreateAdapterFromEvent(e cloudevents.Event) (*ProjectCreateAdapter, error) {
	pcData := &keptnv2.ProjectCreateFinishedEventData{}
	err := e.DataAs(pcData)
	if err != nil {
		return nil, fmt.Errorf("could not parse project create event payload: %v", err)
	}

	adapter := NewProjectCreateAdapter(*pcData, event.GetShKeptnContext(e), e.Source())
	return &adapter, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ProjectCreateAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a ProjectCreateAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a ProjectCreateAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName)
}

// GetProject returns the project
func (a ProjectCreateAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a ProjectCreateAdapter) GetStage() string {
	return ""
}

// GetService returns the service
func (a ProjectCreateAdapter) GetService() string {
	return ""
}

// GetDeployment returns the name of the deployment
func (a ProjectCreateAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a ProjectCreateAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a ProjectCreateAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetImage returns the deployed image
func (a ProjectCreateAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a ProjectCreateAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a ProjectCreateAdapter) GetLabels() map[string]string {
	return nil
}

func (a ProjectCreateAdapter) GetShipyard() (*keptnv2.Shipyard, error) {
	shipyard := &keptnv2.Shipyard{}
	decodedShipyard, err := base64.StdEncoding.DecodeString(a.event.CreatedProject.Shipyard)
	if err != nil {
		log.WithError(err).Error("Could not decode shipyard")
		return nil, errors.New("could not decode Keptn shipyard file")
	}
	err = yaml.Unmarshal(decodedShipyard, shipyard)
	if err != nil {
		return nil, errors.New("could not unmarshal Keptn shipyard file")
	}

	return shipyard, nil
}

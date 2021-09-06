package monitoring

import (
	"encoding/base64"
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ProjectCreateAdapterInterface interface {
	adapter.EventContentAdapter

	GetShipyard() (*keptnv2.Shipyard, error)
}

// ProjectCreateAdapter encapsulates a cloud event and its parsed payload
type ProjectCreateAdapter struct {
	event      keptnv2.ProjectCreateFinishedEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewProjectCreateAdapterFromEvent creates a new ProjectCreateAdapter from a cloudevents Event
func NewProjectCreateAdapterFromEvent(e cloudevents.Event) (*ProjectCreateAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	pcData := &keptnv2.ProjectCreateFinishedEventData{}
	err := ceAdapter.PayloadAs(pcData)
	if err != nil {
		return nil, err
	}

	return &ProjectCreateAdapter{
		*pcData,
		ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ProjectCreateAdapter) GetShKeptnContext() string {
	return a.cloudEvent.Context()
}

// GetSource returns the source specified in the CloudEvent context
func (a ProjectCreateAdapter) GetSource() string {
	return a.cloudEvent.Source()
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

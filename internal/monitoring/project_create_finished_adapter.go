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

type ProjectCreateFinishedAdapterInterface interface {
	adapter.EventContentAdapter

	GetShipyard() (*keptnv2.Shipyard, error)
}

// ProjectCreateFinishedAdapter encapsulates a cloud event and its parsed payload
type ProjectCreateFinishedAdapter struct {
	event      keptnv2.ProjectCreateFinishedEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewProjectCreateFinishedAdapterFromEvent creates a new ProjectCreateFinishedAdapter from a cloudevents Event
func NewProjectCreateFinishedAdapterFromEvent(e cloudevents.Event) (*ProjectCreateFinishedAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	pcData := &keptnv2.ProjectCreateFinishedEventData{}
	err := ceAdapter.PayloadAs(pcData)
	if err != nil {
		return nil, err
	}

	return &ProjectCreateFinishedAdapter{
		*pcData,
		ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ProjectCreateFinishedAdapter) GetShKeptnContext() string {
	return a.cloudEvent.GetShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a ProjectCreateFinishedAdapter) GetSource() string {
	return a.cloudEvent.GetSource()
}

// GetEvent returns the event type
func (a ProjectCreateFinishedAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName)
}

// GetProject returns the project
func (a ProjectCreateFinishedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a ProjectCreateFinishedAdapter) GetStage() string {
	return ""
}

// GetService returns the service
func (a ProjectCreateFinishedAdapter) GetService() string {
	return ""
}

// GetDeployment returns the name of the deployment
func (a ProjectCreateFinishedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a ProjectCreateFinishedAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a ProjectCreateFinishedAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetLabels returns a map of labels
func (a ProjectCreateFinishedAdapter) GetLabels() map[string]string {
	return nil
}

func (a ProjectCreateFinishedAdapter) GetShipyard() (*keptnv2.Shipyard, error) {
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

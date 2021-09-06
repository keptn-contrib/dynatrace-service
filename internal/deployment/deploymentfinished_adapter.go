package deployment

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"os"
	"strings"
)

type DeploymentFinishedAdapterInterface interface {
	adapter.EventContentAdapter
}

// DeploymentFinishedAdapter godoc
type DeploymentFinishedAdapter struct {
	event      keptnv2.DeploymentFinishedEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewDeploymentFinishedAdapterFromEvent creates a new DeploymentFinishedAdapter from a cloudevents Event
func NewDeploymentFinishedAdapterFromEvent(e cloudevents.Event) (*DeploymentFinishedAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	dfData := &keptnv2.DeploymentFinishedEventData{}
	err := ceAdapter.PayloadAs(dfData)
	if err != nil {
		return nil, err
	}

	return &DeploymentFinishedAdapter{
		event:      *dfData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a DeploymentFinishedAdapter) GetShKeptnContext() string {
	return a.cloudEvent.Context()
}

// GetSource returns the source specified in the CloudEvent context
func (a DeploymentFinishedAdapter) GetSource() string {
	return a.cloudEvent.Source()
}

// GetEvent returns the event type
func (a DeploymentFinishedAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName)
}

// GetProject returns the project
func (a DeploymentFinishedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a DeploymentFinishedAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a DeploymentFinishedAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a DeploymentFinishedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a DeploymentFinishedAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a DeploymentFinishedAdapter) GetDeploymentStrategy() string {
	return a.event.Deployment.DeploymentStrategy
}

// GetImage returns the deployed image
func (a DeploymentFinishedAdapter) GetImage() string {
	imageAndTag := a.getImageAndTag()
	if imageAndTag == "n/a" {
		return imageAndTag
	}
	split := strings.Split(imageAndTag, ":")
	return split[0]
}

func (a DeploymentFinishedAdapter) getImageAndTag() string {
	eventHandler := keptnapi.NewEventHandler(os.Getenv("DATASTORE"))

	notAvailable := "n/a"
	events, errObj := eventHandler.GetEvents(&keptnapi.EventFilter{
		Project:      a.GetProject(),
		Stage:        a.GetStage(),
		Service:      a.GetService(),
		EventType:    keptnv2.GetTriggeredEventType(keptnv2.DeploymentTaskName),
		KeptnContext: a.GetShKeptnContext(),
	})
	if errObj != nil || events == nil || len(events) == 0 {
		return notAvailable
	}

	triggeredData := &keptnv2.DeploymentTriggeredEventData{}
	err := keptnv2.Decode(events[0].Data, triggeredData)
	if err != nil {
		return notAvailable
	}
	for key, value := range triggeredData.ConfigurationChange.Values {
		if strings.HasSuffix(key, "image") {
			return value.(string)
		}
	}
	return notAvailable
}

// GetTag returns the deployed tag
func (a DeploymentFinishedAdapter) GetTag() string {
	notAvailable := "n/a"
	imageAndTag := a.getImageAndTag()
	if imageAndTag == notAvailable {
		return imageAndTag
	}
	split := strings.Split(imageAndTag, ":")
	if len(split) == 1 {
		return notAvailable
	}
	return split[1]
}

// GetLabels returns a map of labels
func (a DeploymentFinishedAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := credentials.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels[common.KEPTNSBRIDGE_LABEL] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	if len(a.event.Deployment.DeploymentURIsLocal) > 0 {
		labels["deploymentURILocal"] = a.event.Deployment.DeploymentURIsLocal[0]
	}
	if len(a.event.Deployment.DeploymentURIsPublic) > 0 {
		labels["deploymentURIPublic"] = a.event.Deployment.DeploymentURIsPublic[0]
	}
	return labels
}

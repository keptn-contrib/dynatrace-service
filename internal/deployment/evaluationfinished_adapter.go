package deployment

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"os"
)

type EvaluationFinishedAdapterInterface interface {
	adapter.EventContentAdapter

	IsPartOfRemediation() bool
	GetEvaluationScore() float64
	GetResult() keptnv2.ResultType
}

// EvaluationFinishedAdapter is a content adaptor for events of type sh.keptn.event.evaluation.finished
type EvaluationFinishedAdapter struct {
	event      keptnv2.EvaluationFinishedEventData
	cloudEvent adapter.CloudEventAdapter
}

// NewEvaluationFinishedAdapterFromEvent creates a new EvaluationFinishedAdapter from a cloudevents Event
func NewEvaluationFinishedAdapterFromEvent(e cloudevents.Event) (*EvaluationFinishedAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	efData := &keptnv2.EvaluationFinishedEventData{}
	err := ceAdapter.PayloadAs(efData)
	if err != nil {
		return nil, err
	}

	return &EvaluationFinishedAdapter{
		event:      *efData,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a EvaluationFinishedAdapter) GetShKeptnContext() string {
	return a.cloudEvent.Context()
}

// GetSource returns the source specified in the CloudEvent context
func (a EvaluationFinishedAdapter) GetSource() string {
	return a.cloudEvent.Source()
}

// GetEvent returns the event type
func (a EvaluationFinishedAdapter) GetEvent() string {
	return keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName)
}

// GetProject returns the project
func (a EvaluationFinishedAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a EvaluationFinishedAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a EvaluationFinishedAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a EvaluationFinishedAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a EvaluationFinishedAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a EvaluationFinishedAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetImage returns the deployed image
func (a EvaluationFinishedAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a EvaluationFinishedAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a EvaluationFinishedAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := credentials.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + a.GetShKeptnContext()
	}
	labels["Quality Gate Score"] = fmt.Sprintf("%.2f", a.event.Evaluation.Score)
	labels["No of evaluated SLIs"] = fmt.Sprintf("%d", len(a.event.Evaluation.IndicatorResults))
	labels["Evaluation Start"] = a.event.Evaluation.TimeStart
	labels["Evaluation End"] = a.event.Evaluation.TimeEnd
	return labels
}

// IsPartOfRemediation checks whether the evaluation.finished event is part of a remediation task sequence
func (a EvaluationFinishedAdapter) IsPartOfRemediation() bool {
	eventHandler := keptnapi.NewEventHandler(os.Getenv("DATASTORE"))

	events, errObj := eventHandler.GetEvents(&keptnapi.EventFilter{
		Project:      a.GetProject(),
		Stage:        a.GetStage(),
		Service:      a.GetService(),
		EventType:    keptnv2.GetTriggeredEventType("remediation"),
		KeptnContext: a.GetShKeptnContext(),
	})
	if errObj != nil || events == nil || len(events) == 0 {
		return false
	}
	return true
}

func (a EvaluationFinishedAdapter) GetEvaluationScore() float64 {
	return a.event.Evaluation.Score
}

func (a EvaluationFinishedAdapter) GetResult() keptnv2.ResultType {
	return a.event.Result
}

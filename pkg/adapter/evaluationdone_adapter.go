package adapter

import (
	"fmt"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type EvaluationDoneAdapter struct {
	event   keptnv2.EvaluationFinishedEventData
	context string
	source  string
}

func NewEvaluationDoneAdapter(event keptnv2.EvaluationFinishedEventData, shkeptncontext, source string) EvaluationDoneAdapter {
	return EvaluationDoneAdapter{event: event, context: shkeptncontext}
}

// GetShKeptnContext returns the shkeptncontext
func (a EvaluationDoneAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a EvaluationDoneAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a EvaluationDoneAdapter) GetEvent() string {
	return keptn.EvaluationDoneEventType
}

// GetProject returns the project
func (a EvaluationDoneAdapter) GetProject() string {
	return a.event.Project
}

// GetStage returns the stage
func (a EvaluationDoneAdapter) GetStage() string {
	return a.event.Stage
}

// GetService returns the service
func (a EvaluationDoneAdapter) GetService() string {
	return a.event.Service
}

// GetDeployment returns the name of the deployment
func (a EvaluationDoneAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a EvaluationDoneAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a EvaluationDoneAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetImage returns the deployed image
func (a EvaluationDoneAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a EvaluationDoneAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a EvaluationDoneAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
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

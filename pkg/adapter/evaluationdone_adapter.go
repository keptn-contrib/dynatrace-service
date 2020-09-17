package adapter

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type EvaluationDoneAdapter struct {
	event   keptn.EvaluationDoneEventData
	context string
	source  string
}

func NewEvaluationDoneAdapter(event keptn.EvaluationDoneEventData, shkeptncontext, source string) EvaluationDoneAdapter {
	return EvaluationDoneAdapter{event: event, context: shkeptncontext}
}

func (a EvaluationDoneAdapter) GetContext() string {
	return a.context
}

func (a EvaluationDoneAdapter) GetSource() string {
	return a.source
}

func (a EvaluationDoneAdapter) GetEvent() string {
	return keptn.EvaluationDoneEventType
}

func (a EvaluationDoneAdapter) GetProject() string {
	return a.event.Project
}

func (a EvaluationDoneAdapter) GetStage() string {
	return a.event.Stage
}

func (a EvaluationDoneAdapter) GetService() string {
	return a.event.Service
}

func (a EvaluationDoneAdapter) GetDeployment() string {
	return ""
}

func (a EvaluationDoneAdapter) GetTestStrategy() string {
	return a.event.TestStrategy
}

func (a EvaluationDoneAdapter) GetDeploymentStrategy() string {
	return a.event.DeploymentStrategy
}

func (a EvaluationDoneAdapter) GetImage() string {
	return ""
}

func (a EvaluationDoneAdapter) GetTag() string {
	return ""
}

func (a EvaluationDoneAdapter) GetLabels() map[string]string {
	labels := a.event.Labels
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if labels == nil {
		labels = make(map[string]string)
	}
	if err == nil {
		labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + a.GetContext()
	}
	labels["Quality Gate Score"] = fmt.Sprintf("%.2f", a.event.EvaluationDetails.Score)
	labels["No of evaluated SLIs"] = fmt.Sprintf("%d", len(a.event.EvaluationDetails.IndicatorResults))
	labels["Evaluation Start"] = a.event.EvaluationDetails.TimeStart
	labels["Evaluation End"] = a.event.EvaluationDetails.TimeEnd
	return labels
}

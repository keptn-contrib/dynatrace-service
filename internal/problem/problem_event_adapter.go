package problem

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"strings"
)

const remediationTaskName = "remediation"

const projectKey = "keptn_project"
const stageKey = "keptn_stage"
const serviceKey = "keptn_service"

// ProblemAdapter is a content adaptor for events of type sh.keptn.event.action.finished
type ProblemAdapter struct {
	event       DTProblemEvent
	problemData interface{}
	context     string
	source      string
}

// NewProblemAdapter creates a new ProblemAdapter
func NewProblemAdapter(event DTProblemEvent, shkeptncontext, source string, problemData interface{}) ProblemAdapter {
	return ProblemAdapter{event: event, context: shkeptncontext, source: source, problemData: problemData}
}

// NewProblemAdapterFromEvent creates a new ProblemAdapter from a cloudevents Event
func NewProblemAdapterFromEvent(e cloudevents.Event) (*ProblemAdapter, error) {
	pData := &DTProblemEvent{}
	err := e.DataAs(pData)
	if err != nil {
		return nil, fmt.Errorf("could not parse problem event payload: %v", err)
	}

	var problemEvent interface{}
	err = e.DataAs(problemEvent)
	if err != nil {
		return nil, fmt.Errorf("could not parse problem event payload: %v", err)
	}

	adapter := NewProblemAdapter(*pData, event.GetShKeptnContext(e), e.Source(), problemEvent)
	return &adapter, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ProblemAdapter) GetShKeptnContext() string {
	return a.context
}

// GetSource returns the source specified in the CloudEvent context
func (a ProblemAdapter) GetSource() string {
	return a.source
}

// GetEvent returns the event type
func (a ProblemAdapter) GetEvent() string {
	if a.event.State == "RESOLVED" {
		return keptn.ProblemEventType
	}

	return keptnv2.GetTriggeredEventType(fmt.Sprintf("%s.%s", a.GetStage(), remediationTaskName))
}

// GetProject returns the project
func (a ProblemAdapter) GetProject() string {
	return getPropertyFromTags(a.event.Tags, projectKey)
}

// GetStage returns the stage
func (a ProblemAdapter) GetStage() string {
	return getPropertyFromTags(a.event.Tags, stageKey)
}

// GetService returns the service
func (a ProblemAdapter) GetService() string {
	return getPropertyFromTags(a.event.Tags, serviceKey)
}

// GetDeployment returns the name of the deployment
func (a ProblemAdapter) GetDeployment() string {
	return ""
}

// GetTestStrategy returns the used test strategy
func (a ProblemAdapter) GetTestStrategy() string {
	return ""
}

// GetDeploymentStrategy returns the used deployment strategy
func (a ProblemAdapter) GetDeploymentStrategy() string {
	return ""
}

// GetImage returns the deployed image
func (a ProblemAdapter) GetImage() string {
	return ""
}

// GetTag returns the deployed tag
func (a ProblemAdapter) GetTag() string {
	return ""
}

// GetLabels returns a map of labels
func (a ProblemAdapter) GetLabels() map[string]string {
	return nil
}

func (a ProblemAdapter) IsNotFromDynatrace() bool {
	return a.source != "dynatrace"
}

func (a ProblemAdapter) getProblemEventData() problemEventData {
	return problemEventData{
		EventData: keptnv2.EventData{
			Project: a.GetProject(),
			Stage:   a.GetStage(),
			Service: a.GetService(),
			Labels:  map[string]string{common.PROBLEMURL_LABEL: a.event.ProblemURL},
		},
		Problem: a.problemData,
	}
}

func getPropertyFromTags(tags, property string) string {
	// we analyze the tag list as its possible that the problem was raised for a specific monitored service that has keptn tags
	splittedTags := strings.Split(tags, ",")

	for _, tag := range splittedTags {
		tag = strings.TrimSpace(tag)
		split := strings.Split(tag, ":")
		if len(split) > 1 {
			if split[0] == property {
				return split[1]
			}
		}
	}
	return ""
}

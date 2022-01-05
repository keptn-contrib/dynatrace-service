package problem

import (
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
)

const remediationTaskName = "remediation"

// RawProblem is the raw problem datastructure
type RawProblem map[string]interface{}

type ProblemAdapterInterface interface {
	adapter.EventContentAdapter
	IsNotFromDynatrace() bool
	GetState() string
	GetPID() string
	GetProblemID() string
	IsOpen() bool
	IsResolved() bool
	GetProblemURL() string
	GetRawProblem() RawProblem
}

// ProblemAdapter is a content adaptor for events of type sh.keptn.event.action.finished
type ProblemAdapter struct {
	event      DTProblemEvent
	rawProblem RawProblem
	cloudEvent adapter.CloudEventAdapter
}

// NewProblemAdapterFromEvent creates a new ProblemAdapter from a cloudevents Event
func NewProblemAdapterFromEvent(e cloudevents.Event) (*ProblemAdapter, error) {
	ceAdapter := adapter.NewCloudEventAdapter(e)

	pData := &DTProblemEvent{}
	err := ceAdapter.PayloadAs(pData)
	if err != nil {
		return nil, err
	}

	var problem RawProblem
	err = ceAdapter.PayloadAs(&problem)
	if err != nil {
		return nil, err
	}

	// we need to set the project, stage and service names also from tags, if available
	setProjectStageAndServiceFromTags(pData)

	return &ProblemAdapter{
		event:      *pData,
		rawProblem: problem,
		cloudEvent: ceAdapter,
	}, nil
}

// GetShKeptnContext returns the shkeptncontext
func (a ProblemAdapter) GetShKeptnContext() string {
	return a.cloudEvent.GetShKeptnContext()
}

// GetSource returns the source specified in the CloudEvent context
func (a ProblemAdapter) GetSource() string {
	return a.cloudEvent.GetSource()
}

// GetEvent returns the event type
func (a ProblemAdapter) GetEvent() string {
	return a.cloudEvent.GetType()
}

// GetProject returns the project
func (a ProblemAdapter) GetProject() string {
	return a.event.KeptnProject
}

// GetStage returns the stage
func (a ProblemAdapter) GetStage() string {
	return a.event.KeptnStage
}

// GetService returns the service
func (a ProblemAdapter) GetService() string {
	return a.event.KeptnService
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

// GetLabels returns a map of labels
func (a ProblemAdapter) GetLabels() map[string]string {
	return nil
}

// IsNotFromDynatrace returns true if the source of the event is not dynatrace
func (a ProblemAdapter) IsNotFromDynatrace() bool {
	return a.cloudEvent.GetSource() != "dynatrace"
}

// GetState returns problem state as OPEN or RESOLVED
func (a ProblemAdapter) GetState() string {
	return a.event.State
}

// GetPID returns the PID
func (a ProblemAdapter) GetPID() string {
	return a.event.PID
}

// GetProblemID returns the problem ID
func (a ProblemAdapter) GetProblemID() string {
	return a.event.ProblemID
}

// GetProblemURL returns the problem URL
func (a ProblemAdapter) GetProblemURL() string {
	return a.event.ProblemURL
}

// GetRawProblem returns the raw problem datastructure
func (a ProblemAdapter) GetRawProblem() RawProblem {
	return a.rawProblem
}

// IsOpen returns true if the problem is open
func (a ProblemAdapter) IsOpen() bool {
	return a.GetState() == "OPEN"
}

// IsResolved returns true if the problem is resolved
func (a ProblemAdapter) IsResolved() bool {
	return a.GetState() == "RESOLVED"
}

func setProjectStageAndServiceFromTags(dtProblemEvent *DTProblemEvent) {
	// we analyze the tag list as its possible that the problem was raised for a specific monitored service that has keptn tags
	splittedTags := strings.Split(dtProblemEvent.Tags, ",")

	for _, tag := range splittedTags {
		tag = strings.TrimSpace(tag)
		split := strings.Split(tag, ":")
		if len(split) > 1 {
			if split[0] == "keptn_project" {
				dtProblemEvent.KeptnProject = split[1]
			}
			if split[0] == "keptn_stage" {
				dtProblemEvent.KeptnStage = split[1]
			}
			if split[0] == "keptn_service" {
				dtProblemEvent.KeptnService = split[1]
			}
		}
	}
}

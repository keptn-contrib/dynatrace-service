package problem

import (
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
	"strings"
)

const remediationTaskName = "remediation"

// ProblemAdapter is a content adaptor for events of type sh.keptn.event.action.finished
type ProblemAdapter struct {
	event   DTProblemEvent
	context string
	source  string
}

// NewProblemAdapter creates a new ProblemAdapter
func NewProblemAdapter(event DTProblemEvent, shkeptncontext, source string) ProblemAdapter {

	// we need to set the project, stage and service names also from tags, if available
	setProjectStageAndServiceFromTags(&event)
	return ProblemAdapter{event: event, context: shkeptncontext, source: source}
}

// NewProblemAdapterFromEvent creates a new ProblemAdapter from a cloudevents Event
func NewProblemAdapterFromEvent(e cloudevents.Event) (*ProblemAdapter, error) {
	pData := &DTProblemEvent{}
	err := e.DataAs(pData)
	if err != nil {
		return nil, fmt.Errorf("could not parse problem event payload: %v", err)
	}

	adapter := NewProblemAdapter(*pData, event.GetShKeptnContext(e), e.Source())
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
	if a.IsResolved() {
		return keptn.ProblemEventType
	}

	// fix get stage here below -> needs also tags evaluation
	return keptnv2.GetTriggeredEventType(fmt.Sprintf("%s.%s", a.GetStage(), remediationTaskName))
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

func (a ProblemAdapter) GetState() string {
	return a.event.State
}

func (a ProblemAdapter) GetPID() string {
	return a.event.PID
}

func (a ProblemAdapter) GetProblemID() string {
	return a.event.ProblemID
}

func (a ProblemAdapter) IsResolved() bool {
	return a.GetState() == "RESOLVED"
}

func (a ProblemAdapter) getClosedProblemEventData() keptn.ProblemEventData {
	problemData := keptn.ProblemEventData{
		State:          "CLOSED",
		PID:            a.GetPID(),
		ProblemID:      a.GetProblemID(),
		ProblemTitle:   a.event.ProblemTitle,
		ProblemDetails: json.RawMessage(marshalProblemDetails(a.event.ProblemDetails)),
		ProblemURL:     a.event.ProblemURL,
		ImpactedEntity: a.event.ImpactedEntity,
		Tags:           a.event.Tags,
		Project:        a.GetProject(),
		Stage:          a.GetStage(),
		Service:        a.GetService(),
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	problemData.Labels = make(map[string]string)
	problemData.Labels[common.PROBLEMURL_LABEL] = a.event.ProblemURL

	return problemData
}

func (a ProblemAdapter) getRemediationTriggeredEventData() remediationTriggeredEventData {
	remediationEventData := remediationTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: a.GetProject(),
			Stage:   a.GetStage(),
			Service: a.GetService(),
		},
		Problem: ProblemDetails{
			State:          "OPEN",
			PID:            a.GetPID(),
			ProblemID:      a.GetProblemID(),
			ProblemTitle:   a.event.ProblemTitle,
			ProblemDetails: json.RawMessage(marshalProblemDetails(a.event.ProblemDetails)),
			ProblemURL:     a.event.ProblemURL,
			ImpactedEntity: a.event.ImpactedEntity,
			Tags:           a.event.Tags,
		},
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	remediationEventData.Labels = make(map[string]string)
	remediationEventData.Labels[common.PROBLEMURL_LABEL] = a.event.ProblemURL

	return remediationEventData
}

func marshalProblemDetails(details DTProblemDetails) []byte {
	problemDetailsString, err := json.Marshal(details)
	if err != nil {
		log.WithError(err).Error("Could not marshal problem details")
	}

	return problemDetailsString
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

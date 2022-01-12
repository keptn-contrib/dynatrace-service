package problem

import (
	"encoding/json"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DTProblemEvent struct {
	ImpactedEntities []struct {
		Entity string `json:"entity"`
		Name   string `json:"name"`
		Type   string `json:"type"`
	} `json:"ImpactedEntities"`
	ImpactedEntity     string           `json:"ImpactedEntity"`
	PID                string           `json:"PID"`
	ProblemDetails     DTProblemDetails `json:"ProblemDetails"`
	ProblemDetailsHTML string           `json:"ProblemDetailsHTML"`
	ProblemDetailsText string           `json:"ProblemDetailsText"`
	ProblemID          string           `json:"ProblemID"`
	ProblemImpact      string           `json:"ProblemImpact"`
	ProblemSeverity    string           `json:"ProblemSeverity"`
	ProblemTitle       string           `json:"ProblemTitle"`
	ProblemURL         string           `json:"ProblemURL"`
	State              string           `json:"State"`
	Tags               string           `json:"Tags"`
	EventContext       struct {
		KeptnContext string `json:"keptnContext"`
		Token        string `json:"token"`
	} `json:"eventContext"`
	KeptnProject string `json:"KeptnProject"`
	KeptnService string `json:"KeptnService"`
	KeptnStage   string `json:"KeptnStage"`
}

type DTProblemDetails struct {
	DisplayName   string `json:"displayName"`
	EndTime       int    `json:"endTime"`
	HasRootCause  bool   `json:"hasRootCause"`
	ID            string `json:"id"`
	ImpactLevel   string `json:"impactLevel"`
	SeverityLevel string `json:"severityLevel"`
	StartTime     int64  `json:"startTime"`
	Status        string `json:"status"`
}

type ProblemEventHandler struct {
	event  ProblemAdapterInterface
	client keptn.ClientInterface
}

func NewProblemEventHandler(event ProblemAdapterInterface, client keptn.ClientInterface) ProblemEventHandler {
	return ProblemEventHandler{
		event:  event,
		client: client,
	}
}

type RemediationTriggeredEventData struct {
	keptnv2.EventData

	// Problem contains details about the problem
	Problem ProblemDetails `json:"problem"`
}

type ProblemDetails struct {
	// State is the state of the problem; possible values are: OPEN, RESOLVED
	State string `json:"State,omitempty" jsonschema:"enum=open,enum=resolved"`

	// ProblemID is a unique system identifier of the reported problem
	ProblemID string `json:"ProblemID"`

	// ProblemTitle is the display number of the reported problem.
	ProblemTitle string `json:"ProblemTitle"`

	// ProblemDetails are all problem event details including root cause
	ProblemDetails json.RawMessage `json:"ProblemDetails"`

	// ProblemDetailsHTML are all problem event details including root cause as an HTML-formatted string
	ProblemDetailsHTML string `json:"ProblemDetailsHTML,omitempty"`

	// ProblemDetailsText are all problem event details including root cause as a text-formatted string.
	ProblemDetailsText string `json:"ProblemDetailsText,omitempty"`

	// PID is a unique system identifier of the reported problem.
	PID string `json:"PID"`

	// ProblemImpact is the impact level of the problem. Possible values are APPLICATION, SERVICE, or INFRASTRUCTURE.
	ProblemImpact string `json:"ProblemImpact,omitempty"`

	// ProblemSeverity is the severity level of the problem. Possible values are AVAILABILITY, ERROR, PERFORMANCE, RESOURCE_CONTENTION, or CUSTOM_ALERT.const
	ProblemSeverity string `json:"ProblemSeverity,omitempty"`

	// ProblemURL is a back link to the original problem
	ProblemURL string `json:"ProblemURL,omitempty"`

	// ImpactedEntity is an identifier of the impacted entity
	ImpactedEntity string `json:"ImpactedEntity,omitempty"`

	// Tags is a comma separated list of tags that are defined for all impacted entities.
	Tags string `json:"Tags,omitempty"`
}

func (eh ProblemEventHandler) HandleEvent() error {
	if eh.event.IsNotFromDynatrace() {
		log.WithField("eventSource", eh.event.GetSource()).Debug("Will not handle problem event that did not come from a Dynatrace Problem Notification")
		return nil
	}

	log.WithFields(
		log.Fields{
			"PID":       eh.event.GetPID(),
			"problemId": eh.event.GetProblemID(),
			"state":     eh.event.GetState(),
		}).Info("Received event")

	if eh.event.IsResolved() {
		return eh.handleClosedProblemFromDT()
	}

	if eh.event.GetStage() == "" {
		log.Debug("Dropping open problen event as it has no stage")
		return nil
	}

	return eh.handleOpenedProblemFromDT()
}

func (eh ProblemEventHandler) handleClosedProblemFromDT() error {
	err := eh.sendEvent(NewProblemClosedEventFactory(eh.event))
	if err != nil {
		return err
	}

	log.WithField("PID", eh.event.GetPID()).Debug("Successfully sent Keptn PROBLEM CLOSED event")
	return nil
}

func (eh ProblemEventHandler) handleOpenedProblemFromDT() error {
	err := eh.sendEvent(NewRemediationTriggeredEventFactory(eh.event))
	if err != nil {
		return err
	}

	log.WithField("PID", eh.event.GetPID()).Debug("Successfully sent Keptn PROBLEM OPEN event")
	return nil
}

func (eh ProblemEventHandler) sendEvent(factory adapter.CloudEventFactoryInterface) error {
	err := eh.client.SendCloudEvent(factory)
	if err != nil {
		log.WithError(err).Error("Failed to send cloud event")
	}

	return err
}

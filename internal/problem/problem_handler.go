package problem

import (
	"encoding/json"
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DTProblemEvent struct {
	ImpactedEntities []struct {
		Entity string `json:"entity"`
		Name   string `json:"name"`
		Type   string `json:"type"`
	} `json:"ImpactedEntities"`
	ImpactedEntity string           `json:"ImpactedEntity"`
	PID            string           `json:"PID"`
	ProblemDetails DTProblemDetails `json:"ProblemDetails"`
	ProblemID      string           `json:"ProblemID"`
	ProblemTitle   string           `json:"ProblemTitle"`
	ProblemURL     string           `json:"ProblemURL"`
	State          string           `json:"State"`
	Tags           string           `json:"Tags"`
	EventContext   struct {
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
	Event cloudevents.Event
}

type remediationTriggeredEventData struct {
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
	// PID is a unique system identifier of the reported problem.
	PID string `json:"PID"`
	// ImpactedEntity is an identifier of the impacted entity
	// ProblemURL is a back link to the original problem
	ProblemURL     string `json:"ProblemURL,omitempty"`
	ImpactedEntity string `json:"ImpactedEntity,omitempty"`
	// Tags is a comma separated list of tags that are defined for all impacted entities.
	Tags string `json:"Tags,omitempty"`
}

const eventbroker = "EVENTBROKER"

func (eh ProblemEventHandler) HandleEvent() error {
	keptnEvent, err := NewProblemAdapterFromEvent(eh.Event)
	if err != nil {
		log.WithError(err).Error("Could not unmarshal problem event")
		return err
	}

	if keptnEvent.IsNotFromDynatrace() {
		log.WithField("eventSource", keptnEvent.GetSource()).Debug("Will not handle problem event that did not come from a Dynatrace Problem Notification")
		return nil
	}

	// Log the problem ID and state for better troubleshooting
	log.WithFields(
		log.Fields{
			"PID":       keptnEvent.GetPID(),
			"problemId": keptnEvent.GetProblemID(),
			"state":     keptnEvent.GetState(),
		}).Info("Received event")

	// ignore problem events if they are closed
	if keptnEvent.IsResolved() {
		return eh.handleClosedProblemFromDT(keptnEvent)
	}

	return eh.handleOpenedProblemFromDT(keptnEvent)
}

func (eh ProblemEventHandler) handleClosedProblemFromDT(keptnEvent *ProblemAdapter) error {

	err := createAndSendCE(keptnEvent.getClosedProblemEventData(), keptnEvent.GetShKeptnContext(), keptnEvent.GetEvent())
	if err != nil {
		log.WithError(err).Error("Could not send cloud event")
		return err
	}
	log.WithField("PID", keptnEvent.GetPID()).Debug("Successfully sent Keptn PROBLEM CLOSED event")
	return nil
}

func (eh ProblemEventHandler) handleOpenedProblemFromDT(keptnEvent *ProblemAdapter) error {

	// Send a sh.keptn.event.${STAGE}.remediation.triggered event
	err := createAndSendCE(keptnEvent.getRemediationTriggeredEventData(), keptnEvent.GetShKeptnContext(), keptnEvent.GetEvent())
	if err != nil {
		log.WithError(err).Error("Could not send cloud event")
		return err
	}
	log.WithField("PID", keptnEvent.GetPID()).Debug("Successfully sent Keptn PROBLEM OPEN event")
	return nil
}

func createAndSendCE(problemData interface{}, shkeptncontext string, eventType string) error {
	ce := cloudevents.NewEvent()
	ce.SetType(eventType)
	ce.SetSource(event.GetEventSource())
	ce.SetDataContentType(cloudevents.ApplicationJSON)
	ce.SetData(cloudevents.ApplicationJSON, problemData)
	ce.SetExtension("shkeptncontext", shkeptncontext)

	keptnHandler, err := keptnv2.NewKeptn(&ce, keptncommon.KeptnOpts{})
	if err != nil {
		return errors.New("Could not create Keptn Handler: " + err.Error())
	}

	if err := keptnHandler.SendCloudEvent(ce); err != nil {
		return errors.New("Failed to send cloudevent:, " + err.Error())
	}

	return nil
}

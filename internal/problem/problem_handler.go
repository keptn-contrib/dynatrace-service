package problem

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptn "github.com/keptn/go-utils/pkg/lib"
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
	ImpactedEntity string `json:"ImpactedEntity"`
	PID            string `json:"PID"`
	ProblemDetails struct {
		DisplayName   string `json:"displayName"`
		EndTime       int    `json:"endTime"`
		HasRootCause  bool   `json:"hasRootCause"`
		ID            string `json:"id"`
		ImpactLevel   string `json:"impactLevel"`
		SeverityLevel string `json:"severityLevel"`
		StartTime     int64  `json:"startTime"`
		Status        string `json:"status"`
	} `json:"ProblemDetails"`
	ProblemID    string `json:"ProblemID"`
	ProblemTitle string `json:"ProblemTitle"`
	ProblemURL   string `json:"ProblemURL"`
	State        string `json:"State"`
	Tags         string `json:"Tags"`
	EventContext struct {
		KeptnContext string `json:"keptnContext"`
		Token        string `json:"token"`
	} `json:"eventContext"`
	KeptnProject string `json:"KeptnProject"`
	KeptnService string `json:"KeptnService"`
	KeptnStage   string `json:"KeptnStage"`
}

type ProblemEventHandler struct {
	Event cloudevents.Event
}

type remediationTriggeredEventData struct {
	keptnv2.EventData

	// Problem contains details about the problem
	Problem ProblemDetails `json:"problem"`
}

const remediationTaskName = "remediation"

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

	if eh.Event.Source() != "dynatrace" {
		log.WithField("eventSource", eh.Event.Source()).Debug("Will not handle problem event that did not come from a Dynatrace Problem Notification")
		return nil
	}

	shkeptncontext := event.GetShKeptnContext(eh.Event)

	dtProblemEvent := &DTProblemEvent{}
	err := eh.Event.DataAs(dtProblemEvent)

	if err != nil {
		log.WithError(err).Error("Could not map received event to datastructure")
		return err
	}

	// Log the problem ID and state for better troubleshooting
	log.WithFields(
		log.Fields{
			"PID":       dtProblemEvent.PID,
			"problemId": dtProblemEvent.ProblemID,
			"state":     dtProblemEvent.State,
		}).Info("Received event")

	// ignore problem events if they are closed
	if dtProblemEvent.State == "RESOLVED" {
		return eh.handleClosedProblemFromDT(dtProblemEvent, shkeptncontext)
	}

	return eh.handleOpenedProblemFromDT(dtProblemEvent, shkeptncontext)
}

func (eh ProblemEventHandler) handleClosedProblemFromDT(dtProblemEvent *DTProblemEvent, shkeptncontext string) error {
	problemDetailsString, err := json.Marshal(dtProblemEvent.ProblemDetails)

	project, stage, service := eh.extractContextFromDynatraceProblem(dtProblemEvent)

	newProblemData := keptn.ProblemEventData{
		State:          "CLOSED",
		PID:            dtProblemEvent.PID,
		ProblemID:      dtProblemEvent.ProblemID,
		ProblemTitle:   dtProblemEvent.ProblemTitle,
		ProblemDetails: json.RawMessage(problemDetailsString),
		ProblemURL:     dtProblemEvent.ProblemURL,
		ImpactedEntity: dtProblemEvent.ImpactedEntity,
		Tags:           dtProblemEvent.Tags,
		Project:        project,
		Stage:          stage,
		Service:        service,
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	newProblemData.Labels = make(map[string]string)
	newProblemData.Labels[common.PROBLEMURL_LABEL] = dtProblemEvent.ProblemURL

	err = createAndSendCE(newProblemData, shkeptncontext, keptn.ProblemEventType)
	if err != nil {
		log.WithError(err).Error("Could not send cloud event")
		return err
	}
	log.WithField("PID", dtProblemEvent.PID).Debug("Successfully sent Keptn PROBLEM CLOSED event")
	return nil
}

func (eh ProblemEventHandler) handleOpenedProblemFromDT(dtProblemEvent *DTProblemEvent, shkeptncontext string) error {
	problemDetailsString, err := json.Marshal(dtProblemEvent.ProblemDetails)

	project, stage, service := eh.extractContextFromDynatraceProblem(dtProblemEvent)

	remediationEventData := remediationTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: project,
			Stage:   stage,
			Service: service,
		},
		Problem: ProblemDetails{
			State:          "OPEN",
			PID:            dtProblemEvent.PID,
			ProblemID:      dtProblemEvent.ProblemID,
			ProblemTitle:   dtProblemEvent.ProblemTitle,
			ProblemDetails: json.RawMessage(problemDetailsString),
			ProblemURL:     dtProblemEvent.ProblemURL,
			ImpactedEntity: dtProblemEvent.ImpactedEntity,
			Tags:           dtProblemEvent.Tags,
		},
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	remediationEventData.Labels = make(map[string]string)
	remediationEventData.Labels[common.PROBLEMURL_LABEL] = dtProblemEvent.ProblemURL

	// Send a sh.keptn.event.${STAGE}.remediation.triggered event
	err = createAndSendCE(remediationEventData, shkeptncontext, keptnv2.GetTriggeredEventType(
		fmt.Sprintf("%s.%s", stage, remediationTaskName),
	))
	if err != nil {
		log.WithError(err).Error("Could not send cloud event")
		return err
	}
	log.WithField("PID", dtProblemEvent.PID).Debug("Successfully sent Keptn PROBLEM OPEN event")
	return nil
}

func (eh ProblemEventHandler) extractContextFromDynatraceProblem(dtProblemEvent *DTProblemEvent) (string, string, string) {

	// First we look if project, stage and service was passed in via the problem data fields and use them as defaults
	project := dtProblemEvent.KeptnProject
	stage := dtProblemEvent.KeptnStage
	service := dtProblemEvent.KeptnService

	// Second we analyze the tag list as its possible that the problem was raised for a specific monitored service that has keptn tags
	splittedTags := strings.Split(dtProblemEvent.Tags, ",")

	for _, tag := range splittedTags {
		tag = strings.TrimSpace(tag)
		split := strings.Split(tag, ":")
		if len(split) > 1 {
			if split[0] == "keptn_project" {
				project = split[1]
			}
			if split[0] == "keptn_stage" {
				stage = split[1]
			}
			if split[0] == "keptn_service" {
				service = split[1]
			}
		}
	}
	return project, stage, service
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

func getServiceEndpoint(service string) (url.URL, error) {
	url, err := url.Parse(os.Getenv(service))
	if err != nil {
		return *url, fmt.Errorf("Failed to retrieve value from ENVIRONMENT_VARIABLE: %s", service)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
	}

	return *url, nil
}

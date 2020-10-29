package event_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/url"
	"os"
	"strings"
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
}

type ProblemEventHandler struct {
	Logger *keptncommon.Logger
	Event  cloudevents.Event
}

const eventbroker = "EVENTBROKER"

func (eh ProblemEventHandler) HandleEvent() error {

	if eh.Event.Source() != "dynatrace" {
		eh.Logger.Debug("Will not handle problem event that did not come from a Dynatrace Problem Notification (event source = " + eh.Event.Source() + ")")
		return nil
	}
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)
	dtProblemEvent := &DTProblemEvent{}
	err := eh.Event.DataAs(dtProblemEvent)

	if err != nil {
		eh.Logger.Error("Could not map received event to datastructure: " + err.Error())
		return err
	}

	// ignore problem events if they are closed
	if dtProblemEvent.State == "RESOLVED" {
		eh.Logger.Info("Received RESOLVED problem notification")
		return eh.handleClosedProblemFromDT(dtProblemEvent, shkeptncontext)
	}

	return eh.handleOpenedProblemFromDT(dtProblemEvent, shkeptncontext)
}

func (eh ProblemEventHandler) handleClosedProblemFromDT(dtProblemEvent *DTProblemEvent, shkeptncontext string) error {
	problemDetailsString, err := json.Marshal(dtProblemEvent.ProblemDetails)

	project, stage, service := eh.extractContextFromTags(dtProblemEvent)

	newProblemData := keptn.ProblemEventData{
		State:          "CLOSED",
		PID:            dtProblemEvent.PID,
		ProblemID:      dtProblemEvent.ProblemID,
		ProblemTitle:   dtProblemEvent.ProblemTitle,
		ProblemDetails: json.RawMessage(problemDetailsString),
		ProblemURL:     dtProblemEvent.ProblemURL,
		ImpactedEntity: dtProblemEvent.ImpactedEntity,
		Project:        project,
		Stage:          stage,
		Service:        service,
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	newProblemData.Labels = make(map[string]string)
	newProblemData.Labels["Problem URL"] = dtProblemEvent.ProblemURL

	eh.Logger.Debug("Sending event to eventbroker")
	err = createAndSendCE(eventbroker, newProblemData, shkeptncontext, "sh.keptn.events.problem")
	if err != nil {
		eh.Logger.Error("Could not send cloud event: " + err.Error())
		return err
	}
	eh.Logger.Debug("Event successfully dispatched to eventbroker")
	return nil
}

func (eh ProblemEventHandler) handleOpenedProblemFromDT(dtProblemEvent *DTProblemEvent, shkeptncontext string) error {
	problemDetailsString, err := json.Marshal(dtProblemEvent.ProblemDetails)

	project, stage, service := eh.extractContextFromTags(dtProblemEvent)

	newProblemData := keptn.ProblemEventData{
		State:          "OPEN",
		PID:            dtProblemEvent.PID,
		ProblemID:      dtProblemEvent.ProblemID,
		ProblemTitle:   dtProblemEvent.ProblemTitle,
		ProblemDetails: json.RawMessage(problemDetailsString),
		ProblemURL:     dtProblemEvent.ProblemURL,
		ImpactedEntity: dtProblemEvent.ImpactedEntity,
		Project:        project,
		Stage:          stage,
		Service:        service,
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	newProblemData.Labels = make(map[string]string)
	newProblemData.Labels["Problem URL"] = dtProblemEvent.ProblemURL

	eh.Logger.Debug("Sending event to eventbroker")
	err = createAndSendCE(eventbroker, newProblemData, shkeptncontext, keptn.ProblemOpenEventType)
	if err != nil {
		eh.Logger.Error("Could not send cloud event: " + err.Error())
		return err
	}
	eh.Logger.Debug("Event successfully dispatched to eventbroker")
	return nil
}

func (eh ProblemEventHandler) extractContextFromTags(dtProblemEvent *DTProblemEvent) (string, string, string) {
	splittedTags := strings.Split(dtProblemEvent.Tags, ",")

	project := ""
	stage := ""
	service := ""

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

func createAndSendCE(eventbroker string, problemData keptn.ProblemEventData, shkeptncontext string, eventType string) error {
	source, _ := url.Parse("dynatrace-service")

	ce := cloudevents.NewEvent()
	ce.SetType(eventType)
	ce.SetSource(source.String())
	ce.SetDataContentType(cloudevents.ApplicationJSON)
	ce.SetData(cloudevents.ApplicationJSON, problemData)

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

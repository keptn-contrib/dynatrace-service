package event_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cloudeventsclient "github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	keptn "github.com/keptn/go-utils/pkg/lib"
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
	Logger *keptn.Logger
	Event  cloudevents.Event
}

const eventbroker = "EVENTBROKER"

const TEST_NOTIFICATION_CONTEXT = "39393939-3920-4020-a020-202020202020"

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

	// Log the problem ID and state for better troubleshooting
	eh.Logger.Info(fmt.Sprintf("Received PID=%s, ProblemID=%s, State=%s", dtProblemEvent.PID, dtProblemEvent.ProblemID, dtProblemEvent.State))

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

	err = createAndSendCE(eventbroker, newProblemData, shkeptncontext, keptn.ProblemEventType)
	if err != nil {
		eh.Logger.Error("Could not send cloud event: " + err.Error())
		return err
	}
	eh.Logger.Debug(fmt.Sprintf("Successfully sent Keptn PROBLEM CLOSED event for PID: %s", dtProblemEvent.PID))
	return nil
}

func (eh ProblemEventHandler) handleOpenedProblemFromDT(dtProblemEvent *DTProblemEvent, shkeptncontext string) error {
	problemDetailsString, err := json.Marshal(dtProblemEvent.ProblemDetails)

	project, stage, service := eh.extractContextFromDynatraceProblem(dtProblemEvent)

	// check for special keptn context that comes from a Dynatrace "Send Test Notification Event"
	if strings.Compare(shkeptncontext, TEST_NOTIFICATION_CONTEXT) == 0 {
		// if the incoming context is the one for the test event -> thats the one that gets created when {PID} == "99999" then clear it so that we get a new context
		uuid.SetRand(nil)
		shkeptncontext = uuid.New().String()
	}

	newProblemData := keptn.ProblemEventData{
		State:          "OPEN",
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

	err = createAndSendCE(eventbroker, newProblemData, shkeptncontext, keptn.ProblemOpenEventType)
	if err != nil {
		eh.Logger.Error("Could not send cloud event: " + err.Error())
		return err
	}
	eh.Logger.Debug(fmt.Sprintf("Successfully sent Keptn PROBLEM OPEN event for PID: %s", dtProblemEvent.PID))
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

func createAndSendCE(eventbroker string, problemData keptn.ProblemEventData, shkeptncontext string, eventType string) error {
	source, _ := url.Parse("dynatrace-service")
	contentType := "application/json"

	endPoint, err := getServiceEndpoint(eventbroker)

	ce := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        eventType,
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": shkeptncontext},
		}.AsV02(),
		Data: problemData,
	}

	t, err := cloudeventshttp.New(
		cloudeventshttp.WithTarget(endPoint.String()),
		cloudeventshttp.WithEncoding(cloudeventshttp.StructuredV02),
	)
	if err != nil {
		return errors.New("Failed to create transport:" + err.Error())
	}

	c, err := cloudeventsclient.New(t)
	if err != nil {
		return errors.New("Failed to create HTTP client:" + err.Error())
	}

	if _, _, err := c.Send(context.Background(), ce); err != nil {
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

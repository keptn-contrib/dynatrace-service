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

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cloudeventsclient "github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
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
	State        string `json:"State"`
	Tags         string `json:"Tags"`
	EventContext struct {
		KeptnContext string `json:"keptnContext"`
		Token        string `json:"token"`
	} `json:"eventContext"`
}

type ProblemEventHandler struct {
	Logger *keptnutils.Logger
	Event  cloudevents.Event
}

const eventbroker = "EVENTBROKER"

func (eh ProblemEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)
	dtProblemEvent := &DTProblemEvent{}
	err := eh.Event.DataAs(dtProblemEvent)

	if err != nil {
		return err
		eh.Logger.Error("Could not map received event to datastructure: " + err.Error())
	}

	// ignore problem events if they are closed
	if dtProblemEvent.ProblemDetails.Status == "CLOSED" {
		eh.Logger.Info("Received CLOSED problem")
		return nil
	}

	problemDetailsString, err := json.Marshal(dtProblemEvent.ProblemDetails)

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
	newProblemData := keptnevents.ProblemEventData{
		State:          "OPEN",
		PID:            dtProblemEvent.PID,
		ProblemID:      dtProblemEvent.ProblemID,
		ProblemTitle:   dtProblemEvent.ProblemTitle,
		ProblemDetails: json.RawMessage(problemDetailsString),
		ImpactedEntity: dtProblemEvent.ImpactedEntity,
		Project:        project,
		Stage:          stage,
		Service:        service,
	}

	eh.Logger.Debug("Sending event to eventbroker")
	err = createAndSendCE(eventbroker, newProblemData, shkeptncontext)
	if err != nil {
		eh.Logger.Error("Could not send cloud event: " + err.Error())
		return err
	}
	eh.Logger.Debug("Event successfully dispatched to eventbroker")
	return nil
}

func createAndSendCE(eventbroker string, problemData keptnevents.ProblemEventData, shkeptncontext string) error {
	source, _ := url.Parse("dynatrace")
	contentType := "application/json"

	endPoint, err := getServiceEndpoint(eventbroker)

	ce := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        keptnevents.ProblemOpenEventType,
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

	if _, err := c.Send(context.Background(), ce); err != nil {
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

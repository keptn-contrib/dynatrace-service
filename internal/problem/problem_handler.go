package problem

import (
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DTProblemEvent struct {
	State      string `json:"State"`
	ProblemURL string `json:"ProblemURL"`
	Tags       string `json:"Tags"`
}

type ProblemEventHandler struct {
	event *ProblemAdapter
}

func NewProblemEventHandler(event *ProblemAdapter) ProblemEventHandler {
	return ProblemEventHandler{
		event: event,
	}
}

type problemEventData struct {
	keptnv2.EventData

	// Problem contains details about the problem
	Problem interface{} `json:"problem"`
}

func (eh ProblemEventHandler) HandleEvent() error {
	if eh.event.IsNotFromDynatrace() {
		log.WithField("eventSource", eh.event.GetSource()).Debug("Will not handle problem event that did not come from a Dynatrace Problem Notification")
		return nil
	}

	// Send a sh.keptn.event.${STAGE}.remediation.triggered event
	err := createAndSendCE(eh.event.getProblemEventData(), eh.event.GetShKeptnContext(), eh.event.GetEvent())
	if err != nil {
		log.WithError(err).Error("Could not send cloud event")
		return err
	}
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

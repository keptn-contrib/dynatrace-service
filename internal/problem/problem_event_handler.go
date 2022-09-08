package problem

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DTProblemEvent struct {
	PID          string `json:"PID"`
	ProblemID    string `json:"ProblemID"`
	ProblemURL   string `json:"ProblemURL"`
	State        string `json:"State"`
	Tags         string `json:"Tags"`
	KeptnProject string `json:"KeptnProject"`
	KeptnService string `json:"KeptnService"`
	KeptnStage   string `json:"KeptnStage"`
}

type ProblemEventHandler struct {
	event             ProblemAdapterInterface
	eventSenderClient keptn.EventSenderClientInterface
}

func NewProblemEventHandler(event ProblemAdapterInterface, client keptn.EventSenderClientInterface) ProblemEventHandler {
	return ProblemEventHandler{
		event:             event,
		eventSenderClient: client,
	}
}

type RemediationTriggeredEventData struct {
	keptnv2.EventData

	// Problem contains details about the problem
	Problem RawProblem `json:"problem"`
}

// HandleEvent handles a problem event.
func (eh ProblemEventHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	if eh.event.IsNotFromDynatrace() {
		log.WithField("eventSource", eh.event.GetSource()).Debug("Will not handle problem event that did not come from a Dynatrace Problem Notification")
		return nil
	}

	if eh.event.IsOpen() {
		return eh.handleOpenedProblemFromDT()
	}
	if eh.event.IsResolved() {
		return eh.handleClosedProblemFromDT()
	}

	return nil
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
	if eh.event.GetStage() == "" {
		log.Debug("Dropping open problem event as it has no stage")
		return nil
	}

	err := eh.sendEvent(NewRemediationTriggeredEventFactory(eh.event))
	if err != nil {
		return err
	}

	log.WithField("PID", eh.event.GetPID()).Debug("Successfully sent Keptn PROBLEM OPEN event")
	return nil
}

func (eh ProblemEventHandler) sendEvent(factory adapter.CloudEventFactoryInterface) error {
	err := eh.eventSenderClient.SendCloudEvent(factory)
	if err != nil {
		log.WithError(err).Error("Failed to send cloud event")
	}

	return err
}

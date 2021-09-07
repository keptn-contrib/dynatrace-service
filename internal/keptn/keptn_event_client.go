package keptn

import (
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

type EventClientBaseInterface interface {
	GetEvents(filter *keptnapi.EventFilter) ([]*models.KeptnContextExtendedCE, error)
}

type EventClientBase struct {
	client *keptnapi.EventHandler
}

func NewEventClientBase() *EventClientBase {
	return &EventClientBase{
		client: keptnapi.NewEventHandler(os.Getenv("DATASTORE")),
	}
}

func (c *EventClientBase) GetEvents(filter *keptnapi.EventFilter) ([]*models.KeptnContextExtendedCE, error) {
	events, err := c.client.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("could not get events: %s", err.GetMessage())
	}

	return events, nil
}

const problemURLLabel = "Problem URL"

type EventClientInterface interface {
	IsPartOfRemediation(event adapter.EventContentAdapter) (bool, error)
	FindProblemID(keptnEvent adapter.EventContentAdapter) (string, error)
	GetImageAndTag(keptnEvent adapter.EventContentAdapter) common.ImageAndTag
}

type EventClient struct {
	client EventClientBaseInterface
}

func NewEventClient(client EventClientBaseInterface) *EventClient {
	return &EventClient{
		client: client,
	}
}

func NewDefaultEventClient() *EventClient {
	return NewEventClient(
		NewEventClientBase())
}

// IsPartOfRemediation checks whether the evaluation.finished event is part of a remediation task sequence
func (c *EventClient) IsPartOfRemediation(event adapter.EventContentAdapter) (bool, error) {
	events, err := c.client.GetEvents(
		&keptnapi.EventFilter{
			Project:      event.GetProject(),
			Stage:        event.GetStage(),
			Service:      event.GetService(),
			EventType:    keptnv2.GetTriggeredEventType("remediation"),
			KeptnContext: event.GetShKeptnContext(),
		})

	if err != nil {
		return false, err
	}

	if events == nil || len(events) == 0 {
		return false, nil
	}

	return true, nil
}

// FindProblemID finds the Problem ID that is associated with this Keptn Workflow
// It first parses it from Problem URL label - if it cant be found there it will look for the Initial Problem Open Event and gets the ID from there!
func (c *EventClient) FindProblemID(keptnEvent adapter.EventContentAdapter) (string, error) {

	// Step 1 - see if we have a Problem Url in the labels
	// iterate through the labels and find Problem URL
	for labelName, labelValue := range keptnEvent.GetLabels() {
		if labelName == problemURLLabel {
			// the value should be of form https://dynatracetenant/#problems/problemdetails;pid=8485558334848276629_1604413609638V2
			// so - lets get the last part after pid=

			ix := strings.LastIndex(labelValue, ";pid=")
			if ix > 0 {
				return labelValue[ix+5:], nil
			}
		}
	}

	// Step 2 - lets see if we have a ProblemOpenEvent for this KeptnContext - if so - we try to extract the Problem ID
	events, err := c.client.GetEvents(
		&keptnapi.EventFilter{
			Project:      keptnEvent.GetProject(),
			EventType:    keptncommon.ProblemOpenEventType,
			KeptnContext: keptnEvent.GetShKeptnContext(),
		})

	if err != nil {
		return "", fmt.Errorf("cannot send DT problem comment: Could not retrieve problem.open event for incoming event: %s", err.Error())
	}

	if len(events) == 0 {
		return "", errors.New("cannot send DT problem comment: Could not retrieve problem.open event for incoming event: no events returned")
	}

	problemOpenEvent := &keptncommon.ProblemEventData{}
	err = keptnv2.Decode(events[0].Data, problemOpenEvent)
	if err != nil {
		return "", fmt.Errorf("could not decode problem.open event: %s", err.Error())
	}

	if problemOpenEvent.PID == "" {
		return "", errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	return problemOpenEvent.PID, nil
}

func (c *EventClient) GetImageAndTag(event adapter.EventContentAdapter) common.ImageAndTag {

	events, err := c.client.GetEvents(
		&keptnapi.EventFilter{
			Project:      event.GetProject(),
			Stage:        event.GetStage(),
			Service:      event.GetService(),
			EventType:    keptnv2.GetTriggeredEventType(keptnv2.DeploymentTaskName),
			KeptnContext: event.GetShKeptnContext(),
		})

	if err != nil {
		log.WithError(err).Error("Could not retrieve image and tag for event")
		return common.NewNotAvailableImageAndTag()
	}

	if len(events) == 0 {
		return common.NewNotAvailableImageAndTag()
	}

	triggeredData := &keptnv2.DeploymentTriggeredEventData{}
	err = keptnv2.Decode(events[0].Data, triggeredData)
	if err != nil {
		log.WithError(err).Error("Could not decode event data")
		return common.NewNotAvailableImageAndTag()
	}

	for key, value := range triggeredData.ConfigurationChange.Values {
		if strings.HasSuffix(key, "image") {
			imageAndTag := value.(string)
			return common.NewImageAndTag(
				getImage(imageAndTag),
				getTag(imageAndTag))
		}
	}

	return common.NewNotAvailableImageAndTag()
}

// getImage returns the deployed image
func getImage(imageAndTag string) string {
	if imageAndTag == common.NotAvailable {
		return common.NotAvailable
	}

	split := strings.Split(imageAndTag, ":")
	return split[0]
}

// getTag returns the deployed tag
func getTag(imageAndTag string) string {
	if imageAndTag == common.NotAvailable {
		return common.NotAvailable
	}

	split := strings.Split(imageAndTag, ":")
	if len(split) == 1 {
		return common.NotAvailable
	}

	return split[1]
}

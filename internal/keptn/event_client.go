package keptn

import (
	"errors"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	api "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

const problemURLLabel = "Problem URL"

// EventClientInterface encapsulates functionality built on top of Keptn events.
type EventClientInterface interface {
	// IsPartOfRemediation checks whether the sequence includes a remediation triggered event or returns an error.
	IsPartOfRemediation(event adapter.EventContentAdapter) (bool, error)

	// FindProblemID finds the Problem ID that is associated with the specified Keptn event or returns an error.
	FindProblemID(keptnEvent adapter.EventContentAdapter) (string, error)

	// GetImageAndTag extracts the image and tag associated with a deployment triggered as part of the sequence.
	GetImageAndTag(keptnEvent adapter.EventContentAdapter) common.ImageAndTag
}

// EventClient implements offers EventClientInterface using api.EventsV1Interface.
type EventClient struct {
	client api.EventsV1Interface
}

// NewEventClient creates a new EventClient using the specified api.EventsV1Interface.
func NewEventClient(client api.EventsV1Interface) *EventClient {
	return &EventClient{
		client: client,
	}
}

// IsPartOfRemediation checks whether the sequence includes a remediation triggered event or returns an error.
func (c *EventClient) IsPartOfRemediation(event adapter.EventContentAdapter) (bool, error) {
	events, err := c.client.GetEvents(
		&api.EventFilter{
			Project:      event.GetProject(),
			Stage:        event.GetStage(),
			Service:      event.GetService(),
			EventType:    keptnv2.GetTriggeredEventType("remediation"),
			KeptnContext: event.GetShKeptnContext(),
		})

	if err != nil {
		return false, errors.New(err.GetMessage())
	}

	if len(events) == 0 {
		return false, nil
	}

	return true, nil
}

// FindProblemID finds the Problem ID that is associated with the specified Keptn event or returns an error.
// It first parses it from Problem URL label and if it cant be found there it will look for the Initial Problem Open Event and gets the ID from there.
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
	events, mErr := c.client.GetEvents(
		&api.EventFilter{
			Project:      keptnEvent.GetProject(),
			EventType:    keptncommon.ProblemOpenEventType,
			KeptnContext: keptnEvent.GetShKeptnContext(),
		})

	if mErr != nil {
		return "", fmt.Errorf("cannot send DT problem comment: Could not retrieve problem.open event for incoming event: %s", mErr.GetMessage())
	}

	if len(events) == 0 {
		return "", errors.New("cannot send DT problem comment: Could not retrieve problem.open event for incoming event: no events returned")
	}

	problemOpenEvent := &keptncommon.ProblemEventData{}
	err := keptnv2.Decode(events[0].Data, problemOpenEvent)
	if err != nil {
		return "", fmt.Errorf("could not decode problem.open event: %w", err)
	}

	if problemOpenEvent.PID == "" {
		return "", errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	return problemOpenEvent.PID, nil
}

// GetImageAndTag extracts the image and tag associated with a deployment triggered as part of the sequence.
func (c *EventClient) GetImageAndTag(event adapter.EventContentAdapter) common.ImageAndTag {

	events, mErr := c.client.GetEvents(
		&api.EventFilter{
			Project:      event.GetProject(),
			Stage:        event.GetStage(),
			Service:      event.GetService(),
			EventType:    keptnv2.GetTriggeredEventType(keptnv2.DeploymentTaskName),
			KeptnContext: event.GetShKeptnContext(),
		})

	if mErr != nil {
		log.WithError(errors.New(mErr.GetMessage())).Error("Could not retrieve image and tag for event")
		return common.NewNotAvailableImageAndTag()
	}

	if len(events) == 0 {
		return common.NewNotAvailableImageAndTag()
	}

	triggeredData := &keptnv2.DeploymentTriggeredEventData{}
	err := keptnv2.Decode(events[0].Data, triggeredData)
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

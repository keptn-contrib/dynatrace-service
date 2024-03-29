package keptn

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/keptn/go-utils/pkg/api/models"
	v2 "github.com/keptn/go-utils/pkg/api/utils/v2"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

// EventClientInterface encapsulates functionality built on top of Keptn events.
type EventClientInterface interface {
	// IsPartOfRemediation checks whether the sequence includes a remediation triggered event or returns an error.
	IsPartOfRemediation(ctx context.Context, event adapter.EventContentAdapter) (bool, error)

	// FindProblemID finds the Problem ID that is associated with the specified Keptn event or returns an error.
	FindProblemID(ctx context.Context, keptnEvent adapter.EventContentAdapter) (string, error)

	// GetImageAndTag extracts the image and tag associated with a deployment triggered as part of the sequence.
	GetImageAndTag(ctx context.Context, keptnEvent adapter.EventContentAdapter) common.ImageAndTag

	// GetEventTimeStampForType tries to get the time stamp of a certain event as part of the sequence.
	GetEventTimeStampForType(ctx context.Context, keptnEvent adapter.EventContentAdapter, eventType string) (*time.Time, error)
}

// EventClient implements offers EventClientInterface using api.EventsV1Interface.
type EventClient struct {
	client v2.EventsInterface
}

// NewEventClient creates a new EventClient using the specified api.EventsV1Interface.
func NewEventClient(client v2.EventsInterface) *EventClient {
	return &EventClient{
		client: client,
	}
}

// IsPartOfRemediation checks whether the sequence includes a remediation triggered event or returns an error.
func (c *EventClient) IsPartOfRemediation(ctx context.Context, event adapter.EventContentAdapter) (bool, error) {
	events, err := c.client.GetEvents(ctx,
		&v2.EventFilter{
			Project:      event.GetProject(),
			Stage:        event.GetStage(),
			Service:      event.GetService(),
			EventType:    keptnv2.GetTriggeredEventType("remediation"),
			KeptnContext: event.GetShKeptnContext(),
		},
		v2.EventsGetEventsOptions{})

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
func (c *EventClient) FindProblemID(ctx context.Context, keptnEvent adapter.EventContentAdapter) (string, error) {
	// Step 1 - see if we have a Problem Url in the labels
	problemID := TryGetProblemIDFromLabels(keptnEvent)
	if problemID != "" {
		return problemID, nil
	}

	// Step 2 - lets see if we have a ProblemOpenEvent for this KeptnContext - if so - we try to extract the Problem ID
	events, mErr := c.client.GetEvents(ctx,
		&v2.EventFilter{
			Project:      keptnEvent.GetProject(),
			EventType:    keptncommon.ProblemOpenEventType,
			KeptnContext: keptnEvent.GetShKeptnContext(),
		},
		v2.EventsGetEventsOptions{})

	if mErr != nil {
		return "", fmt.Errorf("could not retrieve problem.open event for incoming event: %s", mErr.GetMessage())
	}

	if len(events) == 0 {
		return "", errors.New("could not retrieve problem.open event for incoming event: no events returned")
	}

	problemOpenEvent := &keptncommon.ProblemEventData{}
	err := keptnv2.Decode(events[0].Data, problemOpenEvent)
	if err != nil {
		return "", fmt.Errorf("could not decode problem.open event: %w", err)
	}

	if problemOpenEvent.PID == "" {
		return "", errors.New("no problem ID is included in problem.open event")
	}

	return problemOpenEvent.PID, nil
}

// GetImageAndTag extracts the image and tag associated with a deployment triggered as part of the sequence.
func (c *EventClient) GetImageAndTag(ctx context.Context, event adapter.EventContentAdapter) common.ImageAndTag {

	gotEvent, err := c.getEventDataForType(ctx, event, keptnv2.GetTriggeredEventType(keptnv2.DeploymentTaskName))
	if err != nil {
		log.WithError(err).Error("Could not retrieve image and tag for event")
		return common.NewNotAvailableImageAndTag()
	}

	triggeredData := &keptnv2.DeploymentTriggeredEventData{}
	err = keptnv2.Decode(gotEvent.Data, triggeredData)
	if err != nil {
		log.WithError(err).Error("Could not decode event data")
		return common.NewNotAvailableImageAndTag()
	}

	for key, value := range triggeredData.ConfigurationChange.Values {
		if strings.HasSuffix(key, "image") {
			return common.TryParseImageAndTag(value)
		}
	}

	return common.NewNotAvailableImageAndTag()
}

func (c *EventClient) getEventDataForType(ctx context.Context, event adapter.EventContentAdapter, eventType string) (*models.KeptnContextExtendedCE, error) {
	events, mErr := c.client.GetEvents(ctx,
		&v2.EventFilter{
			Project:      event.GetProject(),
			Stage:        event.GetStage(),
			Service:      event.GetService(),
			EventType:    eventType,
			KeptnContext: event.GetShKeptnContext(),
		},
		v2.EventsGetEventsOptions{})

	if mErr != nil {
		return nil, errors.New(mErr.GetMessage())
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("could not find any event for type '%s' in project '%s', stage '%s' and service '%s' for shkeptncontext '%s'", eventType, event.GetProject(), event.GetStage(), event.GetService(), event.GetShKeptnContext())
	}

	return events[0], nil
}

func (c *EventClient) GetEventTimeStampForType(ctx context.Context, event adapter.EventContentAdapter, eventType string) (*time.Time, error) {
	gotEvent, err := c.getEventDataForType(ctx, event, eventType)
	if err != nil {
		return nil, err
	}

	return &gotEvent.Time, nil
}

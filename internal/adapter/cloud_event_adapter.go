package adapter

import (
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/types"
	log "github.com/sirupsen/logrus"
)

const shKeptnContext = "shkeptncontext"

type TriggeredCloudEventContentAdapter interface {
	CloudEventContentAdapter

	GetEventID() string
}

type CloudEventContentAdapter interface {
	GetShKeptnContext() string
}

type CloudEventPayloadParseError struct {
	cause error
}

func (e *CloudEventPayloadParseError) Error() string {
	return fmt.Sprintf("could not parse cloud event payload: %v", e.cause)
}

type CloudEventAdapter struct {
	ce cloudevents.Event
}

func NewCloudEventAdapter(ce cloudevents.Event) CloudEventAdapter {
	return CloudEventAdapter{ce: ce}
}

func (a CloudEventAdapter) GetShKeptnContext() string {
	context, err := types.ToString(a.ce.Context.GetExtensions()[shKeptnContext])
	if err != nil {
		log.WithError(err).Debug("Event does not contain " + shKeptnContext)
	}
	return context
}

func (a CloudEventAdapter) GetSource() string {
	return a.ce.Source()
}

func (a CloudEventAdapter) GetEventID() string {
	return a.ce.ID()
}

func (a CloudEventAdapter) GetType() string {
	return a.ce.Type()
}

func (a CloudEventAdapter) GetTime() time.Time {
	return a.ce.Time()
}

// PayloadAs attempts to populate the provided content object with the event payload. Will return an error otherwise.
// content should be a pointer type.
func (a CloudEventAdapter) PayloadAs(content interface{}) error {
	err := a.ce.DataAs(content)
	if err != nil {
		return &CloudEventPayloadParseError{cause: err}
	}

	return nil
}

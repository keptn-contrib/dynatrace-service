package adapter

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/types"
	log "github.com/sirupsen/logrus"
)

const shKeptnContext = "shkeptncontext"

type CloudEventAdapter struct {
	ce cloudevents.Event
}

func NewCloudEventAdapter(ce cloudevents.Event) CloudEventAdapter {
	return CloudEventAdapter{ce: ce}
}

func (a CloudEventAdapter) Context() string {
	// TODO 2021-08-27: remove event/helper.go GetShKeptnContext() later on
	context, err := types.ToString(a.ce.Context.GetExtensions()[shKeptnContext])
	if err != nil {
		log.WithError(err).Debug("Event does not contain " + shKeptnContext)
	}
	return context
}

func (a CloudEventAdapter) Source() string {
	return a.ce.Source()
}

func (a CloudEventAdapter) ID() string {
	return a.ce.ID()
}

func (a CloudEventAdapter) Type() string {
	return a.ce.Type()
}

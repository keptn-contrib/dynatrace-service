package monitoring

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type ConfigureMonitoringFinishedEventFactory struct {
	eventData ConfigureMonitoringAdapterInterface
	status    keptnv2.StatusType
	result    keptnv2.ResultType
	message   string
}

func (f *ConfigureMonitoringFinishedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	return f.getEvent(f.status, f.result, f.message)
}

func NewSuccessEventFactory(message string) *ConfigureMonitoringFinishedEventFactory {
	return &ConfigureMonitoringFinishedEventFactory{
		status:  keptnv2.StatusSucceeded,
		result:  keptnv2.ResultPass,
		message: message,
	}
}

func NewFailureEventFactory(message string) *ConfigureMonitoringFinishedEventFactory {
	return &ConfigureMonitoringFinishedEventFactory{
		status:  keptnv2.StatusErrored,
		result:  keptnv2.ResultFailed,
		message: message,
	}
}

func (f *ConfigureMonitoringFinishedEventFactory) getEvent(status keptnv2.StatusType, result keptnv2.ResultType, message string) (*cloudevents.Event, error) {
	cmFinishedEvent := &keptnv2.ConfigureMonitoringFinishedEventData{
		EventData: keptnv2.EventData{
			Project: f.eventData.GetProject(),
			Service: f.eventData.GetService(),
			Status:  status,
			Result:  result,
			Message: message,
		},
	}

	ev := cloudevents.NewEvent()
	ev.SetSource(event.GetEventSource())
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetType(keptnv2.GetFinishedEventType(keptnv2.ConfigureMonitoringTaskName))
	ev.SetExtension("shkeptncontext", f.eventData.GetShKeptnContext())
	ev.SetExtension("triggeredid", f.eventData.GetEventID())

	err := ev.SetData(cloudevents.ApplicationJSON, cmFinishedEvent)
	if err != nil {
		return nil, fmt.Errorf("could not marshal cloud event payload: %v", err)
	}

	return &ev, nil
}

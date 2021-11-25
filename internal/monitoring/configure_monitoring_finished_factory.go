package monitoring

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type ConfigureMonitoringFinishedEventFactory struct {
	eventData ConfigureMonitoringAdapterInterface
	status    keptnv2.StatusType
	result    keptnv2.ResultType
	message   string
}

func (f *ConfigureMonitoringFinishedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	return f.getEventFactory(f.status, f.result, f.message).CreateCloudEvent()
}

func NewSucceededConfigureMonitoringFinishedEventFactory(eventData ConfigureMonitoringAdapterInterface, message string) *ConfigureMonitoringFinishedEventFactory {
	return &ConfigureMonitoringFinishedEventFactory{
		eventData: eventData,
		status:    keptnv2.StatusSucceeded,
		result:    keptnv2.ResultPass,
		message:   message,
	}
}

func NewErroredConfigureMonitoringFinishedEventFactory(eventData ConfigureMonitoringAdapterInterface, err error) *ConfigureMonitoringFinishedEventFactory {
	message := fmt.Sprintf("Configure monitoring failed. dynatrace-service: %s", err.Error())
	return &ConfigureMonitoringFinishedEventFactory{
		eventData: eventData,
		status:    keptnv2.StatusErrored,
		result:    keptnv2.ResultFailed,
		message:   message,
	}
}

func (f *ConfigureMonitoringFinishedEventFactory) getEventFactory(status keptnv2.StatusType, result keptnv2.ResultType, message string) adapter.CloudEventFactoryInterface {
	cmFinishedEvent := &keptnv2.ConfigureMonitoringFinishedEventData{
		EventData: keptnv2.EventData{
			Project: f.eventData.GetProject(),
			Service: f.eventData.GetService(),
			Status:  status,
			Result:  result,
			Message: message,
		},
	}

	return adapter.NewCloudEventFactory(
		f.eventData,
		keptnv2.GetFinishedEventType(keptnv2.ConfigureMonitoringTaskName),
		cmFinishedEvent)
}

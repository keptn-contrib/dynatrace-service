package sli

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type GetSliStartedEventFactory struct {
	event GetSLITriggeredAdapterInterface
}

func NewGetSliStartedEventFactory(event GetSLITriggeredAdapterInterface) *GetSliStartedEventFactory {
	return &GetSliStartedEventFactory{
		event: event,
	}
}

func (f *GetSliStartedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	getSLIStartedEvent := keptnv2.GetSLIStartedEventData{
		EventData: keptnv2.EventData{
			Project: f.event.GetProject(),
			Stage:   f.event.GetStage(),
			Service: f.event.GetService(),
			Labels:  f.event.GetLabels(),
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
		},
	}

	return adapter.NewCloudEventFactory(f.event, keptnv2.GetStartedEventType(keptnv2.GetSLITaskName), getSLIStartedEvent).CreateCloudEvent()

}

type GetSliFinishedEventFactory struct {
	event           GetSLITriggeredAdapterInterface
	indicatorValues []*keptnv2.SLIResult
}

func NewGetSLIFinishedEventFactory(event GetSLITriggeredAdapterInterface, indicatorValues []*keptnv2.SLIResult) *GetSliFinishedEventFactory {
	return &GetSliFinishedEventFactory{
		event:           event,
		indicatorValues: indicatorValues,
	}
}

func (f *GetSliFinishedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	getSLIFinishedEvent := keptnv2.GetSLIFinishedEventData{
		EventData: keptnv2.EventData{
			Project: f.event.GetProject(),
			Stage:   f.event.GetStage(),
			Service: f.event.GetService(),
			Labels:  f.event.GetLabels(),
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
		},
		GetSLI: keptnv2.GetSLIFinished{
			IndicatorValues: f.indicatorValues,
			Start:           f.event.GetSLIStart(),
			End:             f.event.GetSLIEnd(),
		},
	}

	return adapter.NewCloudEventFactory(f.event, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), getSLIFinishedEvent).CreateCloudEvent()
}

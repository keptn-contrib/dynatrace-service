package sli

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
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
	status          keptnv2.StatusType
	indicatorValues []result.SLIResult
	err             error
}

func NewSucceededGetSLIFinishedEventFactory(event GetSLITriggeredAdapterInterface, indicatorValues []result.SLIResult, err error) *GetSliFinishedEventFactory {
	return &GetSliFinishedEventFactory{
		event:           event,
		status:          keptnv2.StatusSucceeded,
		indicatorValues: indicatorValues,
		err:             err,
	}
}

func NewErroredGetSLIFinishedEventFactory(event GetSLITriggeredAdapterInterface, indicatorValues []result.SLIResult, err error) *GetSliFinishedEventFactory {
	return &GetSliFinishedEventFactory{
		event:           event,
		status:          keptnv2.StatusErrored,
		indicatorValues: indicatorValues,
		err:             err,
	}
}

func (f *GetSliFinishedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	sliResultSummarizer := result.NewSLIResultSummarizer(f.indicatorValues)
	result := sliResultSummarizer.Result()
	message := sliResultSummarizer.SummaryMessage()

	if f.err != nil {
		result = keptnv2.ResultFailed
		message = f.err.Error()
	}

	if f.status == keptnv2.StatusErrored {
		result = keptnv2.ResultFailed
	}

	getSLIFinishedEvent := keptnv2.GetSLIFinishedEventData{
		EventData: keptnv2.EventData{
			Project: f.event.GetProject(),
			Stage:   f.event.GetStage(),
			Service: f.event.GetService(),
			Labels:  f.event.GetLabels(),
			Status:  f.status,
			Result:  result,
			Message: message,
		},
		GetSLI: keptnv2.GetSLIFinished{
			IndicatorValues: getKeptnIndicatorValues(f.indicatorValues),
			Start:           f.event.GetSLIStart(),
			End:             f.event.GetSLIEnd(),
		},
	}

	return adapter.NewCloudEventFactory(f.event, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), getSLIFinishedEvent).CreateCloudEvent()
}

// getKeptnIndicatorValues unwraps the indicator values to Keptn SLIResults.
func getKeptnIndicatorValues(indicatorValues []result.SLIResult) []*keptnv2.SLIResult {
	var keptnIndicatorValues []*keptnv2.SLIResult
	for _, indicator := range indicatorValues {
		keptnSLIResult := indicator.KeptnSLIResult()
		keptnIndicatorValues = append(keptnIndicatorValues, &keptnSLIResult)
	}
	return keptnIndicatorValues
}

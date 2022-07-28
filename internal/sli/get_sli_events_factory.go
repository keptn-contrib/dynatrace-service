package sli

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// getSLIFinishedEventData is a keptnv2.GetSLIFinishedEventData using the getSLIFinished type defined here.
type getSLIFinishedEventData struct {
	keptnv2.EventData
	GetSLI getSLIFinished `json:"get-sli"`
}

// getSLIFinished is a keptnv2.GetSLIFinished using the sliResult type defined here.
type getSLIFinished struct {
	// Start defines the start timestamp
	Start string `json:"start"`
	// End defines the end timestamp
	End string `json:"end"`
	// IndicatorValues defines the fetched SLI values
	IndicatorValues []sliResult `json:"indicatorValues,omitempty"`
}

// sliResult is a simplified keptnv2.SLIResult with an additional query field.
type sliResult struct {
	Metric  string  `json:"metric"`
	Value   float64 `json:"value"`
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	Query   string  `json:"query,omitempty"`
}

// GetSLIStartedEventFactory is a factory for get-sli.started cloud events.
type GetSLIStartedEventFactory struct {
	event GetSLITriggeredAdapterInterface
}

// NewGetSLIStartedEventFactory creates a new GetSliStartedEventFactory.
func NewGetSLIStartedEventFactory(event GetSLITriggeredAdapterInterface) *GetSLIStartedEventFactory {
	return &GetSLIStartedEventFactory{
		event: event,
	}
}

// CreateCloudEvent creates a cloud event based on the factory or returns an error if this can't be done.
func (f *GetSLIStartedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
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

// GetSLIFinishedEventFactory is a factory for get-sli.finished cloud events.
type GetSLIFinishedEventFactory struct {
	event           GetSLITriggeredAdapterInterface
	status          keptnv2.StatusType
	indicatorValues []result.SLIResult
	err             error
}

// NewSucceededGetSLIFinishedEventFactory creates a new GetSliFinishedEventFactory with status succeeded.
func NewSucceededGetSLIFinishedEventFactory(event GetSLITriggeredAdapterInterface, indicatorValues []result.SLIResult, err error) *GetSLIFinishedEventFactory {
	return &GetSLIFinishedEventFactory{
		event:           event,
		status:          keptnv2.StatusSucceeded,
		indicatorValues: indicatorValues,
		err:             err,
	}
}

// NewErroredGetSLIFinishedEventFactory creates a new GetSliFinishedEventFactory with status errored.
func NewErroredGetSLIFinishedEventFactory(event GetSLITriggeredAdapterInterface, indicatorValues []result.SLIResult, err error) *GetSLIFinishedEventFactory {
	return &GetSLIFinishedEventFactory{
		event:           event,
		status:          keptnv2.StatusErrored,
		indicatorValues: indicatorValues,
		err:             err,
	}
}

// CreateCloudEvent creates a cloud event based on the factory or returns an error if this can't be done.
func (f *GetSLIFinishedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
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

	getSLIFinishedEvent := getSLIFinishedEventData{
		EventData: keptnv2.EventData{
			Project: f.event.GetProject(),
			Stage:   f.event.GetStage(),
			Service: f.event.GetService(),
			Labels:  f.event.GetLabels(),
			Status:  f.status,
			Result:  result,
			Message: message,
		},
		GetSLI: getSLIFinished{
			IndicatorValues: convertIndicatorValues(f.indicatorValues),
			Start:           f.event.GetSLIStart(),
			End:             f.event.GetSLIEnd(),
		},
	}

	return adapter.NewCloudEventFactory(f.event, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), getSLIFinishedEvent).CreateCloudEvent()
}

// convertIndicatorValues converts the indicator values to sliResults for serialization.
func convertIndicatorValues(indicatorValues []result.SLIResult) []sliResult {
	var convertedIndicatorValues []sliResult
	for _, indicator := range indicatorValues {
		convertedIndicatorValues = append(convertedIndicatorValues,
			sliResult{
				Metric:  indicator.Metric(),
				Value:   indicator.Value(),
				Success: indicator.Success(),
				Message: indicator.Message(),
				Query:   indicator.Query()})
	}
	return convertedIndicatorValues
}

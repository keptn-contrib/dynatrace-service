package sli

import (
	"errors"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/dashboard"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

func NewErroredGetSLIFinishedEventFactory(incomingEvent GetSLITriggeredAdapterInterface, err error) *GetSLIFinishedEventFactory {
	return &GetSLIFinishedEventFactory{
		incomingEvent: incomingEvent,
		eventData:     newGetSLIFinishedEventData(incomingEvent, keptnv2.StatusErrored, keptnv2.ResultFailed, err.Error(), nil),
	}
}

func NewSuccessfulGetSLIFinishedEventFactoryFromError(incomingEvent GetSLITriggeredAdapterInterface, err error) *GetSLIFinishedEventFactory {
	return &GetSLIFinishedEventFactory{
		incomingEvent: incomingEvent,
		eventData:     newGetSLIFinishedEventData(incomingEvent, keptnv2.StatusSucceeded, keptnv2.ResultFailed, err.Error(), makeSLIResultsForError(err, incomingEvent)),
	}
}

func NewSuccessfulGetSLIFinishedEventFactoryFromResults(incomingEvent GetSLITriggeredAdapterInterface, results []result.SLIWithSLO) *GetSLIFinishedEventFactory {
	resultSummarizer := result.NewSummarizer(results)

	return &GetSLIFinishedEventFactory{
		incomingEvent: incomingEvent,
		eventData:     newGetSLIFinishedEventData(incomingEvent, keptnv2.StatusSucceeded, resultSummarizer.OverallResult(), resultSummarizer.SummaryMessage(), convertResults(results)),
	}
}

func newGetSLIFinishedEventData(incomingEvent GetSLITriggeredAdapterInterface, status keptnv2.StatusType, result keptnv2.ResultType, message string, results []sliResult) *getSLIFinishedEventData {
	return &getSLIFinishedEventData{
		EventData: keptnv2.EventData{
			Project: incomingEvent.GetProject(),
			Stage:   incomingEvent.GetStage(),
			Service: incomingEvent.GetService(),
			Labels:  incomingEvent.GetLabels(),
			Status:  status,
			Result:  result,
			Message: message,
		},
		GetSLI: getSLIFinished{
			IndicatorValues: results,
			Start:           incomingEvent.GetSLIStart(),
			End:             incomingEvent.GetSLIEnd(),
		},
	}
}

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
	incomingEvent GetSLITriggeredAdapterInterface
	eventData     *getSLIFinishedEventData
}

// NewGetSLIFinishedEventFactory creates a new GetSliFinishedEventFactory.
func NewGetSLIFinishedEventFactory(incomingEvent GetSLITriggeredAdapterInterface, eventData *getSLIFinishedEventData) *GetSLIFinishedEventFactory {
	return &GetSLIFinishedEventFactory{
		incomingEvent: incomingEvent,
		eventData:     eventData,
	}
}

// CreateCloudEvent creates a cloud event based on the factory or returns an error if this can't be done.
func (f *GetSLIFinishedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	return adapter.NewCloudEventFactory(f.incomingEvent, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), f.eventData).CreateCloudEvent()
}

func makeSLIResultsForError(err error, eventData GetSLITriggeredAdapterInterface) []sliResult {
	indicators := eventData.GetIndicators()

	var errType *dashboard.DashboardError
	if len(indicators) == 0 || errors.As(err, &errType) {
		return []sliResult{makeFailedResult(NoMetricIndicator, err)}
	}

	sliResults := make([]sliResult, len(indicators))
	for i, indicatorName := range indicators {
		sliResults[i] = makeFailedResult(indicatorName, err)
	}

	return sliResults
}

func makeFailedResult(indicatorName string, err error) sliResult {
	return sliResult{Metric: indicatorName, Success: false, Message: err.Error()}
}

// convertResults converts the indicator values to sliResults for serialization.
func convertResults(results []result.SLIWithSLO) []sliResult {
	convertedIndicatorValues := make([]sliResult, len(results))
	for i, r := range results {
		sr := r.SLIResult()
		convertedIndicatorValues[i] =
			sliResult{
				Metric:  sr.Metric,
				Value:   sr.Value,
				Success: sr.Success,
				Message: sr.Message,
				Query:   sr.Query}
	}
	return convertedIndicatorValues
}

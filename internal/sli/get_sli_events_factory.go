package sli

import (
	"strings"

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
	status          keptnv2.StatusType
	indicatorValues []*keptnv2.SLIResult
	err             error
}

func NewSucceededGetSLIFinishedEventFactory(event GetSLITriggeredAdapterInterface, indicatorValues []*keptnv2.SLIResult, err error) *GetSliFinishedEventFactory {
	return &GetSliFinishedEventFactory{
		event:           event,
		status:          keptnv2.StatusSucceeded,
		indicatorValues: indicatorValues,
		err:             err,
	}
}

func NewErroredGetSLIFinishedEventFactory(event GetSLITriggeredAdapterInterface, indicatorValues []*keptnv2.SLIResult, err error) *GetSliFinishedEventFactory {
	return &GetSliFinishedEventFactory{
		event:           event,
		status:          keptnv2.StatusErrored,
		indicatorValues: indicatorValues,
		err:             err,
	}
}

func (f *GetSliFinishedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	result := keptnv2.ResultPass
	message := ""
	if f.err != nil {
		result = keptnv2.ResultFailed
		message = f.err.Error()
	}

	// get error messages if only some SLIs failed and there was no error
	sliErrorMessages := getErrorMessagesFromSLIResults(f.indicatorValues)
	if f.err == nil && len(sliErrorMessages) > 0 {
		result = keptnv2.ResultFailed
		message = strings.Join(sliErrorMessages, "; ")
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
			IndicatorValues: f.indicatorValues,
			Start:           f.event.GetSLIStart(),
			End:             f.event.GetSLIEnd(),
		},
	}

	return adapter.NewCloudEventFactory(f.event, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), getSLIFinishedEvent).CreateCloudEvent()
}

func getErrorMessagesFromSLIResults(indicatorValues []*keptnv2.SLIResult) []string {
	var errorMessages []string
	for _, indicator := range indicatorValues {
		if indicator.Success == false {
			errorMessages = append(errorMessages, indicator.Message)
		}
	}

	return errorMessages
}

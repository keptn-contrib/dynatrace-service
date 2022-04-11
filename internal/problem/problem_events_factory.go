package problem

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

type ProblemClosedEventFactory struct {
	event ProblemAdapterInterface
}

func NewProblemClosedEventFactory(event ProblemAdapterInterface) *ProblemClosedEventFactory {
	return &ProblemClosedEventFactory{
		event: event,
	}
}

func (f *ProblemClosedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	rawProblem := shallowCopyRawProblem(f.event.GetRawProblem())

	rawProblem["State"] = "CLOSED"
	rawProblem["project"] = f.event.GetProject()
	rawProblem["stage"] = f.event.GetStage()
	rawProblem["service"] = f.event.GetService()

	// create labels map if it is not already already there
	labels, ok := rawProblem["labels"].(map[string]interface{})
	if !ok {
		labels = make(map[string]interface{})
		rawProblem["labels"] = labels
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	labels[common.ProblemURLLabel] = f.event.GetProblemURL()

	return adapter.NewCloudEventFactoryBase(f.event, keptn.ProblemEventType, rawProblem).CreateCloudEvent()
}

func shallowCopyRawProblem(rawProblem RawProblem) RawProblem {
	rawProblemCopy := make(RawProblem, len(rawProblem))
	for key, value := range rawProblem {
		rawProblemCopy[key] = value
	}
	return rawProblemCopy
}

type RemediationTriggeredEventFactory struct {
	event ProblemAdapterInterface
}

func NewRemediationTriggeredEventFactory(event ProblemAdapterInterface) *RemediationTriggeredEventFactory {
	return &RemediationTriggeredEventFactory{
		event: event,
	}
}

func (f *RemediationTriggeredEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	remediationEventData := RemediationTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: f.event.GetProject(),
			Stage:   f.event.GetStage(),
			Service: f.event.GetService(),
		},
		Problem: f.event.GetRawProblem(),
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	remediationEventData.Labels = make(map[string]string)
	remediationEventData.Labels[common.ProblemURLLabel] = f.event.GetProblemURL()

	eventType := keptnv2.GetTriggeredEventType(f.event.GetStage() + "." + remediationTaskName)

	return adapter.NewCloudEventFactoryBase(f.event, eventType, remediationEventData).CreateCloudEvent()
}

package problem

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
)

const problemURLLabel = "Problem URL"

type ProblemClosedEventFactory struct {
	event ProblemAdapterInterface
}

func NewProblemClosedEventFactory(event ProblemAdapterInterface) *ProblemClosedEventFactory {
	return &ProblemClosedEventFactory{
		event: event,
	}
}

func (f *ProblemClosedEventFactory) CreateCloudEvent() (*cloudevents.Event, error) {
	problemData := keptn.ProblemEventData{
		State:          "CLOSED",
		PID:            f.event.GetPID(),
		ProblemID:      f.event.GetProblemID(),
		ProblemTitle:   f.event.GetProblemTitle(),
		ProblemDetails: f.event.GetProblemDetails(),
		ProblemURL:     f.event.GetProblemURL(),
		ImpactedEntity: f.event.GetImpactedEntity(),
		Tags:           f.event.GetProblemTags(),
		Project:        f.event.GetProject(),
		Stage:          f.event.GetStage(),
		Service:        f.event.GetService(),
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	problemData.Labels = make(map[string]string)
	problemData.Labels[problemURLLabel] = f.event.GetProblemURL()

	return adapter.NewCloudEventFactoryBase(f.event, keptn.ProblemEventType, problemData).CreateCloudEvent()

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
		Problem: ProblemDetails{
			State:              "OPEN",
			PID:                f.event.GetPID(),
			ProblemID:          f.event.GetProblemID(),
			ProblemTitle:       f.event.GetProblemTitle(),
			ProblemDetails:     f.event.GetProblemDetails(),
			ProblemDetailsHTML: f.event.GetProblemDetailsHTML(),
			ProblemDetailsText: f.event.GetProblemDetailsText(),
			ProblemImpact:      f.event.GetProblemImpact(),
			ProblemSeverity:    f.event.GetProblemSeverity(),
			ProblemURL:         f.event.GetProblemURL(),
			ImpactedEntity:     f.event.GetImpactedEntity(),
			Tags:               f.event.GetProblemTags(),
		},
	}

	// https://github.com/keptn-contrib/dynatrace-service/issues/176
	// add problem URL as label so it becomes clickable
	remediationEventData.Labels = make(map[string]string)
	remediationEventData.Labels[problemURLLabel] = f.event.GetProblemURL()

	eventType := keptnv2.GetTriggeredEventType(f.event.GetStage() + "." + remediationTaskName)

	return adapter.NewCloudEventFactoryBase(f.event, eventType, remediationEventData).CreateCloudEvent()
}

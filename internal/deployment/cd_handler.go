package deployment

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/config"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type CDEventHandler struct {
	event          cloudevents.Event
	dtConfigGetter config.DynatraceConfigGetterInterface
}

func NewCDEventHandler(event cloudevents.Event, configGetter config.DynatraceConfigGetterInterface) CDEventHandler {
	return CDEventHandler{
		event:          event,
		dtConfigGetter: configGetter,
	}
}

func (eh CDEventHandler) HandleEvent() error {
	switch eh.event.Type() {
	case keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName):
		return eh.handleDeploymentFinishedEvent()
	case keptnv2.GetTriggeredEventType(keptnv2.TestTaskName):
		return eh.handleTestTriggeredEvent()
	case keptnv2.GetFinishedEventType(keptnv2.TestTaskName):
		return eh.handleTestFinishedEvent()
	case keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName):
		return eh.handleEvaluationFinishedEvent()
	case keptnv2.GetTriggeredEventType(keptnv2.ReleaseTaskName):
		return eh.handleReleaseTriggeredEvent()
	case keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName):
	// continue
	default:
		log.WithField("EventType", eh.event.Type()).Info("Ignoring event")
	}

	return nil
}

func (eh *CDEventHandler) handleDeploymentFinishedEvent() error {
	keptnEvent, err := NewDeploymentFinishedAdapterFromEvent(eh.event)
	if err != nil {
		return err
	}

	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace config")
		return err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		log.WithError(err).Error("failed to load Dynatrace credentials")
		return err
	}

	// send Deployment Event
	de := event.CreateDeploymentEvent(keptnEvent, dynatraceConfig)

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewEventsClient(dtHelper).SendEvent(de)

	return nil
}

func (eh *CDEventHandler) handleTestTriggeredEvent() error {
	keptnEvent, err := NewTestTriggeredAdapterFromEvent(eh.event)
	if err != nil {
		return err
	}

	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("failed to load Dynatrace config")
		return err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		log.WithError(err).Error("failed to load Dynatrace credentials")
		return err
	}

	// Send Annotation Event
	ie := event.CreateAnnotationEvent(keptnEvent, dynatraceConfig)
	if ie.AnnotationType == "" {
		ie.AnnotationType = "Start Tests: " + keptnEvent.GetTestStrategy()
	}
	if ie.AnnotationDescription == "" {
		ie.AnnotationDescription = "Start running tests: " + keptnEvent.GetTestStrategy() + " against " + keptnEvent.GetService()
	}

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewEventsClient(dtHelper).SendEvent(ie)

	return nil
}

func (eh *CDEventHandler) handleTestFinishedEvent() error {
	keptnEvent, err := NewTestFinishedAdapterFromEvent(eh.event)
	if err != nil {
		return err
	}

	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace config")
		return err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace credentials")
		return err
	}

	// Send Annotation Event
	ae := event.CreateAnnotationEvent(keptnEvent, dynatraceConfig)
	if ae.AnnotationType == "" {
		ae.AnnotationType = "Stop Tests"
	}
	if ae.AnnotationDescription == "" {
		ae.AnnotationDescription = "Stop running tests: against " + keptnEvent.GetService()
	}

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewEventsClient(dtHelper).SendEvent(ae)

	return nil
}

func (eh *CDEventHandler) handleEvaluationFinishedEvent() error {
	keptnEvent, err := NewEvaluationFinishedAdapterFromEvent(eh.event)
	if err != nil {
		return err
	}

	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace config")
		return err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace credentials")
		return err
	}
	dtHelper := dynatrace.NewClient(creds)

	// Send Info Event
	ie := event.CreateInfoEvent(keptnEvent, dynatraceConfig)
	qualityGateDescription := fmt.Sprintf("Quality Gate Result in stage %s: %s (%.2f/100)", keptnEvent.GetStage(), keptnEvent.GetResult(), keptnEvent.GetEvaluationScore())
	ie.Title = fmt.Sprintf("Evaluation result: %s", keptnEvent.GetResult())

	if keptnEvent.IsPartOfRemediation() {
		if keptnEvent.GetResult() == keptnv2.ResultPass || keptnEvent.GetResult() == keptnv2.ResultWarning {
			ie.Title = "Remediation action successful"
		} else {
			ie.Title = "Remediation action not successful"
		}
		// If evaluation was done in context of a problem remediation workflow then post comments to the Dynatrace Problem
		pid, err := common.FindProblemIDForEvent(keptnEvent)
		if err == nil && pid != "" {
			// Comment we push over
			comment := fmt.Sprintf("[Keptn remediation evaluation](%s) resulted in %s (%.2f/100)", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], keptnEvent.GetResult(), keptnEvent.GetEvaluationScore())

			// this is posting the Event on the problem as a comment
			dynatrace.NewProblemsClient(dtHelper).AddProblemComment(pid, comment)
		}
	}
	ie.Description = qualityGateDescription

	dynatrace.NewEventsClient(dtHelper).SendEvent(ie)

	return nil
}

func (eh *CDEventHandler) handleReleaseTriggeredEvent() error {
	keptnEvent, err := NewReleaseTriggeredAdapterFromEvent(eh.event)
	if err != nil {
		return err
	}

	strategy, err := keptnevents.GetDeploymentStrategy(keptnEvent.GetDeploymentStrategy())
	if err != nil {
		log.WithError(err).Error("Could not determine deployment strategy")
		return err
	}
	dynatraceConfig, err := eh.dtConfigGetter.GetDynatraceConfig(keptnEvent)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace config")
		return err
	}
	creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
	if err != nil {
		log.WithError(err).Error("Failed to load Dynatrace credentials")
		return err
	}

	ie := event.CreateInfoEvent(keptnEvent, dynatraceConfig)
	if strategy == keptnevents.Direct && keptnEvent.GetResult() == keptnv2.ResultPass || keptnEvent.GetResult() == keptnv2.ResultWarning {
		title := fmt.Sprintf("PROMOTING from %s to next stage", keptnEvent.GetStage())
		ie.Title = title
		ie.Description = title
	} else if keptnEvent.GetResult() == keptnv2.ResultFailed {
		if strategy == keptnevents.Duplicate {
			title := "Rollback Artifact (Switch Blue/Green) in " + keptnEvent.GetStage()
			ie.Title = title
			ie.Description = title
		} else {
			title := fmt.Sprintf("NOT PROMOTING from %s to next stage", keptnEvent.GetStage())
			ie.Title = title
			ie.Description = title
		}
	}

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewEventsClient(dtHelper).SendEvent(ie)

	return nil
}

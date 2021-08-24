package deployment

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
)

type CDEventHandler struct {
	event          cloudevents.Event
	dtConfigGetter adapter.DynatraceConfigGetterInterface
}

func NewCDEventHandler(event cloudevents.Event, configGetter adapter.DynatraceConfigGetterInterface) CDEventHandler {
	return CDEventHandler{
		event:          event,
		dtConfigGetter: configGetter,
	}
}

func (eh CDEventHandler) HandleEvent() error {
	keptnHandler, err := keptnv2.NewKeptn(&eh.event, keptncommon.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not create Keptn handler")
	}

	switch eh.event.Type() {
	case keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName):
		return eh.handleDeploymentFinishedEvent(keptnHandler)
	case keptnv2.GetTriggeredEventType(keptnv2.TestTaskName):
		return eh.handleTestTriggeredEvent(keptnHandler)
	case keptnv2.GetFinishedEventType(keptnv2.TestTaskName):
		return eh.handleTestFinishedEvent(keptnHandler)
	case keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName):
		return eh.handleEvaluationFinishedEvent(keptnHandler)
	case keptnv2.GetTriggeredEventType(keptnv2.ReleaseTaskName):
		return eh.handleReleaseTriggeredEvent(keptnHandler)
	case keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName):
	// continue
	default:
		log.WithField("EventType", eh.event.Type()).Info("Ignoring event")
	}

	return nil
}

func (eh *CDEventHandler) handleDeploymentFinishedEvent(keptnHandler *keptnv2.Keptn) error {
	dfData := &keptnv2.DeploymentFinishedEventData{}
	err := eh.event.DataAs(dfData)
	if err != nil {
		log.WithError(err).Error("Could not parse event payload")
		return err
	}

	// initialize our objects
	keptnEvent := adapter.NewDeploymentFinishedAdapter(*dfData, keptnHandler.KeptnContext, eh.event.Source())

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

func (eh *CDEventHandler) handleTestTriggeredEvent(keptnHandler *keptnv2.Keptn) error {
	ttData := &keptnv2.TestTriggeredEventData{}
	err := eh.event.DataAs(ttData)
	if err != nil {
		log.WithError(err).Error("Could not parse event payload")
		return err
	}

	// initialize our objects
	keptnEvent := adapter.NewTestTriggeredAdapter(*ttData, keptnHandler.KeptnContext, eh.event.Source())

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
		ie.AnnotationType = "Start Tests: " + ttData.Test.TestStrategy
	}
	if ie.AnnotationDescription == "" {
		ie.AnnotationDescription = "Start running tests: " + ttData.Test.TestStrategy + " against " + ttData.Service
	}

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewEventsClient(dtHelper).SendEvent(ie)

	return nil
}

func (eh *CDEventHandler) handleTestFinishedEvent(keptnHandler *keptnv2.Keptn) error {
	tfData := &keptnv2.TestFinishedEventData{}
	err := eh.event.DataAs(tfData)
	if err != nil {
		log.WithError(err).Error("Could not parse event payload")
		return err
	}

	// initialize our objects
	keptnEvent := adapter.NewTestFinishedAdapter(*tfData, keptnHandler.KeptnContext, eh.event.Source())

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
		ae.AnnotationDescription = "Stop running tests: against " + tfData.Service
	}

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewEventsClient(dtHelper).SendEvent(ae)

	return nil
}

func (eh *CDEventHandler) handleEvaluationFinishedEvent(keptnHandler *keptnv2.Keptn) error {
	edData := &keptnv2.EvaluationFinishedEventData{}
	err := eh.event.DataAs(edData)
	if err != nil {
		log.WithError(err).Error("Error while parsing JSON payload")
		return err
	}
	// initialize our objects
	keptnEvent := adapter.NewEvaluationDoneAdapter(*edData, keptnHandler.KeptnContext, eh.event.Source())

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
	qualityGateDescription := fmt.Sprintf("Quality Gate Result in stage %s: %s (%.2f/100)", edData.Stage, edData.Result, edData.Evaluation.Score)
	ie.Title = fmt.Sprintf("Evaluation result: %s", edData.Result)

	if keptnEvent.IsPartOfRemediation() {
		if edData.Result == keptnv2.ResultPass || edData.Result == keptnv2.ResultWarning {
			ie.Title = "Remediation action successful"
		} else {
			ie.Title = "Remediation action not successful"
		}
		// If evaluation was done in context of a problem remediation workflow then post comments to the Dynatrace Problem
		pid, err := common.FindProblemIDForEvent(keptnHandler, keptnEvent.GetLabels())
		if err == nil && pid != "" {
			// Comment we push over
			comment := fmt.Sprintf("[Keptn remediation evaluation](%s) resulted in %s (%.2f/100)", keptnEvent.GetLabels()[common.KEPTNSBRIDGE_LABEL], edData.Result, edData.Evaluation.Score)

			// this is posting the Event on the problem as a comment
			dynatrace.NewProblemsClient(dtHelper).AddProblemComment(pid, comment)
		}
	}
	ie.Description = qualityGateDescription

	dynatrace.NewEventsClient(dtHelper).SendEvent(ie)

	return nil
}

func (eh *CDEventHandler) handleReleaseTriggeredEvent(keptnHandler *keptnv2.Keptn) error {
	rtData := &keptnv2.ReleaseTriggeredEventData{}
	err := eh.event.DataAs(rtData)
	if err != nil {
		log.WithError(err).Error("Error while parsing JSON payload")
		return err
	}
	keptnEvent := adapter.NewReleaseTriggeredAdapter(*rtData, keptnHandler.KeptnContext, eh.event.Source())

	strategy, err := keptnevents.GetDeploymentStrategy(rtData.Deployment.DeploymentStrategy)
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
	if strategy == keptnevents.Direct && rtData.Result == keptnv2.ResultPass || rtData.Result == keptnv2.ResultWarning {
		title := fmt.Sprintf("PROMOTING from %s to next stage", rtData.Stage)
		ie.Title = title
		ie.Description = title
	} else if rtData.Result == keptnv2.ResultFailed {
		if strategy == keptnevents.Duplicate {
			title := "Rollback Artifact (Switch Blue/Green) in " + rtData.Stage
			ie.Title = title
			ie.Description = title
		} else {
			title := fmt.Sprintf("NOT PROMOTING from %s to next stage", rtData.Stage)
			ie.Title = title
			ie.Description = title
		}
	}

	dtHelper := dynatrace.NewClient(creds)
	dynatrace.NewEventsClient(dtHelper).SendEvent(ie)

	return nil
}

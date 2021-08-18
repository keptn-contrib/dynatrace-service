package event_handler

import (
	"fmt"

	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/event"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type CDEventHandler struct {
	Event          cloudevents.Event
	dtConfigGetter adapter.DynatraceConfigGetterInterface
}

func (eh CDEventHandler) HandleEvent() error {
	shkeptncontext := event.GetShKeptnContext(eh.Event)

	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptncommon.KeptnOpts{})
	if err != nil {
		log.WithError(err).Error("Could not create Keptn handler")
	}
	if eh.Event.Type() == keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName) {
		dfData := &keptnv2.DeploymentFinishedEventData{}
		err := eh.Event.DataAs(dfData)
		if err != nil {
			log.WithError(err).Error("Could not parse event payload")
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewDeploymentFinishedAdapter(*dfData, shkeptncontext, eh.Event.Source())

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
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds)

		// send Deployment Event
		de := event.CreateDeploymentEvent(keptnEvent, dynatraceConfig)
		dtHelper.SendEvent(de)
	} else if eh.Event.Type() == keptnv2.GetTriggeredEventType(keptnv2.TestTaskName) {
		ttData := &keptnv2.TestTriggeredEventData{}
		err := eh.Event.DataAs(ttData)
		if err != nil {
			log.WithError(err).Error("Could not parse event payload")
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewTestTriggeredAdapter(*ttData, shkeptncontext, eh.Event.Source())

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
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds)

		// Send Annotation Event
		ie := event.CreateAnnotationEvent(keptnEvent, dynatraceConfig)
		if ie.AnnotationType == "" {
			ie.AnnotationType = "Start Tests: " + ttData.Test.TestStrategy
		}
		if ie.AnnotationDescription == "" {
			ie.AnnotationDescription = "Start running tests: " + ttData.Test.TestStrategy + " against " + ttData.Service
		}
		dtHelper.SendEvent(ie)
	} else if eh.Event.Type() == keptnv2.GetFinishedEventType(keptnv2.TestTaskName) {
		tfData := &keptnv2.TestFinishedEventData{}
		err := eh.Event.DataAs(tfData)
		if err != nil {
			log.WithError(err).Error("Could not parse event payload")
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewTestFinishedAdapter(*tfData, shkeptncontext, eh.Event.Source())

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
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds)

		// Send Annotation Event
		ie := event.CreateAnnotationEvent(keptnEvent, dynatraceConfig)

		if ie.AnnotationType == "" {
			ie.AnnotationType = "Stop Tests"
		}
		if ie.AnnotationDescription == "" {
			ie.AnnotationDescription = "Stop running tests: against " + tfData.Service
		}
		dtHelper.SendEvent(ie)
	} else if eh.Event.Type() == keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName) {
		edData := &keptnv2.EvaluationFinishedEventData{}
		err := eh.Event.DataAs(edData)
		if err != nil {
			log.WithError(err).Error("Error while parsing JSON payload")
			return err
		}
		// initialize our objects
		keptnEvent := adapter.NewEvaluationDoneAdapter(*edData, shkeptncontext, eh.Event.Source())

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
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds)

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
				err = dtHelper.SendProblemComment(pid, comment)
			}
		}
		ie.Description = qualityGateDescription
		dtHelper.SendEvent(ie)
	} else if eh.Event.Type() == keptnv2.GetTriggeredEventType(keptnv2.ReleaseTaskName) {
		rtData := &keptnv2.ReleaseTriggeredEventData{}
		err := eh.Event.DataAs(rtData)
		if err != nil {
			log.WithError(err).Error("Error while parsing JSON payload")
			return err
		}
		keptnEvent := adapter.NewReleaseTriggeredAdapter(*rtData, shkeptncontext, eh.Event.Source())

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
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds)

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
		dtHelper.SendEvent(ie)
	} else if eh.Event.Type() == keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName) {

	} else {
		log.WithField("EventType", eh.Event.Type()).Info("Ignoring event")
	}
	return nil
}

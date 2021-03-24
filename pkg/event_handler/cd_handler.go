package event_handler

import (
	"fmt"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
)

type CDEventHandler struct {
	Logger *keptncommon.Logger
	Event  cloudevents.Event
}

func (eh CDEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptncommon.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}
	if eh.Event.Type() == keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName) {
		dfData := &keptnv2.DeploymentFinishedEventData{}
		err := eh.Event.DataAs(dfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewDeploymentFinishedAdapter(*dfData, shkeptncontext, eh.Event.Source())

		dynatraceConfig, err := adapter.GetDynatraceConfig(keptnEvent, eh.Logger)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace config: " + err.Error())
			return err
		}
		creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace credentials: " + err.Error())
			return err
		}
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

		// send Deployment Event
		de := createDeploymentEvent(keptnEvent, dynatraceConfig, eh.Logger)
		dtHelper.SendEvent(de)
	} else if eh.Event.Type() == keptnv2.GetTriggeredEventType(keptnv2.TestTaskName) {
		ttData := &keptnv2.TestTriggeredEventData{}
		err := eh.Event.DataAs(ttData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewTestTriggeredAdapter(*ttData, shkeptncontext, eh.Event.Source())

		dynatraceConfig, err := adapter.GetDynatraceConfig(keptnEvent, eh.Logger)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace config: " + err.Error())
			return err
		}
		creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace credentials: " + err.Error())
			return err
		}
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

		// Send Annotation Event
		ie := createAnnotationEvent(keptnEvent, dynatraceConfig, eh.Logger)
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
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewTestFinishedAdapter(*tfData, shkeptncontext, eh.Event.Source())

		dynatraceConfig, err := adapter.GetDynatraceConfig(keptnEvent, eh.Logger)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace config: " + err.Error())
			return err
		}
		creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace credentials: " + err.Error())
			return err
		}
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

		// Send Annotation Event
		ie := createAnnotationEvent(keptnEvent, dynatraceConfig, eh.Logger)

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
			fmt.Println("Error while parsing JSON payload: " + err.Error())
			return err
		}
		// initialize our objects
		keptnEvent := adapter.NewEvaluationDoneAdapter(*edData, shkeptncontext, eh.Event.Source())

		dynatraceConfig, err := adapter.GetDynatraceConfig(keptnEvent, eh.Logger)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace config: " + err.Error())
			return err
		}
		creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace credentials: " + err.Error())
			return err
		}
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

		// Send Info Event
		ie := createInfoEvent(keptnEvent, dynatraceConfig, eh.Logger)
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
			fmt.Println("Error while parsing JSON payload: " + err.Error())
			return err
		}
		keptnEvent := adapter.NewReleaseTriggeredAdapter(*rtData, shkeptncontext, eh.Event.Source())

		strategy, err := keptnevents.GetDeploymentStrategy(rtData.Deployment.DeploymentStrategy)
		if err != nil {
			eh.Logger.Error(fmt.Sprintf("Could not determine deployment strategy: %s", err.Error()))
			return err
		}
		dynatraceConfig, err := adapter.GetDynatraceConfig(keptnEvent, eh.Logger)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace config: " + err.Error())
			return err
		}
		creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
		if err != nil {
			eh.Logger.Error("failed to load Dynatrace credentials: " + err.Error())
			return err
		}
		dtHelper := lib.NewDynatraceHelper(keptnHandler, creds, eh.Logger)

		ie := createInfoEvent(keptnEvent, dynatraceConfig, eh.Logger)
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
		eh.Logger.Info(fmt.Sprintf("Ignoring event of type %s", eh.Event.Type()))
	}
	return nil
}

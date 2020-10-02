package event_handler

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type CDEventHandler struct {
	Logger *keptn.Logger
	Event  cloudevents.Event
}

func (eh CDEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}

	if eh.Event.Type() == keptn.DeploymentFinishedEventType {
		dfData := &keptn.DeploymentFinishedEventData{}
		err := eh.Event.DataAs(dfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewDeploymentFinishedAdapter(*dfData, shkeptncontext, eh.Event.Source())

		dynatraceConfig, err := config.GetDynatraceConfig(keptnEvent, eh.Logger)
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

		// send Deployment EVent
		de := createDeploymentEvent(keptnEvent, dynatraceConfig, eh.Logger)
		dtHelper.SendEvent(de)

		// TODO: an additional channel (e.g. start-tests) to correctly determine the time when the tests actually start
		// ie := createInfoEvent(keptnEvent, eh.logger)
		ie := createAnnotationEvent(keptnEvent, dynatraceConfig, eh.Logger)
		if dfData.TestStrategy != "" {
			if ie.AnnotationType == "" {
				ie.AnnotationType = "Start Tests: " + dfData.TestStrategy
			}
			if ie.AnnotationDescription == "" {
				ie.AnnotationDescription = "Start running tests: " + dfData.TestStrategy + " against " + dfData.Service
			}
			dtHelper.SendEvent(ie)
		}
	} else if eh.Event.Type() == keptn.TestsFinishedEventType {
		tfData := &keptn.TestsFinishedEventData{}
		err := eh.Event.DataAs(tfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewTestFinishedAdapter(*tfData, shkeptncontext, eh.Event.Source())

		dynatraceConfig, err := config.GetDynatraceConfig(keptnEvent, eh.Logger)
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
		// ie := createInfoEvent(keptnEvent, eh.logger)
		ie := createAnnotationEvent(keptnEvent, dynatraceConfig, eh.Logger)
		if tfData.TestStrategy != "" {
			if ie.AnnotationType == "" {
				ie.AnnotationType = "Stop Tests: " + tfData.TestStrategy
			}
			if ie.AnnotationDescription == "" {
				ie.AnnotationDescription = "Stop running tests: " + tfData.TestStrategy + " against " + tfData.Service
			}
			dtHelper.SendEvent(ie)
		}
	} else if eh.Event.Type() == keptn.EvaluationDoneEventType {
		edData := &keptn.EvaluationDoneEventData{}
		err := eh.Event.DataAs(edData)
		if err != nil {
			fmt.Println("Error while parsing JSON payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent := adapter.NewEvaluationDoneAdapter(*edData, shkeptncontext, eh.Event.Source())

		dynatraceConfig, err := config.GetDynatraceConfig(keptnEvent, eh.Logger)
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
		// If DeploymentStrategy == "" it means we are doing Quality-Gates Only!
		if edData.DeploymentStrategy == "" {
			ie.Title = fmt.Sprintf("Quality Gate Result: %s (%.2f/100)", edData.Result, edData.EvaluationDetails.Score)
		} else if edData.Result == "pass" || edData.Result == "warning" {
			if edData.TestStrategy == "real-user" {
				ie.Title = "Remediation action successful"
			} else {
				ie.Title = fmt.Sprintf("Quality Gate Result: %s (%.2f/100): PROMOTING from %s to next stage", edData.Result, edData.EvaluationDetails.Score, edData.Stage)
				// ie.Title = "Promote Artifact from " + edData.Stage + " to next stage"
			}

		} else if edData.Result == "fail" && edData.DeploymentStrategy == "blue_green_service" {
			if edData.TestStrategy == "real-user" {
				ie.Title = "Remediation action not successful"
			} else {
				ie.Title = "Rollback Artifact (Switch Blue/Green) in " + edData.Stage
			}
		} else if edData.Result == "fail" && edData.DeploymentStrategy == "direct" {
			if edData.TestStrategy == "real-user" {
				ie.Title = "Remediation action not successful"
			} else {
				ie.Title = fmt.Sprintf("Quality Gate Result: %s (%.2f/100): NOT PROMOTING artifact from %s", edData.Result, edData.EvaluationDetails.Score, edData.Stage)
				// ie.Title = "NOT PROMOTING Artifact from " + edData.Stage + " due to failed evaluation"
			}
		} else {
			eh.Logger.Error("No valid deployment strategy defined in keptn event.")
			return nil
		}
		ie.Description = "Keptn evaluation status: " + edData.Result
		dtHelper.SendEvent(ie)
	} else {
		eh.Logger.Info(fmt.Sprintf("Ignoring event of type %s", eh.Event.Type()))
	}
	return nil
}

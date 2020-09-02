package event_handler

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type CDEventHandler struct {
	Logger *keptn.Logger
	Event  cloudevents.Event
}

/**
 * Initializes baseKeptnEvent and returns it + dynatraceConfig
 */
func (eh CDEventHandler) initObjectsForCDEventHandler(project, stage, service, testStrategy, image, tag string, labels map[string]string, context string) (*baseKeptnEvent, *DynatraceConfigFile, string) {
	keptnEvent := &baseKeptnEvent{}
	keptnEvent.project = project
	keptnEvent.stage = stage
	keptnEvent.service = service
	keptnEvent.testStrategy = testStrategy
	keptnEvent.image = image
	keptnEvent.tag = tag
	keptnEvent.labels = labels
	keptnEvent.context = context
	dynatraceConfig, _ := getDynatraceConfig(keptnEvent, eh.Logger)
	keptnBridgeURL, err := common.GetKeptnBridgeURL()
	if keptnEvent.labels == nil {
		keptnEvent.labels = make(map[string]string)
	}
	if err == nil {
		keptnEvent.labels["Keptns Bridge"] = keptnBridgeURL + "/trace/" + context
	}

	dtCreds := ""
	if dynatraceConfig != nil {
		dtCreds = dynatraceConfig.DtCreds
	}

	return keptnEvent, dynatraceConfig, dtCreds
}

func (eh CDEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	clientSet, err := common.GetKubernetesClient()
	if err != nil {
		eh.Logger.Error("could not create k8s client")
		return err
	}

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}

	dtHelper, err := lib.NewDynatraceHelper(keptnHandler)
	if err != nil {
		eh.Logger.Error("Could not create Dynatrace Helper: " + err.Error())
		return err
	}
	dtHelper.KubeApi = clientSet
	dtHelper.Logger = eh.Logger

	eh.Logger.Info("Check if event of type " + eh.Event.Type() + " should be sent to Dynatrace.")

	if eh.Event.Type() == keptn.DeploymentFinishedEventType {
		dfData := &keptn.DeploymentFinishedEventData{}
		err := eh.Event.DataAs(dfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent, dynatraceConfig, dtCreds := eh.initObjectsForCDEventHandler(dfData.Project, dfData.Stage, dfData.Service, dfData.TestStrategy, dfData.Image, dfData.Tag, dfData.Labels, shkeptncontext)
		if dfData.DeploymentURILocal != "" {
			keptnEvent.labels["deploymentURILocal"] = dfData.DeploymentURILocal
		}
		if dfData.DeploymentURIPublic != "" {
			keptnEvent.labels["deploymentURIPublic"] = dfData.DeploymentURIPublic
		}

		// send Deployment EVent
		de := CreateDeploymentEvent(keptnEvent, dynatraceConfig, eh.Logger)
		dtHelper.SendEvent(de, dtCreds)

		// TODO: an additional channel (e.g. start-tests) to correctly determine the time when the tests actually start
		// ie := createInfoEvent(keptnEvent, eh.Logger)
		ie := CreateAnnotationEvent(keptnEvent, dynatraceConfig, eh.Logger)
		if dfData.TestStrategy != "" {
			if ie.AnnotationType == "" {
				ie.AnnotationType = "Start Tests: " + dfData.TestStrategy
			}
			if ie.AnnotationDescription == "" {
				ie.AnnotationDescription = "Start running tests: " + dfData.TestStrategy + " against " + dfData.Service
			}
			dtHelper.SendEvent(ie, dtCreds)
		}
	} else if eh.Event.Type() == keptn.TestsFinishedEventType {
		tfData := &keptn.TestsFinishedEventData{}
		err := eh.Event.DataAs(tfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent, dynatraceConfig, dtCreds := eh.initObjectsForCDEventHandler(tfData.Project, tfData.Stage, tfData.Service, tfData.TestStrategy, "", "", tfData.Labels, shkeptncontext)

		// Send Annotation Event
		// ie := createInfoEvent(keptnEvent, eh.Logger)
		ie := CreateAnnotationEvent(keptnEvent, dynatraceConfig, eh.Logger)
		if tfData.TestStrategy != "" {
			if ie.AnnotationType == "" {
				ie.AnnotationType = "Stop Tests: " + tfData.TestStrategy
			}
			if ie.AnnotationDescription == "" {
				ie.AnnotationDescription = "Stop running tests: " + tfData.TestStrategy + " against " + tfData.Service
			}
			dtHelper.SendEvent(ie, dtCreds)
		}
	} else if eh.Event.Type() == keptn.EvaluationDoneEventType {
		edData := &keptn.EvaluationDoneEventData{}
		err := eh.Event.DataAs(edData)
		if err != nil {
			fmt.Println("Error while parsing JSON payload: " + err.Error())
			return err
		}

		// initialize our objects
		keptnEvent, dynatraceConfig, dtCreds := eh.initObjectsForCDEventHandler(edData.Project, edData.Stage, edData.Service, edData.TestStrategy, "", "", edData.Labels, shkeptncontext)
		keptnEvent.labels["Quality Gate Score"] = fmt.Sprintf("%.2f", edData.EvaluationDetails.Score)
		keptnEvent.labels["No of evaluated SLIs"] = fmt.Sprintf("%d", len(edData.EvaluationDetails.IndicatorResults))
		keptnEvent.labels["Evaluation Start"] = edData.EvaluationDetails.TimeStart
		keptnEvent.labels["Evaluation End"] = edData.EvaluationDetails.TimeEnd

		// Send Info Event
		ie := CreateInfoEvent(keptnEvent, dynatraceConfig, eh.Logger)
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
		dtHelper.SendEvent(ie, dtCreds)
	} else {
		eh.Logger.Info("Ignoring event.")
	}
	return nil
}

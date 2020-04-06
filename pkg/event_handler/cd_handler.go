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

	eh.Logger.Info("Checking if event of type " + eh.Event.Type() + " should be sent to Dynatrace...")

	if eh.Event.Type() == keptn.DeploymentFinishedEventType {
		dfData := &keptn.DeploymentFinishedEventData{}
		err := eh.Event.DataAs(dfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}
		de := createDeploymentEvent(dfData, shkeptncontext)

		dtHelper.SendEvent(de)

		// TODO: an additional channel (e.g. start-tests) to correctly determine the time when the tests actually start
		ie := createInfoEvent(dfData.Project, dfData.Stage, dfData.Service, dfData.TestStrategy, dfData.Image, dfData.Tag, shkeptncontext)
		if dfData.TestStrategy != "" {
			ie.Title = "Start Running Tests: " + dfData.TestStrategy
			ie.Description = "Start running tests: " + dfData.TestStrategy + " against " + dfData.Service
			dtHelper.SendEvent(ie)
		}
	} else if eh.Event.Type() == keptn.TestsFinishedEventType {
		tfData := &keptn.TestsFinishedEventData{}
		err := eh.Event.DataAs(tfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}
		ie := createInfoEvent(tfData.Project, tfData.Stage, tfData.Service, tfData.TestStrategy, "", "", shkeptncontext)
		ie.Title = "Stop Running Tests: " + tfData.TestStrategy
		ie.Description = "Stop running tests: " + tfData.TestStrategy + " against " + tfData.Service
		dtHelper.SendEvent(ie)

	} else if eh.Event.Type() == keptn.EvaluationDoneEventType {
		edData := &keptn.EvaluationDoneEventData{}
		err := eh.Event.DataAs(edData)
		if err != nil {
			fmt.Println("Error while parsing JSON payload: " + err.Error())
			return err
		}
		ie := createInfoEvent(edData.Project, edData.Stage, edData.Service, edData.TestStrategy, "", "", shkeptncontext)
		if edData.Result == "pass" || edData.Result == "warning" {
			if edData.TestStrategy == "real-user" {
				ie.Title = "Remediation action successful"
			} else {
				ie.Title = "Promote Artifact from " + edData.Stage + " to next stage"
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
				ie.Title = "NOT PROMOTING Artifact from " + edData.Stage + " due to failed evaluation"
			}
		} else {
			eh.Logger.Error("No valid deployment strategy defined in keptn event.")
			return nil
		}
		ie.Description = "Keptn evaluation status: " + edData.Result
		dtHelper.SendEvent(ie)
	} else {
		eh.Logger.Info("    Ignoring event.")
	}
	return nil
}

type dtTag struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

type dtTagRule struct {
	MeTypes []string `json:"meTypes"`
	Tags    []dtTag  `json:"tags"`
}

type dtAttachRules struct {
	TagRule []dtTagRule `json:"tagRule"`
}

type dtCustomProperties struct {
	Project            string `json:"Project"`
	Stage              string `json:"Stage"`
	Service            string `json:"Service"`
	TestStrategy       string `json:"Test strategy"`
	DeploymentStrategy string `json:"Deployment strategy"`
	Image              string `json:"Image"`
	Tag                string `json:"Tag"`
	KeptnContext       string `json:"Keptn context"`
}

type dtDeploymentEvent struct {
	EventType         string             `json:"eventType"`
	Source            string             `json:"source"`
	AttachRules       dtAttachRules      `json:"attachRules"`
	CustomProperties  dtCustomProperties `json:"customProperties"`
	DeploymentVersion string             `json:"deploymentVersion"`
	DeploymentName    string             `json:"deploymentName"`
	DeploymentProject string             `json:"deploymentProject"`
}

type dtInfoEvent struct {
	EventType        string             `json:"eventType"`
	Source           string             `json:"source"`
	AttachRules      dtAttachRules      `json:"attachRules"`
	CustomProperties dtCustomProperties `json:"customProperties"`
	Description      string             `json:"description"`
	Title            string             `json:"title"`
}

func createAttachRules(project string, stage string, service string) dtAttachRules {
	ar := dtAttachRules{
		TagRule: []dtTagRule{
			dtTagRule{
				MeTypes: []string{"SERVICE"},
				Tags: []dtTag{
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   project,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   stage,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   service,
					},
				},
			},
		},
	}
	return ar
}

func createCustomProperties(project string, stage string, service string, testStrategy string, image string, tag string, keptnContext string) dtCustomProperties {
	var customProperties dtCustomProperties
	customProperties.Project = project
	customProperties.Stage = stage
	customProperties.Service = service
	customProperties.TestStrategy = testStrategy
	customProperties.Image = image
	customProperties.Tag = tag
	customProperties.KeptnContext = keptnContext
	return customProperties
}

func createInfoEvent(project string, stage string, service string, testStrategy string, image string, tag string, keptnContext string) dtInfoEvent {
	ar := createAttachRules(project, stage, service)
	customProperties := createCustomProperties(project, stage, service, testStrategy, image, tag, keptnContext)

	var ie dtInfoEvent
	ie.AttachRules = ar
	ie.CustomProperties = customProperties
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "Keptn dynatrace-service"

	return ie
}

func createDeploymentEvent(event *keptn.DeploymentFinishedEventData, keptnContext string) dtDeploymentEvent {
	ar := createAttachRules(event.Project, event.Stage, event.Service)
	customProperties := createCustomProperties(event.Project, event.Stage, event.Service, event.TestStrategy, event.Image, event.Tag, keptnContext)

	var de dtDeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "Keptn dynatrace-service"
	de.DeploymentName = "Deploy " + event.Service + " " + event.Tag + " with strategy " + event.DeploymentStrategy
	de.DeploymentProject = event.Project
	de.DeploymentVersion = event.Tag
	de.AttachRules = ar
	de.CustomProperties = customProperties

	return de
}

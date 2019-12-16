package event_handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

type CDEventHandler struct {
	Logger *keptnutils.Logger
	Event  cloudevents.Event
}

func (eh CDEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	eh.Logger.Info("Received new event of type " + eh.Event.Type())

	dtTenant := os.Getenv("DT_TENANT")
	dtAPIToken := os.Getenv("DT_API_TOKEN")

	if dtTenant == "" || dtAPIToken == "" {
		eh.Logger.Error("No Dynatrace credentials defined in cluster. Could not send event.")
		return errors.New("no Dynatrace credentials defined in cluster")
	}
	eh.Logger.Info("Trying to send event to DT Tenant: " + os.Getenv("DT_TENANT"))

	if eh.Event.Type() == keptnevents.DeploymentFinishedEventType {
		dfData := &keptnevents.DeploymentFinishedEventData{}
		err := eh.Event.DataAs(dfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}
		de := createDeploymentEvent(dfData, shkeptncontext)
		sendDynatraceRequest(dtTenant, dtAPIToken, de, eh.Logger)

		// TODO: an additional channel (e.g. start-tests) to correctly determine the time when the tests actually start
		ie := createInfoEvent(dfData.Project, dfData.Stage, dfData.Service, dfData.TestStrategy, dfData.Image, dfData.Tag, shkeptncontext)
		if dfData.TestStrategy != "" {
			ie.Title = "Start Running Tests: " + dfData.TestStrategy
			ie.Description = "Start running tests: " + dfData.TestStrategy + " against " + dfData.Service
			sendDynatraceRequest(dtTenant, dtAPIToken, ie, eh.Logger)
		}
	} else if eh.Event.Type() == keptnevents.TestsFinishedEventType {
		tfData := &keptnevents.TestsFinishedEventData{}
		err := eh.Event.DataAs(tfData)
		if err != nil {
			eh.Logger.Error("Could not parse event payload: " + err.Error())
			return err
		}
		ie := createInfoEvent(tfData.Project, tfData.Stage, tfData.Service, tfData.TestStrategy, "", "", shkeptncontext)
		ie.Title = "Stop Running Tests: " + tfData.TestStrategy
		ie.Description = "Stop running tests: " + tfData.TestStrategy + " against " + tfData.Service
		sendDynatraceRequest(dtTenant, dtAPIToken, ie, eh.Logger)

	} else if eh.Event.Type() == keptnevents.EvaluationDoneEventType {
		edData := &keptnevents.EvaluationDoneEventData{}
		err := eh.Event.DataAs(edData)
		if err != nil {
			fmt.Println("Error while parsing JSON payload: " + err.Error())
			return err
		}
		ie := createInfoEvent(edData.Project, edData.Stage, edData.Service, edData.TestStrategy, "", "", shkeptncontext)
		if edData.Result == "pass" || edData.Result == "warning" {
			ie.Title = "Promote Artifact from " + edData.Stage + " to next stage"
		} else if edData.Result == "fail" && edData.DeploymentStrategy == "blue_green_service" {
			ie.Title = "Rollback Artifact (Switch Blue/Green) in " + edData.Stage
		} else if edData.Result == "fail" && edData.DeploymentStrategy == "direct" {
			ie.Title = "NOT PROMOTING Artifact from " + edData.Stage + " due to failed evaluation"
		} else {
			eh.Logger.Error("No valid deployment strategy defined in keptn event.")
			return nil
		}
		ie.Description = "Keptn evaluation status: " + edData.Result
		sendDynatraceRequest(dtTenant, dtAPIToken, ie, eh.Logger)
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

func createDeploymentEvent(event *keptnevents.DeploymentFinishedEventData, keptnContext string) dtDeploymentEvent {
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

func sendDynatraceRequest(dtTenant string, dtAPIToken string, dtEvent interface{}, logger *keptnutils.Logger) {
	jsonString, err := json.Marshal(dtEvent)
	if err != nil {
		logger.Error("Error while generating Dynatrace API Request payload.")
		return
	}
	url := "https://" + dtTenant + "/api/v1/events?Api-Token=" + dtAPIToken

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonString))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error while sending request to Dynatrace: " + err.Error())
		return
	}
	defer resp.Body.Close()

	logger.Debug("Response Status:" + resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	logger.Debug("Response Body:" + string(body))
}

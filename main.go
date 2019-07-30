package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

type keptnEvent struct {
	Specversion    string `json:"specversion"`
	Type           string `json:"type"`
	Source         string `json:"source"`
	ID             string `json:"id"`
	Time           string `json:"time"`
	Contenttype    string `json:"contenttype"`
	Shkeptncontext string `json:"shkeptncontext"`
	Data           struct {
		Githuborg          string `json:"githuborg"`
		Project            string `json:"project"`
		Teststrategy       string `json:"teststrategy"`
		Deploymentstrategy string `json:"deploymentstrategy"`
		Stage              string `json:"stage"`
		Service            string `json:"service"`
		Image              string `json:"image"`
		Tag                string `json:"tag"`
		EvaluationPassed   bool   `json:evaluationpassed,omitempty`
	} `json:"data"`
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
	GitHubOrg          string `json:"GitHub Org"`
	Project            string `json:"Project"`
	TestStrategy       string `json:"Test strategy"`
	DeploymentStrategy string `json:"Deployment strategy"`
	Stage              string `json:"Stage"`
	Service            string `json:"Service"`
	Image              string `json:"Image"`
	Tag                string `json:"Tag"`
	KeptnContext       string `json:"keptn context"`
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

func handler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var event keptnEvent
	err := decoder.Decode(&event)
	if err != nil {
		fmt.Println("Error while parsing JSON payload: " + err.Error())
		return
	}

	logger := keptnutils.NewLogger(event.Shkeptncontext, event.ID, "dynatrace-service")
	logger.Info("Received new event of type " + event.Type)

	dtTenant := os.Getenv("DT_TENANT")
	dtAPIToken := os.Getenv("DT_API_TOKEN")

	if dtTenant == "" || dtAPIToken == "" {
		logger.Error("No Dynatrace credentials defined in cluster. Could not send event.")
		return
	}
	logger.Info("Trying to send event to DT Tenant " + os.Getenv("DT_TENANT"))

	if event.Type == "sh.keptn.events.deployment-finished" {
		de := createDeploymentEvent(event)
		sendDynatraceRequest(dtTenant, dtAPIToken, de, event, logger)
		// We need an additional channel (e.g. start-tests) to correctly determine the time when the tests actually start
		ie := createInfoEvent(event)
		if event.Data.Teststrategy != "" {
			ie.Title = "Start Running Tests: " + event.Data.Teststrategy
			ie.Description = "Start Running Tests: " + event.Data.Teststrategy + " against " + event.Data.Service
			sendDynatraceRequest(dtTenant, dtAPIToken, ie, event, logger)
		}
	} else if event.Type == "sh.keptn.events.evaluation-done" {
		ie := createInfoEvent(event)
		if event.Data.EvaluationPassed {
			ie.Title = "Promote Artifact from " + event.Data.Stage + " to next stage"
		} else if !event.Data.EvaluationPassed && event.Data.Deploymentstrategy == "blue_green_service" {
			ie.Title = "Rollback Artifact (Switch Blue/Green) in " + event.Data.Stage
		} else if !event.Data.EvaluationPassed && event.Data.Deploymentstrategy == "direct" {
			ie.Title = "NOT PROMOTING Artifact from " + event.Data.Stage + " due to failed evaluation"
		} else {
			logger.Error("No valid deployment strategy defined in keptn event.")
			return
		}
		ie.Description = "keptn evaluation status: " + strconv.FormatBool(event.Data.EvaluationPassed)
		sendDynatraceRequest(dtTenant, dtAPIToken, ie, event, logger)
	} else if event.Type == "sh.keptn.events.tests-finished" {
		ie := createInfoEvent(event)
		ie.Title = "Stop Running Tests: " + event.Data.Teststrategy
		ie.Description = "Stop Running Tests: " + event.Data.Teststrategy + " against " + event.Data.Service
		sendDynatraceRequest(dtTenant, dtAPIToken, ie, event, logger)
	}
}

func createAttachRules(event keptnEvent) dtAttachRules {
	ar := dtAttachRules{
		TagRule: []dtTagRule{
			dtTagRule{
				MeTypes: []string{"SERVICE"},
				Tags: []dtTag{
					dtTag{
						Context: "ENVIRONMENT",
						Key:     "application",
						Value:   event.Data.Project,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "service",
						Value:   event.Data.Service,
					},
					dtTag{
						Context: "CONTEXTLESS",
						Key:     "environment",
						Value:   event.Data.Project + "-" + event.Data.Stage,
					},
				},
			},
		},
	}
	return ar
}

func createCustomProperties(event keptnEvent) dtCustomProperties {
	var customProperties dtCustomProperties
	customProperties.GitHubOrg = event.Data.Githuborg
	customProperties.Project = event.Data.Project
	customProperties.TestStrategy = event.Data.Teststrategy
	customProperties.Stage = event.Data.Stage
	customProperties.Service = event.Data.Service
	customProperties.Image = event.Data.Image
	customProperties.Tag = event.Data.Tag
	customProperties.KeptnContext = event.Shkeptncontext
	return customProperties
}

func createInfoEvent(event keptnEvent) dtInfoEvent {
	ar := createAttachRules(event)
	customProperties := createCustomProperties(event)

	var ie dtInfoEvent
	ie.AttachRules = ar
	ie.CustomProperties = customProperties
	ie.EventType = "CUSTOM_INFO"
	ie.Source = "keptn dynatrace-service"

	return ie
}

func createDeploymentEvent(event keptnEvent) dtDeploymentEvent {
	ar := createAttachRules(event)

	customProperties := createCustomProperties(event)

	var de dtDeploymentEvent
	de.EventType = "CUSTOM_DEPLOYMENT"
	de.Source = "keptn dynatrace-service"
	de.DeploymentName = "Deploy " + event.Data.Service + " " + event.Data.Tag + " with strategy " + event.Data.Deploymentstrategy
	de.DeploymentProject = event.Data.Project
	de.DeploymentVersion = event.Data.Tag
	de.AttachRules = ar
	de.CustomProperties = customProperties

	return de
}

func sendDynatraceRequest(dtTenant string, dtAPIToken string, dtEvent interface{}, event keptnEvent, logger *keptnutils.Logger) {
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

func main() {
	log.Print("Dynatrace service started.")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

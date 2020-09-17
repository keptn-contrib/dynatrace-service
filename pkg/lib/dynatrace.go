package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	keptnutils "github.com/keptn/go-utils/pkg/api/utils"
	"io/ioutil"
	"net/http"
	"os"

	keptn "github.com/keptn/go-utils/pkg/lib"
)

const DefaultOperatorVersion = "v0.8.0"
const sliResourceURI = "dynatrace/sli.yaml"
const Throughput = "throughput"
const ErrorRate = "error_rate"
const ResponseTimeP50 = "response_time_p50"
const ResponseTimeP90 = "response_time_p90"
const ResponseTimeP95 = "response_time_p95"

type criteriaObject struct {
	Operator        string
	Value           float64
	CheckPercentage bool
	IsComparison    bool
	CheckIncrease   bool
}

type DTAPIListResponse struct {
	Values []Values `json:"values"`
}
type Values struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DynatraceHelper struct {
	DynatraceCreds *credentials.DTCredentials
	Logger         keptn.LoggerInterface
	OperatorTag    string
	KeptnHandler   *keptn.Keptn
	KeptnBridge    string
}

func NewDynatraceHelper(keptnHandler *keptn.Keptn, dynatraceCreds *credentials.DTCredentials, logger keptn.LoggerInterface) *DynatraceHelper {
	return &DynatraceHelper{
		DynatraceCreds: dynatraceCreds,
		KeptnHandler:   keptnHandler,
		Logger:         logger,
	}
}

func (dt *DynatraceHelper) CreateCalculatedMetrics(project string) error {
	dt.Logger.Info("creating metric calc:service.topurlresponsetime" + project)
	responseTimeMetric := CreateCalculatedMetric("calc:service.topurlresponsetime"+project, "Top URL Response Time", "RESPONSE_TIME", "MICRO_SECOND", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SUM")
	responseTimeJSONPayload, _ := json.Marshal(&responseTimeMetric)
	_, err := dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.topurlresponsetime"+project, "PUT", responseTimeJSONPayload)
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.topurlresponsetime" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.topurlservicecalls" + project)
	topServiceCalls := CreateCalculatedMetric("calc:service.topurlservicecalls"+project, "Top URL Service Calls", "NON_DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topServiceCallsJSONPayload, _ := json.Marshal(&topServiceCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.topurlservicecalls"+project, "PUT", topServiceCallsJSONPayload)
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.topurlservicecalls" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.topurldbcalls" + project)
	topDBCalls := CreateCalculatedMetric("calc:service.topurldbcalls"+project, "Top URL DB Calls", "DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topDBCallsJSONPayload, _ := json.Marshal(&topDBCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.topurldbcalls"+project, "PUT", topDBCallsJSONPayload)
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.topurldbcalls" + project + ". " + err.Error())
	}

	return nil
}

func (dt *DynatraceHelper) CreateTestStepCalculatedMetrics(project string) error {
	dt.Logger.Info("creating metric calc:service.teststepresponsetime" + project)
	responseTimeMetric := CreateCalculatedTestStepMetric("calc:service.teststepresponsetime"+project, "Test Step Response Time", "RESPONSE_TIME", "MICRO_SECOND", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SUM")
	responseTimeJSONPayload, _ := json.Marshal(&responseTimeMetric)
	_, err := dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepresponsetime"+project, "PUT", responseTimeJSONPayload)
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepresponsetime" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.teststepservicecalls" + project)
	topServiceCalls := CreateCalculatedTestStepMetric("calc:service.teststepservicecalls"+project, "Test Step Service Calls", "NON_DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topServiceCallsJSONPayload, _ := json.Marshal(&topServiceCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepservicecalls"+project, "PUT", topServiceCallsJSONPayload)
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepservicecalls" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.teststepdbcalls" + project)
	topDBCalls := CreateCalculatedTestStepMetric("calc:service.teststepdbcalls"+project, "Test Step DB Calls", "DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topDBCallsJSONPayload, _ := json.Marshal(&topDBCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepdbcalls"+project, "PUT", topDBCallsJSONPayload)
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepdbcalls" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.teststepfailurerate" + project)
	failureRate := CreateCalculatedTestStepMetric("calc:service.teststepfailurerate"+project, "Test Step DB Calls", "FAILURE_RATE", "PERCENT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "OF_INTEREST_RATIO")
	failureRateJSONPayload, _ := json.Marshal(&failureRate)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepfailurerate"+project, "PUT", failureRateJSONPayload)
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepfailurerate" + project + ". " + err.Error())
	}

	return nil
}
func (dt *DynatraceHelper) ConfigureMonitoring(project string, shipyard keptn.Shipyard) error {

	dt.EnsureDTTaggingRulesAreSetUp()

	dt.EnsureProblemNotificationsAreSetUp()

	if project != "" {
		dt.CreateManagementZones(project, shipyard)

		configHandler := keptnutils.NewServiceHandler("configuration-service:8080")
		dt.CreateDashboard(project, shipyard)

		// try to create metric events - if one fails, don't fail the whole setup
		for _, stage := range shipyard.Stages {
			if stage.RemediationStrategy == "automated" {
				services, err := configHandler.GetAllServices(project, stage.Name)
				if err != nil {
					return fmt.Errorf("failed to retrieve services of project %s: %v", project, err.Error())
				}
				for _, service := range services {
					dt.CreateMetricEvents(project, stage.Name, service.ServiceName)
				}
			}
		}
	}
	return nil
}

/**
 * if dtCredsSecretName is passed and it is not dynatrace (=default) then we try to pull the secret based on that name and is it for this API Call
 */
func (dt *DynatraceHelper) sendDynatraceAPIRequest(apiPath string, method string, body []byte) (string, error) {

	if common.RunLocal || common.RunLocalTest {
		dt.Logger.Info("Dynatrace.sendDynatraceAPIRequest(RUNLOCAL) - not sending event to " +
			dt.DynatraceCreds.Tenant + "). Here is the payload: " + string(body))
		return "", nil
	}

	req, err := http.NewRequest(method, "https://"+dt.DynatraceCreds.Tenant+apiPath, bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Token "+dt.DynatraceCreds.ApiToken)
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+os.Getenv("version"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send Dynatrace API request: %v", err)
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return string(responseBody), fmt.Errorf("api request failed with status %s and response %s", resp.Status, string(responseBody))
	}

	return string(responseBody), nil
}

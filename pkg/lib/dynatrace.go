package lib

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/keptn/go-utils/pkg/models"

	"k8s.io/client-go/kubernetes"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DefaultOperatorVersion = "v0.5.2"
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

type DTCredentials struct {
	Tenant    string `json:"DT_TENANT" yaml:"DT_TENANT"`
	ApiToken  string `json:"DT_API_TOKEN" yaml:"DT_API_TOKEN"`
	PaaSToken string `json:"DT_PAAS_TOKEN" yaml:"DT_PAAS_TOKEN"`
}

type DynatraceHelper struct {
	KubeApi        *kubernetes.Clientset
	DynatraceCreds *DTCredentials
	Logger         keptnutils.LoggerInterface
	OperatorTag    string
}

func NewDynatraceHelper() (*DynatraceHelper, error) {
	dtHelper := &DynatraceHelper{}
	dtCreds, err := dtHelper.GetDTCredentials()
	if err != nil {
		return nil, err
	}
	dtHelper.DynatraceCreds = dtCreds
	return dtHelper, nil
}

func (dt *DynatraceHelper) CreateCalculatedMetrics(project string) error {
	dt.Logger.Info("creating metric calc:service.topurlresponsetime" + project)
	responseTimeMetric := CreateCalculatedMetric("calc:service.topurlresponsetime"+project, "Top URL Response Time", "RESPONSE_TIME", "MICRO_SECOND", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SUM")
	responseTimeJSONPayload, _ := json.Marshal(&responseTimeMetric)
	_, err := dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.topurlresponsetime"+project, "PUT", string(responseTimeJSONPayload))
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.topurlresponsetime" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.topurlservicecalls" + project)
	topServiceCalls := CreateCalculatedMetric("calc:service.topurlservicecalls"+project, "Top URL Service Calls", "NON_DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topServiceCallsJSONPayload, _ := json.Marshal(&topServiceCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.topurlservicecalls"+project, "PUT", string(topServiceCallsJSONPayload))
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.topurlservicecalls" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.topurldbcalls" + project)
	topDBCalls := CreateCalculatedMetric("calc:service.topurldbcalls"+project, "Top URL DB Calls", "DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topDBCallsJSONPayload, _ := json.Marshal(&topDBCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.topurldbcalls"+project, "PUT", string(topDBCallsJSONPayload))
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.topurldbcalls" + project + ". " + err.Error())
	}

	return nil
}

func (dt *DynatraceHelper) CreateTestStepCalculatedMetrics(project string) error {
	dt.Logger.Info("creating metric calc:service.teststepresponsetime" + project)
	responseTimeMetric := CreateCalculatedTestStepMetric("calc:service.teststepresponsetime"+project, "Test Step Response Time", "RESPONSE_TIME", "MICRO_SECOND", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SUM")
	responseTimeJSONPayload, _ := json.Marshal(&responseTimeMetric)
	_, err := dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepresponsetime"+project, "PUT", string(responseTimeJSONPayload))
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepresponsetime" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.teststepservicecalls" + project)
	topServiceCalls := CreateCalculatedTestStepMetric("calc:service.teststepservicecalls"+project, "Test Step Service Calls", "NON_DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topServiceCallsJSONPayload, _ := json.Marshal(&topServiceCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepservicecalls"+project, "PUT", string(topServiceCallsJSONPayload))
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepservicecalls" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.teststepdbcalls" + project)
	topDBCalls := CreateCalculatedTestStepMetric("calc:service.teststepdbcalls"+project, "Test Step DB Calls", "DATABASE_CHILD_CALL_COUNT", "COUNT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "SINGLE_VALUE")
	topDBCallsJSONPayload, _ := json.Marshal(&topDBCalls)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepdbcalls"+project, "PUT", string(topDBCallsJSONPayload))
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepdbcalls" + project + ". " + err.Error())
	}

	dt.Logger.Info("creating metric calc:service.teststepfailurerate" + project)
	failureRate := CreateCalculatedTestStepMetric("calc:service.teststepfailurerate"+project, "Test Step DB Calls", "FAILURE_RATE", "PERCENT", "CONTEXTLESS", "keptn_project", project, "URL", "{URL:Path}", "OF_INTEREST_RATIO")
	failureRateJSONPayload, _ := json.Marshal(&failureRate)
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/customMetric/service/"+"calc:service.teststepfailurerate"+project, "PUT", string(failureRateJSONPayload))
	if err != nil {
		dt.Logger.Error("could not create calculated metric calc:service.teststepfailurerate" + project + ". " + err.Error())
	}

	return nil
}

func (dt *DynatraceHelper) CreateManagementZones(project string, shipyard models.Shipyard) error {
	// get existing management zones
	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/managementZones", "GET", "")
	if err != nil {
		dt.Logger.Error("Could not retrieve management zones: " + err.Error())
	}
	mzs := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), mzs)
	if err != nil {
		dt.Logger.Error("Could not parse management zones list: " + err.Error())
	}

	found := false
	for _, mz := range mzs.Values {
		if mz.Name == "Keptn: "+project {
			found = true
		}
	}

	if !found {
		managementZone := CreateManagementZoneForProject(project)
		mzPayload, _ := json.Marshal(managementZone)
		_, err := dt.sendDynatraceAPIRequest("/api/config/v1/managementZones", "POST", string(mzPayload))
		if err != nil {
			dt.Logger.Error("Could not create management zone: " + err.Error())
		}
	}

	for _, stage := range shipyard.Stages {
		found := false
		for _, mz := range mzs.Values {
			if mz.Name == "Keptn: "+project+" "+stage.Name {
				found = true
			}
		}

		if !found {
			managementZone := CreateManagementZoneForStage(project, stage.Name)
			mzPayload, _ := json.Marshal(managementZone)
			_, err := dt.sendDynatraceAPIRequest("/api/config/v1/managementZones", "POST", string(mzPayload))
			if err != nil {
				dt.Logger.Error("Could not create management zone: " + err.Error())
			}
		}
	}

	return nil
}

func (dt *DynatraceHelper) sendDynatraceAPIRequest(apiPath string, method string, body string) (string, error) {
	req, err := http.NewRequest(method, "https://"+dt.DynatraceCreds.Tenant+apiPath, strings.NewReader(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Token "+dt.DynatraceCreds.ApiToken)
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+os.Getenv("version"))
	dt.Logger.Debug("Dynatrace Service version: " + os.Getenv("version"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		dt.Logger.Error("could not send Dynatrace API request: " + err.Error())
		return "", err
	}

	defer resp.Body.Close()
	responseBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		dt.Logger.Error("Response Status:" + resp.Status)
		return string(responseBody), errors.New(resp.Status)
	}

	return string(responseBody), nil
}

func (dt *DynatraceHelper) GetDTCredentials() (*DTCredentials, error) {
	kubeAPI, err := keptnutils.GetClientset(true)
	if err != nil {
		return nil, err
	}
	secret, err := kubeAPI.CoreV1().Secrets("keptn").Get("dynatrace", metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	if string(secret.Data["DT_TENANT"]) == "" || string(secret.Data["DT_API_TOKEN"]) == "" || string(secret.Data["DT_PAAS_TOKEN"]) == "" {
		return nil, errors.New("invalid or no Dynatrace credentials found")
	}

	dtCreds := &DTCredentials{}

	dtCreds.Tenant = string(secret.Data["DT_TENANT"])
	dtCreds.ApiToken = string(secret.Data["DT_API_TOKEN"])
	dtCreds.PaaSToken = string(secret.Data["DT_PAAS_TOKEN"])

	return dtCreds, nil
}

func writeFile(fileName string, content string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func deleteFile(fileName string) error {
	err := os.Remove(fileName)
	return err
}

func getHTTPResource(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

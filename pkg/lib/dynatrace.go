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

	v1 "k8s.io/api/core/v1"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DEFAULT_OPERATOR_VERSION = "v0.5.2"

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

type resourceKind struct {
	Kind string `yaml:"kind"`
}

type dtOperatorReleaseInfo struct {
	TagName string `json:"tag_name"`
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

func (dt *DynatraceHelper) EnsureDTIsInstalled() error {
	if dt.isDynatraceDeployed() {
		dt.Logger.Info("Skipping Dynatrace installation because Dynatrace is already deployed in the cluster.")
		return nil
	}
	// ensure that the namespace 'dynatrace' is available
	err := dt.createOrUpdateDynatraceNamespace()
	if err != nil {
		return err
	}
	// check if DT Secret is available
	err = dt.deployDTOperator()
	if err != nil {
		return err
	}

	err = dt.deployDynatrace()
	return err
}

func (dt *DynatraceHelper) EnsureDTTaggingRulesAreSetUp() error {
	dt.Logger.Info("Setting up auto-tagging rules in Dynatrace Tenant")

	// serviceRule := createAutoTaggingRule("keptn_service")

	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/autoTags", "GET", "")

	existingDTRules := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), existingDTRules)
	if err != nil {
		dt.Logger.Info("No existing Dynatrace tagging rules found")
	}

	for _, ruleName := range []string{"keptn_service", "keptn_stage", "keptn_project", "keptn_deployment"} {
		dt.deleteExistingDTTaggingRule(ruleName, existingDTRules)
		rule := createAutoTaggingRule(ruleName)
		err = dt.createDTTaggingRule(rule)
		if err != nil {
			dt.Logger.Error("Could not create auto tagging rule: " + err.Error())
		}

	}
	return nil
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

func (dt *DynatraceHelper) CreateDashboard(project string, shipyard models.Shipyard, services []string) error {
	keptnDomainCM, err := dt.KubeApi.CoreV1().ConfigMaps("keptn").Get("keptn-domain", metav1.GetOptions{})
	if err != nil {
		dt.Logger.Error("Could not retrieve keptn-domain ConfigMap: " + err.Error())
	}

	keptnDomain := keptnDomainCM.Data["app_domain"]

	// first, check if dashboard for this project already exists and delete that
	err = dt.DeleteExistingDashboard(project)
	if err != nil {
		return err
	}

	dt.Logger.Info("Creating Dashboard for project " + project)
	dashboard, err := CreateDynatraceDashboard(project, shipyard, keptnDomain, services)
	if err != nil {
		dt.Logger.Error("Could not create Dynatrace Dashboard for project " + project + ": " + err.Error())
		return err
	}

	dashboardPayload, _ := json.Marshal(dashboard)

	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/dashboards", "POST", string(dashboardPayload))

	if err != nil {
		dt.Logger.Error("Could not create Dynatrace Dashboard for project " + project + ": " + err.Error())
		return err
	}
	dt.Logger.Info("Dynatrace dashboard created successfully. You can view it here: https://" + dt.DynatraceCreds.Tenant + "/#dashboards")
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

func (dt *DynatraceHelper) DeleteExistingDashboard(project string) error {
	res, err := dt.sendDynatraceAPIRequest("/api/config/v1/dashboards", "GET", "")
	if err != nil {
		dt.Logger.Error("Could not retrieve list of existing Dynatrace dashboards: " + err.Error())
		return err
	}

	dtDashboardsResponse := &DTDashboardsResponse{}
	err = json.Unmarshal([]byte(res), dtDashboardsResponse)

	if err != nil {
		dt.Logger.Error("Could not parse list of existing Dynatrace dashboards: " + err.Error())
		return err
	}

	for _, dashboardItem := range dtDashboardsResponse.Dashboards {
		if dashboardItem.Name == project+"@keptn: Digital Delivery & Operations Dashboard" {
			res, err = dt.sendDynatraceAPIRequest("/api/config/v1/dashboards/"+dashboardItem.ID, "DELETE", "")
			if err != nil {
				dt.Logger.Error("Could not delete previous dashboard for project " + project + ": " + err.Error())
				return err
			}
		}
	}
	return nil
}

func (dt *DynatraceHelper) createDTTaggingRule(rule *DTTaggingRule) error {
	dt.Logger.Info("Creating DT tagging rule: " + rule.Name)
	payload, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/autoTags", "POST", string(payload))
	return err
}

func (dt *DynatraceHelper) deleteExistingDTTaggingRule(ruleName string, existingRules *DTAPIListResponse) {
	dt.Logger.Info("Deleting rule " + ruleName)
	for _, rule := range existingRules.Values {
		if rule.Name == ruleName {
			_, err := dt.sendDynatraceAPIRequest("/api/config/v1/autoTags/"+rule.ID, "DELETE", "")
			if err != nil {
				dt.Logger.Info("Could not delete rule " + rule.ID + ": " + err.Error())
			}
		}
	}
}

func createAutoTaggingRule(ruleName string) *DTTaggingRule {
	return &DTTaggingRule{
		Name: ruleName,
		Rules: []Rules{
			{
				Type:             "SERVICE",
				Enabled:          true,
				ValueFormat:      "{ProcessGroup:Environment:" + ruleName + "}",
				PropagationTypes: []string{"SERVICE_TO_PROCESS_GROUP_LIKE"},
				Conditions: []Conditions{
					{
						Key: Key{
							Attribute: "PROCESS_GROUP_CUSTOM_METADATA",
							DynamicKey: DynamicKey{
								Source: "ENVIRONMENT",
								Key:    ruleName,
							},
							Type: "PROCESS_CUSTOM_METADATA_KEY",
						},
						ComparisonInfo: ComparisonInfo{
							Type:          "STRING",
							Operator:      "EXISTS",
							Value:         nil,
							Negate:        false,
							CaseSensitive: nil,
						},
					},
				},
			},
		},
	}
}

func (dt *DynatraceHelper) sendDynatraceAPIRequest(apiPath string, method string, body string) (string, error) {
	req, err := http.NewRequest(method, "https://"+dt.DynatraceCreds.Tenant+apiPath, strings.NewReader(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Token "+dt.DynatraceCreds.ApiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		dt.Logger.Error("could not send Dynatrace API request: " + err.Error())
		return "", err
	}

	defer resp.Body.Close()
	responseBody, _ := ioutil.ReadAll(resp.Body)
	return string(responseBody), nil
}

func (dt *DynatraceHelper) isDynatraceDeployed() bool {
	_, err := dt.KubeApi.AppsV1().Deployments("dynatrace").Get("dynatrace-oneagent-operator", metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func (dt *DynatraceHelper) deployDTOperator() error {
	// get latest operator version
	resp, err := http.Get("https://api.github.com/repos/dynatrace/dynatrace-oneagent-operator/releases/latest")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		operatorInfo := &dtOperatorReleaseInfo{}

		err = json.Unmarshal(body, operatorInfo)

		if err == nil {
			dt.OperatorTag = operatorInfo.TagName
		}
	}
	if dt.OperatorTag != "" {
		dt.Logger.Info("Deploying Dynatrace operator " + dt.OperatorTag)
	} else {
		dt.Logger.Info("Could not fetch latest Dynatrace operator version. Using " + DEFAULT_OPERATOR_VERSION + " per default.")
		dt.OperatorTag = DEFAULT_OPERATOR_VERSION
	}

	platform := os.Getenv("PLATFORM")

	if platform == "" {
		platform = "kubernetes"
	}

	operatorYaml, err := getHTTPResource("https://raw.githubusercontent.com/Dynatrace/dynatrace-oneagent-operator/" + dt.OperatorTag + "/deploy/" + platform + ".yaml")
	if err != nil {
		dt.Logger.Error("could not fetch operator config: " + err.Error())
	}
	if platform == "openshift" {
		operatorYaml = strings.ReplaceAll(operatorYaml, "registry.connect.redhat.com", "quay.io")
	}

	operatorFileName := "operator.yaml"
	err = writeFile(operatorFileName, operatorYaml)
	if err != nil {
		dt.Logger.Error("could not save operator config: " + err.Error())
	}

	if err != nil {
		dt.Logger.Error("could not fetch Dynatrace operator config: " + err.Error())
		return err
	}

	_, err = keptnutils.ExecuteCommand("kubectl", []string{"apply", "-f", operatorFileName})
	if err != nil {
		_ = deleteFile(operatorFileName)
		return err
	}

	_ = deleteFile(operatorFileName)

	return nil
}

func (dt *DynatraceHelper) deployDynatrace() error {
	err := dt.createOrUpdateDTSecret()
	if err != nil {
		dt.Logger.Error("could not fetch Dynatrace CR: " + err.Error())
		return err
	}

	resp, err := http.Get("https://raw.githubusercontent.com/Dynatrace/dynatrace-oneagent-operator/" + dt.OperatorTag + "/deploy/cr.yaml")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		dt.Logger.Error("could not fetch Dynatrace CR: " + err.Error())
		return err
	}

	dynatraceDeploymentYaml := strings.Replace(string(body), "ENVIRONMENTID.live.dynatrace.com", dt.DynatraceCreds.Tenant, -1)

	dtYamlFileName := "dynatrace.yaml"
	err = writeFile(dtYamlFileName, dynatraceDeploymentYaml)

	if err != nil {
		dt.Logger.Error("could not store dynatrace config yaml: " + err.Error())
		return err
	}

	_, err = keptnutils.ExecuteCommand("kubectl", []string{"apply", "-f", dtYamlFileName})
	if err != nil {
		_ = deleteFile(dtYamlFileName)
		return err
	}

	_ = deleteFile(dtYamlFileName)

	return nil
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

func (dt *DynatraceHelper) createOrUpdateDTSecret() error {
	dtSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oneagent",
			Namespace: "dynatrace",
		},
		Data: map[string][]byte{
			"apiToken":  []byte(dt.DynatraceCreds.ApiToken),
			"paasToken": []byte(dt.DynatraceCreds.PaaSToken),
		},
		StringData: nil,
		Type:       "Opaque",
	}

	_, err := dt.KubeApi.CoreV1().Secrets("dynatrace").Create(dtSecret)
	if err != nil {
		_, err := dt.KubeApi.CoreV1().Secrets("dynatrace").Update(dtSecret)
		return err
	}
	return nil
}

// CreateOrUpdateDynatraceNamespace creates or updates the Dynatrace namespace
func (dt *DynatraceHelper) createOrUpdateDynatraceNamespace() error {
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dynatrace",
		},
	}

	_, err := dt.KubeApi.CoreV1().Namespaces().Create(namespace)

	if err != nil {
		_, err = dt.KubeApi.CoreV1().Namespaces().Update(namespace)
		if err != nil {
			return err
		}
	}
	return nil
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

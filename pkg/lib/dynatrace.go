package lib

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/api/core/v1"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DEFAULT_OPERATOR_VERSION = "v0.5.2"

type DTTaggingRule struct {
	Name  string  `json:"name"`
	Rules []Rules `json:"rules"`
}
type DynamicKey struct {
	Source string `json:"source"`
	Key    string `json:"key"`
}
type Key struct {
	Attribute  string     `json:"attribute"`
	DynamicKey DynamicKey `json:"dynamicKey"`
	Type       string     `json:"type"`
}
type ComparisonInfo struct {
	Type          string      `json:"type"`
	Operator      string      `json:"operator"`
	Value         interface{} `json:"value"`
	Negate        bool        `json:"negate"`
	CaseSensitive interface{} `json:"caseSensitive"`
}
type Conditions struct {
	Key            Key            `json:"key"`
	ComparisonInfo ComparisonInfo `json:"comparisonInfo"`
}
type Rules struct {
	Type             string       `json:"type"`
	Enabled          bool         `json:"enabled"`
	ValueFormat      string       `json:"valueFormat"`
	PropagationTypes []string     `json:"propagationTypes"`
	Conditions       []Conditions `json:"conditions"`
}

type DTTagResponse struct {
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

func NewDynatraceHelper() *DynatraceHelper {
	return &DynatraceHelper{}
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

	dtCreds, err := dt.GetDTCredentials()
	if err != nil {
		return err
	}
	dt.DynatraceCreds = dtCreds

	err = dt.deployDTOperator()
	if err != nil {
		return err
	}

	err = dt.deployDynatrace()
	return err
}

func (dt *DynatraceHelper) EnsureDTTaggingRulesAreSetUp() error {
	dtCreds, err := dt.GetDTCredentials()
	if err != nil {
		return err
	}
	dt.DynatraceCreds = dtCreds

	dt.Logger.Info("Setting up auto-tagging rules in Dynatrace Tenant")

	// serviceRule := createAutoTaggingRule("keptn_service")

	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/autoTags", "GET", "")

	existingDTRules := &DTTagResponse{}

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

func (dt *DynatraceHelper) createDTTaggingRule(rule *DTTaggingRule) error {
	dt.Logger.Info("Creating DT tagging rule: " + rule.Name)
	payload, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/autoTags", "POST", string(payload))
	return err
}

func (dt *DynatraceHelper) deleteExistingDTTaggingRule(ruleName string, existingRules *DTTagResponse) {
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

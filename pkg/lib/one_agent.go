package lib

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	keptn "github.com/keptn/go-utils/pkg/lib"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dtOperatorReleaseInfo struct {
	TagName string `json:"tag_name"`
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

func (dt *DynatraceHelper) isDynatraceDeployed() bool {
	_, err := dt.KubeApi.AppsV1().Deployments("dynatrace").Get("dynatrace-oneagent-operator", metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func (dt *DynatraceHelper) deployDTOperator() error {
	dt.OperatorTag = DefaultOperatorVersion

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

	_, err = keptn.ExecuteCommand("kubectl", []string{"apply", "-f", operatorFileName})
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

	_, err = keptn.ExecuteCommand("kubectl", []string{"apply", "-f", dtYamlFileName})
	if err != nil {
		_ = deleteFile(dtYamlFileName)
		return err
	}

	_ = deleteFile(dtYamlFileName)

	return nil
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

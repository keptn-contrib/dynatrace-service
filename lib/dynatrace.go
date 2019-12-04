package lib

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/api/policy/v1beta1"

	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/rest"

	"github.com/ghodss/yaml"
	appsv1beta1 "k8s.io/api/apps/v1beta1"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const DEFAULT_OPERATOR_VERSION = "v0.5.2"

type dtCredentials struct {
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
	DynatraceCreds *dtCredentials
	Logger         *keptnutils.Logger
}

func NewDynatraceHelper() *DynatraceHelper {
	return &DynatraceHelper{}
}

func (dt *DynatraceHelper) EnsureDTIsInstalled() error {

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

	return nil
}

func (dt *DynatraceHelper) deployDTOperator() error {
	var operatorTagName string
	// get latest operator version
	resp, err := http.Get("https://api.github.com/repos/dynatrace/dynatrace-oneagent-operator/releases/latest")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		operatorInfo := &dtOperatorReleaseInfo{}

		err = json.Unmarshal(body, operatorInfo)

		if err == nil {
			operatorTagName = operatorInfo.TagName
		}
	}
	if operatorTagName != "" {
		dt.Logger.Info("Deploying Dynatrace operator " + operatorTagName)
	} else {
		dt.Logger.Info("Could not fetch latest Dynatrace operator version. Using " + DEFAULT_OPERATOR_VERSION + " per default.")
		operatorTagName = DEFAULT_OPERATOR_VERSION
	}

	resp, err = http.Get("https://raw.githubusercontent.com/Dynatrace/dynatrace-oneagent-operator/" + operatorTagName + "/deploy/kubernetes.yaml")
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	operatorResources := strings.Split(string(body), "---")

	config, _ := rest.InClusterConfig()
	apixClient, err := apixv1beta1client.NewForConfig(config)
	if err != nil {
		return err
	}

	for _, resource := range operatorResources {

		decode := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decode([]byte(resource), nil, nil)
		if err != nil {
			dt.Logger.Error("could not determine resource kind: " + err.Error())
			return err
		}

		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case "ServiceAccount":
			err := dt.createOrUpdateServiceAccount(obj.(*corev1.ServiceAccount))
			if err != nil {
				return err
			}
			break
		case "PodSecurityPolicy":
			err := dt.createOrUpdatePodSecurityPolicy(obj.(*v1beta1.PodSecurityPolicy))
			if err != nil {
				return err
			}
			break
		case "Role":
			err := dt.createOrUpdateRole(obj.(*rbacv1beta1.Role))
			if err != nil {
				return err
			}
			break
		case "RoleBinding":
			err := dt.createOrUpdateRoleBinding(obj.(*rbacv1beta1.RoleBinding))
			if err != nil {
				return err
			}
			break
		case "ClusterRoleBinding":
			err := dt.createOrUpdateClusterRoleBinding(obj.(*rbacv1beta1.ClusterRoleBinding))
			if err != nil {
				return err
			}
			break
		case "ClusterRole":
			err := dt.createOrUpdateClusterRole(obj.(*rbacv1beta1.ClusterRole))
			if err != nil {
				return err
			}
			break
		case "Deployment":
			err := dt.createOrUpdateDeployment(obj.(*appsv1beta1.Deployment))
			if err != nil {
				return err
			}
			break
		case "CustomResourceDefinition":

			err := dt.createOrUpdateCRD(apixClient, obj.(*apixv1beta1.CustomResourceDefinition))
			if err != nil {
				return err
			}
			break
		}

		resKind := &resourceKind{}

		err = yaml.Unmarshal([]byte(resource), resKind)
		if err != nil {
			dt.Logger.Error("could not determine resource kind: " + err.Error())
			return err
		}

	}
	/*
		dynClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}


			dtResource = schema.GroupVersionResource{
				Group:    "dynatrace.com",
				Version:  "v1alpha1",
				Resource: "oneagents.dynatrace.com",
			}


				dtSecret :=
					dt.KubeApi.CoreV1().Secrets("dynatrace").Create()


	*/
	return nil
}

func (dt *DynatraceHelper) createOrUpdateCRD(apixClient *apixv1beta1client.ApiextensionsV1beta1Client, obj *apixv1beta1.CustomResourceDefinition) error {
	api := apixClient.CustomResourceDefinitions()
	_, err := api.Get(obj.Name, metav1.GetOptions{})
	if err != nil {
		_, err = api.Create(obj)
	} else {
		_, err = api.Update(obj)
	}
	return err
}

func (dt *DynatraceHelper) createOrUpdateClusterRole(obj *rbacv1beta1.ClusterRole) error {
	api := dt.KubeApi.RbacV1beta1().ClusterRoles()
	_, err := api.Get(obj.Name, metav1.GetOptions{})
	if err != nil {
		_, err = api.Create(obj)
	} else {
		_, err = api.Update(obj)
	}
	return err
}

func (dt *DynatraceHelper) createOrUpdateClusterRoleBinding(obj *rbacv1beta1.ClusterRoleBinding) error {
	api := dt.KubeApi.RbacV1beta1().ClusterRoleBindings()
	_, err := api.Get(obj.Name, metav1.GetOptions{})

	if err != nil {
		_, err = api.Create(obj)
	} else {
		_, err = api.Update(obj)
	}

	return err
}

func (dt *DynatraceHelper) createOrUpdateRoleBinding(obj *rbacv1beta1.RoleBinding) error {
	api := dt.KubeApi.RbacV1beta1().RoleBindings("dynatrace")
	_, err := api.Get(obj.Name, metav1.GetOptions{})

	if err != nil {
		_, err = api.Create(obj)
	} else {
		_, err = api.Update(obj)
	}

	return err
}

func (dt *DynatraceHelper) createOrUpdateRole(obj *rbacv1beta1.Role) error {
	api := dt.KubeApi.RbacV1beta1().Roles("dynatrace")

	_, err := api.Get(obj.Name, metav1.GetOptions{})

	if err != nil {
		_, err = api.Create(obj)
	} else {
		_, err = api.Update(obj)
	}

	return err
}

func (dt *DynatraceHelper) createOrUpdatePodSecurityPolicy(obj *v1beta1.PodSecurityPolicy) error {
	api := dt.KubeApi.PolicyV1beta1().PodSecurityPolicies()

	_, err := api.Get(obj.Name, metav1.GetOptions{})

	if err != nil {
		_, err = api.Create(obj)
	} else {
		_, err = api.Update(obj)
	}

	return err
}

func (dt *DynatraceHelper) createOrUpdateDeployment(deployment *appsv1beta1.Deployment) error {
	deploymentAPI := dt.KubeApi.AppsV1beta1().Deployments("dynatrace")

	_, err := deploymentAPI.Get(deployment.Name, metav1.GetOptions{})

	if err != nil {
		_, err := deploymentAPI.Create(deployment)
		if err != nil {
			return err
		}
	} else {
		deploymentAPI.Update(deployment)
	}
	return nil
}

func (dt *DynatraceHelper) createOrUpdateServiceAccount(serviceAccount *corev1.ServiceAccount) error {
	saAPI := dt.KubeApi.CoreV1().ServiceAccounts("dynatrace")

	_, err := saAPI.Get(serviceAccount.Name, metav1.GetOptions{})
	if err != nil {
		_, err := saAPI.Create(serviceAccount)
		if err != nil {
			return err
		}
	} else {
		saAPI.Update(serviceAccount)
	}
	return nil
}

func (dt *DynatraceHelper) GetDTCredentials() (*dtCredentials, error) {
	secret, err := dt.KubeApi.CoreV1().Secrets("keptn").Get("dynatrace", metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	if string(secret.Data["DT_TENANT"]) == "" || string(secret.Data["DT_API_TOKEN"]) == "" || string(secret.Data["DT_PAAS_TOKEN"]) == "" {
		return nil, errors.New("invalid or no Dynatrace credentials found")
	}

	dtCreds := &dtCredentials{}

	dtCreds.Tenant = string(secret.Data["DT_TENANT"])
	dtCreds.ApiToken = string(secret.Data["DT_API_TOKEN"])
	dtCreds.PaaSToken = string(secret.Data["DT_PAAS_TOKEN"])

	return dtCreds, nil
}

// CreateOrUpdateDynatraceNamespace creates or updates the Dynatrace namespace
func (p *DynatraceHelper) createOrUpdateDynatraceNamespace() error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dynatrace",
		},
	}
	_, err := p.KubeApi.CoreV1().Namespaces().Create(namespace)

	if err != nil {
		_, err = p.KubeApi.CoreV1().Namespaces().Update(namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

package common

import (
	"errors"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var RunLocal = (os.Getenv("env") == "runlocal")
var RunLocalTest = (os.Getenv("env") == "runlocaltest")

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	if RunLocal || RunLocalTest {
		return nil, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

/**
 * Returns the Keptn Domain stored in the keptn-domainconfigmap
 */
func GetKeptnDomain() (string, error) {
	kubeAPI, err := GetKubernetesClient()
	if kubeAPI == nil || err != nil {
		return "", err
	}

	keptnDomainCM, errCM := kubeAPI.CoreV1().ConfigMaps("keptn").Get("keptn-domain", metav1.GetOptions{})
	if errCM != nil {
		return "", errors.New("Could not retrieve keptn-domain ConfigMap: " + errCM.Error())
	}

	keptnDomain := keptnDomainCM.Data["app_domain"]
	return keptnDomain, nil
}

package credentials

import (
	"context"
	"errors"
	"fmt"
	"os"

	keptnkubeutils "github.com/keptn/kubernetes-utils/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var namespace = getPodNamespace()

var ErrSecretNotFound = errors.New("secret not found")

type K8sSecretReader struct {
	K8sClient kubernetes.Interface
}

func NewK8sSecretReader(k8sClient kubernetes.Interface) (*K8sSecretReader, error) {
	k8sSecretReader := &K8sSecretReader{}
	if k8sClient != nil {
		k8sSecretReader.K8sClient = k8sClient
	} else {
		client, err := getKubernetesClient()
		if err != nil {
			return nil, fmt.Errorf("could not initialize NewK8sSecretReader: %s", err.Error())
		}
		k8sSecretReader.K8sClient = client
	}
	return k8sSecretReader, nil
}

func (kcr *K8sSecretReader) ReadSecret(secretName, namespace, secretKey string) (string, error) {
	secret, err := kcr.K8sClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if string(secret.Data[secretKey]) == "" {
		return "", ErrSecretNotFound
	}
	return string(secret.Data[secretKey]), nil
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	useInClusterConfig := os.Getenv("KUBERNETES_SERVICE_HOST") != ""
	return keptnkubeutils.GetClientset(useInClusterConfig)
}

func getPodNamespace() string {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return "keptn"
	}

	return ns
}

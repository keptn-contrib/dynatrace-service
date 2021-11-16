package credentials

import (
	"context"
	"fmt"
	"os"

	keptnkubeutils "github.com/keptn/kubernetes-utils/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sSecretReader struct {
	K8sClient kubernetes.Interface
}

func NewK8sSecretReader(k8sClient kubernetes.Interface) *K8sSecretReader {
	return &K8sSecretReader{K8sClient: k8sClient}
}

func NewDefaultK8sSecretReader() (*K8sSecretReader, error) {
	useInClusterConfig := os.Getenv("KUBERNETES_SERVICE_HOST") != ""
	k8sClient, err := keptnkubeutils.GetClientset(useInClusterConfig)
	if err != nil {
		return nil, fmt.Errorf("could not initialize K8sSecretReader: %s", err.Error())
	}
	return &K8sSecretReader{K8sClient: k8sClient}, nil
}

func (kcr *K8sSecretReader) ReadSecret(secretName string, secretKey string) (string, error) {
	secret, err := kcr.K8sClient.CoreV1().Secrets(getPodNamespace()).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	secretData, found := secret.Data[secretKey]
	if !found {
		return "", fmt.Errorf("key \"%s\" was not found in secret \"%s\"", secretKey, secretName)
	}
	return string(secretData), nil
}

func getPodNamespace() string {
	// TODO: 2021-11-16: centralize access to environment variables, maybe use mock?
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return "keptn"
	}

	return ns
}

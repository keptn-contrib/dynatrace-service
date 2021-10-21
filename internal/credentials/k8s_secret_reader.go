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

var ErrSecretNotFound = errors.New("secret not found")

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
	if string(secret.Data[secretKey]) == "" {
		return "", ErrSecretNotFound
	}
	return string(secret.Data[secretKey]), nil
}

func getPodNamespace() string {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return "keptn"
	}

	return ns
}

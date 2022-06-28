package credentials

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
	"github.com/keptn/go-utils/pkg/common/kubeutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8sSecretReader reads from a K8s secret.
type K8sSecretReader struct {
	K8sClient kubernetes.Interface
}

// NewK8sSecretReader creates a new K8sSecretReader using the specified kubernetes.Interface.
func NewK8sSecretReader(k8sClient kubernetes.Interface) *K8sSecretReader {
	return &K8sSecretReader{K8sClient: k8sClient}
}

// NewDefaultK8sSecretReader creates a new K8sSecretReader using the default K8s client.
func NewDefaultK8sSecretReader() (*K8sSecretReader, error) {
	useInClusterConfig := env.GetKubernetesServiceHost() != ""
	k8sClient, err := kubeutils.GetClientSet(useInClusterConfig)
	if err != nil {
		return nil, fmt.Errorf("could not initialize K8sSecretReader: %s", err.Error())
	}
	return NewK8sSecretReader(k8sClient), nil
}

// ReadSecret reads the value of a key from the specified secret or returns an error.
func (kcr *K8sSecretReader) ReadSecret(ctx context.Context, secretName string, secretKey string) (string, error) {
	secret, err := kcr.K8sClient.CoreV1().Secrets(env.GetPodNamespace()).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	secretData, found := secret.Data[secretKey]
	if !found {
		return "", fmt.Errorf("key \"%s\" was not found in secret \"%s\"", secretKey, secretName)
	}
	return string(secretData), nil
}

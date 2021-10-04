package credentials

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"

	"github.com/keptn-contrib/dynatrace-service/internal/url"
	keptnkubeutils "github.com/keptn/kubernetes-utils/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var namespace = getPodNamespace()

var ErrSecretNotFound = errors.New("secret not found")

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

type SecretReader interface {
	ReadSecret(secretName, namespace, secretKey string) (string, error)
}

type K8sCredentialReader struct {
	K8sClient kubernetes.Interface
}

func NewK8sCredentialReader(k8sClient kubernetes.Interface) (*K8sCredentialReader, error) {
	k8sCredentialReader := &K8sCredentialReader{}
	if k8sClient != nil {
		k8sCredentialReader.K8sClient = k8sClient
	} else {
		client, err := getKubernetesClient()
		if err != nil {
			return nil, fmt.Errorf("could not initialize K8sCredentialReader: %s", err.Error())
		}
		k8sCredentialReader.K8sClient = client
	}
	return k8sCredentialReader, nil
}

func (kcr *K8sCredentialReader) ReadSecret(secretName, namespace, secretKey string) (string, error) {
	secret, err := kcr.K8sClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if string(secret.Data[secretKey]) == "" {
		return "", ErrSecretNotFound
	}
	return string(secret.Data[secretKey]), nil
}

type OSEnvCredentialReader struct{}

func (OSEnvCredentialReader) ReadSecret(secretName, namespace, secretKey string) (string, error) {
	secret := os.Getenv(secretKey)
	if secret == "" {
		return secret, ErrSecretNotFound
	}
	return secret, nil
}

//go:generate moq --skip-ensure -pkg credentials_mock -out ./mock/credential_manager_mock.go . CredentialManagerInterface
type CredentialManagerInterface interface {
	GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error)
	GetKeptnAPICredentials() (*KeptnCredentials, error)
}

type CredentialManager struct {
	SecretReader SecretReader
}

func NewCredentialManager(sr SecretReader) (*CredentialManager, error) {
	cm := &CredentialManager{}
	if sr != nil {
		cm.SecretReader = sr
	} else {
		sr, err := NewK8sCredentialReader(nil)
		if err != nil {
			return nil, fmt.Errorf("could not initialize CredentialManager: %s", err.Error())
		}
		cm.SecretReader = sr
	}
	return cm, nil
}

func (cm *CredentialManager) GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error) {
	dtTenant, err := cm.SecretReader.ReadSecret(secretName, namespace, "DT_TENANT")
	if err != nil {
		return nil, fmt.Errorf("key DT_TENANT was not found in secret \"%s\"", secretName)
	}

	dtAPIToken, err := cm.SecretReader.ReadSecret(secretName, namespace, "DT_API_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("key DT_API_TOKEN was not found in secret \"%s\"", secretName)
	}

	dtAPIToken = strings.TrimSpace(dtAPIToken)
	return NewDynatraceCredentials(dtTenant, dtAPIToken)
}

func (cm *CredentialManager) GetKeptnAPICredentials() (*KeptnCredentials, error) {
	secretName := "dynatrace"

	apiURL, err := cm.SecretReader.ReadSecret(secretName, namespace, "KEPTN_API_URL")
	if err != nil {
		apiURL = os.Getenv("KEPTN_API_URL")
		if apiURL == "" {
			return nil, fmt.Errorf("key KEPTN_API_URL was not found in secret \"%s\" or environment variables", secretName)
		}
	}

	apiToken, err := cm.SecretReader.ReadSecret(secretName, namespace, "KEPTN_API_TOKEN")
	if err != nil {
		apiToken = os.Getenv("KEPTN_API_TOKEN")
		if apiToken == "" {
			return nil, fmt.Errorf("key KEPTN_API_TOKEN was not found in secret \"%s\" or environment variables", secretName)
		}
	}

	apiToken = strings.TrimSpace(apiToken)
	return NewKeptnCredentials(apiURL, apiToken)
}

func (cm *CredentialManager) GetKeptnBridgeURL() (string, error) {
	secretName := "dynatrace"

	bridgeURL, err := cm.SecretReader.ReadSecret(secretName, namespace, "KEPTN_BRIDGE_URL")

	if err != nil {
		bridgeURL = os.Getenv("KEPTN_BRIDGE_URL")
		if bridgeURL == "" {
			return "", fmt.Errorf("key KEPTN_BRIDGE_URL was not found in secret \"%s\" or environment variables", secretName)
		}
	}

	return url.MakeCleanURL(bridgeURL)
}

// GetKeptnCredentials retrieves the Keptn Credentials from the "dynatrace" secret
func GetKeptnCredentials() (*KeptnCredentials, error) {
	cm, err := NewCredentialManager(nil)
	if err != nil {
		return nil, err
	}
	return cm.GetKeptnAPICredentials()
}

// CheckKeptnConnection verifies wether a connection to the Keptn API can be established
func CheckKeptnConnection(keptnCredentials *KeptnCredentials) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	keptnAuthURL := keptnCredentials.GetAPIURL() + "/v1/auth"
	req, err := http.NewRequest(http.MethodGet, keptnAuthURL, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", keptnCredentials.GetAPIToken())

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not authenticate at Keptn API: " + err.Error())
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid Keptn API Token: received 401 - Unauthorized from " + keptnAuthURL)
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received unexpected response from %s: %d", keptnAuthURL, resp.StatusCode)
	}
	return nil
}

// GetKeptnBridgeURL returns the bridge URL
func GetKeptnBridgeURL() (string, error) {
	cm, err := NewCredentialManager(nil)
	if err != nil {
		return "", err
	}
	return cm.GetKeptnBridgeURL()
}

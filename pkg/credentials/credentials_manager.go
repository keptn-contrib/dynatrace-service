package credentials

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DTCredentials is a struct for the tenant and api token information
type DTCredentials struct {
	Tenant   string `json:"DT_TENANT" yaml:"DT_TENANT"`
	ApiToken string `json:"DT_API_TOKEN" yaml:"DT_API_TOKEN"`
}

type KeptnAPICredentials struct {
	APIURL   string `json:"KEPTN_API_URL" yaml:"KEPTN_API_URL"`
	APIToken string `json:"KEPTN_API_TOKEN" yaml:"KEPTN_API_TOKEN"`
}

var namespace = getPodNamespace()

var ErrSecretNotFound = errors.New("secret not found")

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
		client, err := common.GetKubernetesClient()
		if err != nil {
			return nil, fmt.Errorf("could not initialize K8sCredentialReader: %s", err.Error())
		}
		k8sCredentialReader.K8sClient = client
	}
	return k8sCredentialReader, nil
}

func (kcr *K8sCredentialReader) ReadSecret(secretName, namespace, secretKey string) (string, error) {
	secret, err := kcr.K8sClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
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
	GetDynatraceCredentials(dynatraceConfig *config.DynatraceConfigFile) (*DTCredentials, error)
	GetKeptnAPICredentials() (*KeptnAPICredentials, error)
}

type CredentialManager struct {
	SecretReader SecretReader
}

func NewCredentialManager(sr SecretReader) (*CredentialManager, error) {
	cm := &CredentialManager{}
	if sr != nil {
		cm.SecretReader = sr
	} else if common.RunLocal || common.RunLocalTest {
		cm.SecretReader = &OSEnvCredentialReader{}
	} else {
		sr, err := NewK8sCredentialReader(nil)
		if err != nil {
			return nil, fmt.Errorf("could not initialize CredentialManager: %s", err.Error())
		}
		cm.SecretReader = sr
	}
	return cm, nil
}

func (cm *CredentialManager) GetDynatraceCredentials(dynatraceConfig *config.DynatraceConfigFile) (*DTCredentials, error) {
	secretName := "dynatrace"
	if dynatraceConfig != nil && len(dynatraceConfig.DtCreds) > 0 {
		secretName = dynatraceConfig.DtCreds
	}

	dtTenant, err := cm.SecretReader.ReadSecret(secretName, namespace, "DT_TENANT")
	if err != nil {
		return nil, fmt.Errorf("DT_TENANT was not found in \"%s\" secret.", secretName)
	}

	dtAPIToken, err := cm.SecretReader.ReadSecret(secretName, namespace, "DT_API_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("DT_API_TOKEN was not found in \"%s\" secret.", secretName)
	}

	return &DTCredentials{Tenant: getCleanURL(dtTenant), ApiToken: getCleanToken(dtAPIToken)}, nil
}

func (cm *CredentialManager) GetKeptnAPICredentials() (*KeptnAPICredentials, error) {
	secretName := "dynatrace"

	apiURL, err := cm.SecretReader.ReadSecret(secretName, namespace, "KEPTN_API_URL")
	if err != nil {
		apiURL = os.Getenv("KEPTN_API_URL")
		if apiURL == "" {
			return nil, fmt.Errorf("KEPTN_API_URL was not found in \"%s\" secret or environment variables.", secretName)
		}
	}

	apiToken, err := cm.SecretReader.ReadSecret(secretName, namespace, "KEPTN_API_TOKEN")
	if err != nil {
		apiToken = os.Getenv("KEPTN_API_TOKEN")
		if apiToken == "" {
			return nil, fmt.Errorf("KEPTN_API_TOKEN was not found in \"%s\" secret or environment variables.", secretName)
		}
	}

	return &KeptnAPICredentials{APIURL: getCleanURL(apiURL), APIToken: getCleanToken(apiToken)}, nil
}

func (cm *CredentialManager) GetKeptnBridgeURL() (string, error) {
	secretName := "dynatrace"

	bridgeURL, err := cm.SecretReader.ReadSecret(secretName, namespace, "KEPTN_BRIDGE_URL")

	if err != nil {
		bridgeURL = os.Getenv("KEPTN_BRIDGE_URL")
		if bridgeURL == "" {
			return "", errors.New("KEPTN_BRIDGE_URL was not found in dynatrace secret or environment variables.")
		}
	}

	return getCleanURL(bridgeURL), nil
}

// Trims new lines and trailing slashes, defaults to https if http not specified
func getCleanURL(url string) string {
	url = strings.Trim(url, "\n")
	url = strings.TrimSuffix(url, "/")

	// ensure that url uses https if http has not been explicitly specified
	if !strings.HasPrefix(url, "http://") {
		url = "https://" + strings.TrimPrefix(url, "https://")
	}

	return url
}

func getCleanToken(token string) string {
	return strings.Trim(token, "\n")
}

// GetDynatraceCredentials reads the Dynatrace credentials from the secret. Therefore, it first checks
// if a secret is specified in the dynatrace.conf.yaml and if not defaults to the secret "dynatrace"
func GetDynatraceCredentials(dynatraceConfig *config.DynatraceConfigFile) (*DTCredentials, error) {

	cm, err := NewCredentialManager(nil)
	if err != nil {
		return nil, err
	}
	return cm.GetDynatraceCredentials(dynatraceConfig)
}

// GetKeptnCredentials retrieves the Keptn Credentials from the "dynatrace" secret
func GetKeptnCredentials() (*KeptnAPICredentials, error) {
	cm, err := NewCredentialManager(nil)
	if err != nil {
		return nil, err
	}
	return cm.GetKeptnAPICredentials()
}

// CheckKeptnConnection verifies wether a connection to the Keptn API can be established
func CheckKeptnConnection(keptnCredentials *KeptnAPICredentials) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest(http.MethodGet, keptnCredentials.APIURL+"/v1/auth", nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", keptnCredentials.APIToken)

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not authenticate at Keptn API: " + err.Error())
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid Keptn API Token: received 401 - Unauthorized from " + keptnCredentials.APIURL + "/v1/auth")
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received unexpected response from "+keptnCredentials.APIURL+"/v1/auth: %d", resp.StatusCode)
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

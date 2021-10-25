package credentials

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
)

//go:generate moq --skip-ensure -pkg credentials_mock -out ./mock/keptn_credentials_provider_mock.go . KeptnCredentialsProvider
type KeptnCredentialsProvider interface {
	GetKeptnCredentials() (*KeptnCredentials, error)
}

type KeptnCredentialsReader struct {
	SecretReader              *K8sSecretReader
	EnvironmentVariableReader *env.OSEnvironmentVariableReader
}

func NewKeptnCredentialsReader(sr *K8sSecretReader) *KeptnCredentialsReader {
	return &KeptnCredentialsReader{SecretReader: sr, EnvironmentVariableReader: env.NewOSEnvironmentVariableReader()}
}

func NewDefaultKeptnCredentialsReader() (*KeptnCredentialsReader, error) {
	sr, err := NewDefaultK8sSecretReader()
	if err != nil {
		return nil, fmt.Errorf("could not initialize KeptnCredentialsReader: %s", err.Error())
	}

	return &KeptnCredentialsReader{
		SecretReader:              sr,
		EnvironmentVariableReader: env.NewOSEnvironmentVariableReader(),
	}, nil
}

func (cm *KeptnCredentialsReader) GetKeptnCredentials() (*KeptnCredentials, error) {
	apiURL, err := cm.SecretReader.ReadSecret(dynatraceSecretName, keptnAPIURLName)
	if err != nil {
		val, found := cm.EnvironmentVariableReader.Read(keptnAPIURLName)
		if !found {
			return nil, fmt.Errorf("key %s was not found in secret \"%s\" or environment variables", keptnAPIURLName, dynatraceSecretName)
		}
		apiURL = val
	}

	apiToken, err := cm.SecretReader.ReadSecret(dynatraceSecretName, keptnAPITokenName)
	if err != nil {
		val, found := cm.EnvironmentVariableReader.Read(keptnAPITokenName)
		if !found {
			return nil, fmt.Errorf("key %s was not found in secret \"%s\" or environment variables", keptnAPITokenName, dynatraceSecretName)
		}
		apiToken = val
	}

	bridgeURL, err := cm.SecretReader.ReadSecret(dynatraceSecretName, keptnBridgeURLName)
	if err != nil {
		bridgeURL, _ = cm.EnvironmentVariableReader.Read(keptnBridgeURLName)
	}

	return NewKeptnCredentials(apiURL, apiToken, bridgeURL)
}

// GetKeptnCredentials retrieves the Keptn Credentials from the "dynatrace" secret
func GetKeptnCredentials() (*KeptnCredentials, error) {
	cm, err := NewDefaultKeptnCredentialsReader()
	if err != nil {
		return nil, err
	}
	return cm.GetKeptnCredentials()
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

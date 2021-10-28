package credentials

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
)

const dynatraceSecretName = "dynatrace"

//go:generate moq --skip-ensure -pkg credentials_mock -out ./mock/keptn_credentials_provider_mock.go . KeptnCredentialsProvider
type KeptnCredentialsProvider interface {
	GetKeptnCredentials() (*KeptnCredentials, error)
}

type KeptnCredentialsReader struct {
	secretReader              *K8sSecretReader
	environmentVariableReader *env.OSEnvironmentVariableReader
}

func NewKeptnCredentialsReader(sr *K8sSecretReader) *KeptnCredentialsReader {
	return &KeptnCredentialsReader{secretReader: sr, environmentVariableReader: env.NewOSEnvironmentVariableReader()}
}

func NewDefaultKeptnCredentialsReader() (*KeptnCredentialsReader, error) {
	sr, err := NewDefaultK8sSecretReader()
	if err != nil {
		return nil, fmt.Errorf("could not initialize KeptnCredentialsReader: %w", err)
	}

	return &KeptnCredentialsReader{
		secretReader:              sr,
		environmentVariableReader: env.NewOSEnvironmentVariableReader(),
	}, nil
}

func (cr *KeptnCredentialsReader) GetKeptnCredentials() (*KeptnCredentials, error) {
	apiURL, err := cr.readSecretWithEnvironmentVariableFallback(keptnAPIURLName)
	if err != nil {
		return nil, err
	}

	apiToken, err := cr.readSecretWithEnvironmentVariableFallback(keptnAPITokenName)
	if err != nil {
		return nil, err
	}

	bridgeURL, err := cr.secretReader.ReadSecret(dynatraceSecretName, keptnBridgeURLName)
	if err != nil {
		bridgeURL, _ = cr.environmentVariableReader.Read(keptnBridgeURLName)
	}

	return NewKeptnCredentials(apiURL, apiToken, bridgeURL)
}

func (cr *KeptnCredentialsReader) readSecretWithEnvironmentVariableFallback(secretName string) (string, error) {
	val, err := cr.secretReader.ReadSecret(dynatraceSecretName, secretName)
	if err == nil {
		return val, nil
	}

	val, found := cr.environmentVariableReader.Read(secretName)
	if found {
		return val, nil
	}

	return "", fmt.Errorf("key %s was not found in secret \"%s\" or environment variables: %w", secretName, dynatraceSecretName, err)
}

// GetKeptnCredentials retrieves the Keptn Credentials from the "dynatrace" secret
func GetKeptnCredentials() (*KeptnCredentials, error) {
	cr, err := NewDefaultKeptnCredentialsReader()
	if err != nil {
		return nil, err
	}
	return cr.GetKeptnCredentials()
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
		return fmt.Errorf("could not authenticate at Keptn API: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid Keptn API Token: received 401 - Unauthorized from %s", keptnAuthURL)
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received unexpected response from %s: %d", keptnAuthURL, resp.StatusCode)
	}
	return nil
}

package credentials

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
)

const dynatraceSecretName = "dynatrace"

const keptnAPIURLKey = "KEPTN_API_URL"
const keptnAPITokenKey = "KEPTN_API_TOKEN"
const keptnBridgeURLKey = "KEPTN_BRIDGE_URL"

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
	apiURL, err := cr.readSecretWithEnvironmentVariableFallback(keptnAPIURLKey)
	if err != nil {
		return nil, err
	}

	apiToken, err := cr.readSecretWithEnvironmentVariableFallback(keptnAPITokenKey)
	if err != nil {
		return nil, err
	}

	bridgeURL, err := cr.secretReader.ReadSecret(dynatraceSecretName, keptnBridgeURLKey)
	if err != nil {
		bridgeURL, _ = cr.environmentVariableReader.Read(keptnBridgeURLKey)
	}

	return NewKeptnCredentials(apiURL, apiToken, bridgeURL)
}

func (cr *KeptnCredentialsReader) readSecretWithEnvironmentVariableFallback(key string) (string, error) {
	val, err := cr.secretReader.ReadSecret(dynatraceSecretName, key)
	if err == nil {
		return val, nil
	}

	val, found := cr.environmentVariableReader.Read(key)
	if found {
		return val, nil
	}

	return "", fmt.Errorf("key \"%s\" was not found in secret \"%s\" or environment variables: %w", key, dynatraceSecretName, err)
}

// GetKeptnCredentials retrieves the Keptn Credentials from the "dynatrace" secret
func GetKeptnCredentials() (*KeptnCredentials, error) {
	cr, err := NewDefaultKeptnCredentialsReader()
	if err != nil {
		return nil, err
	}
	return cr.GetKeptnCredentials()
}

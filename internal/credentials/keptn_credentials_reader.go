package credentials

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
)

const dynatraceSecretName = "dynatrace"

const keptnAPIURLKey = "KEPTN_API_URL"
const keptnAPITokenKey = "KEPTN_API_TOKEN"
const keptnBridgeURLKey = "KEPTN_BRIDGE_URL"

// KeptnCredentialsProvider allows Keptn credentials to be read.
type KeptnCredentialsProvider interface {
	// GetKeptnCredentials gets Keptn credentials or returns an error.
	GetKeptnCredentials(ctx context.Context) (*KeptnCredentials, error)
}

// KeptnCredentialsReader is an implementation of KeptnCredentialsProvider that reads from a K8s secret or environment variables.
type KeptnCredentialsReader struct {
	secretReader              *K8sSecretReader
	environmentVariableReader *env.OSEnvironmentVariableReader
}

// NewKeptnCredentialsReader creates a new KeptnCredentialsReader.
func NewKeptnCredentialsReader(sr *K8sSecretReader) *KeptnCredentialsReader {
	return &KeptnCredentialsReader{secretReader: sr, environmentVariableReader: env.NewOSEnvironmentVariableReader()}
}

// NewDefaultKeptnCredentialsReader creates a new KeptnCredentialsReader using a default K8sSecretReader or returns an error.
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

// GetKeptnCredentials gets Keptn credentials or returns an error.
func (cr *KeptnCredentialsReader) GetKeptnCredentials(ctx context.Context) (*KeptnCredentials, error) {
	apiURL, err := cr.readSecretWithEnvironmentVariableFallback(ctx, keptnAPIURLKey)
	if err != nil {
		return nil, err
	}

	apiToken, err := cr.readSecretWithEnvironmentVariableFallback(ctx, keptnAPITokenKey)
	if err != nil {
		return nil, err
	}

	bridgeURL, err := cr.secretReader.ReadSecret(ctx, dynatraceSecretName, keptnBridgeURLKey)
	if err != nil {
		bridgeURL, _ = cr.environmentVariableReader.Read(keptnBridgeURLKey)
	}

	return NewKeptnCredentials(apiURL, apiToken, bridgeURL)
}

func (cr *KeptnCredentialsReader) readSecretWithEnvironmentVariableFallback(ctx context.Context, key string) (string, error) {
	val, err := cr.secretReader.ReadSecret(ctx, dynatraceSecretName, key)
	if err == nil {
		return val, nil
	}

	val, found := cr.environmentVariableReader.Read(key)
	if found {
		return val, nil
	}

	return "", fmt.Errorf("key \"%s\" was not found in secret \"%s\" or environment variables: %w", key, dynatraceSecretName, err)
}

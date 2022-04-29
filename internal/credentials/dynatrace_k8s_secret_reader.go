package credentials

import (
	"context"
	"fmt"
)

const dynatraceTenantKey = "DT_TENANT"
const dynatraceAPITokenKey = "DT_API_TOKEN"

// DynatraceCredentialsProvider allows Dynatrace credentials to be read.
type DynatraceCredentialsProvider interface {
	// GetDynatraceCredentials gets Dynatrace credentials from the secret with the specified name or returns an error.
	GetDynatraceCredentials(ctx context.Context, secretName string) (*DynatraceCredentials, error)
}

// DynatraceK8sSecretReader is an implementation of DynatraceCredentialsProvider that reads from a K8sSecretReader.
type DynatraceK8sSecretReader struct {
	secretReader *K8sSecretReader
}

// NewDynatraceK8sSecretReader creates a new DynatraceK8sSecretReader.
func NewDynatraceK8sSecretReader(sr *K8sSecretReader) *DynatraceK8sSecretReader {
	return &DynatraceK8sSecretReader{secretReader: sr}
}

// NewDefaultDynatraceK8sSecretReader creates a new DynatraceK8sSecretReader that reads from a default K8sSecretReader.
func NewDefaultDynatraceK8sSecretReader() (*DynatraceK8sSecretReader, error) {
	sr, err := NewDefaultK8sSecretReader()
	if err != nil {
		return nil, fmt.Errorf("could not initialize DynatraceK8sSecretReader: %w", err)
	}
	return NewDynatraceK8sSecretReader(sr), nil
}

// GetDynatraceCredentials gets Dynatrace credentials from the secret with the specified name or returns an error.
func (cr *DynatraceK8sSecretReader) GetDynatraceCredentials(ctx context.Context, secretName string) (*DynatraceCredentials, error) {
	tenant, err := cr.secretReader.ReadSecret(ctx, secretName, dynatraceTenantKey)
	if err != nil {
		return nil, err
	}

	apiToken, err := cr.secretReader.ReadSecret(ctx, secretName, dynatraceAPITokenKey)
	if err != nil {
		return nil, err
	}

	return NewDynatraceCredentials(tenant, apiToken)
}

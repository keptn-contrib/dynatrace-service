package credentials

import (
	"fmt"
)

const dynatraceTenantKey = "DT_TENANT"
const dynatraceAPITokenKey = "DT_API_TOKEN"

//go:generate moq --skip-ensure -pkg credentials_mock -out ./mock/dynatrace_credentials_provider_mock.go . DynatraceCredentialsProvider
type DynatraceCredentialsProvider interface {
	GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error)
}

type DynatraceK8sSecretReader struct {
	secretReader *K8sSecretReader
}

func NewDynatraceK8sSecretReader(sr *K8sSecretReader) *DynatraceK8sSecretReader {
	return &DynatraceK8sSecretReader{secretReader: sr}
}

func NewDefaultDynatraceK8sSecretReader() (*DynatraceK8sSecretReader, error) {
	sr, err := NewDefaultK8sSecretReader()
	if err != nil {
		return nil, fmt.Errorf("could not initialize DynatraceK8sSecretReader: %w", err)
	}
	return &DynatraceK8sSecretReader{secretReader: sr}, nil
}

func (cr *DynatraceK8sSecretReader) GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error) {
	tenant, err := cr.secretReader.ReadSecret(secretName, dynatraceTenantKey)
	if err != nil {
		return nil, err
	}

	apiToken, err := cr.secretReader.ReadSecret(secretName, dynatraceAPITokenKey)
	if err != nil {
		return nil, err
	}

	return NewDynatraceCredentials(tenant, apiToken)
}

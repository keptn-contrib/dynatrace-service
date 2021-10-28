package credentials

import (
	"fmt"
)

const dynatraceTenantSecretName = "DT_TENANT"
const dynatraceAPITokenSecretName = "DT_API_TOKEN"
const keptnAPIURLName = "KEPTN_API_URL"
const keptnAPITokenName = "KEPTN_API_TOKEN"
const keptnBridgeURLName = "KEPTN_BRIDGE_URL"

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
	tenant, err := cr.secretReader.ReadSecret(secretName, dynatraceTenantSecretName)
	if err != nil {
		return nil, fmt.Errorf("key %s was not found in secret \"%s\": %w", dynatraceTenantSecretName, secretName, err)
	}

	apiToken, err := cr.secretReader.ReadSecret(secretName, dynatraceAPITokenSecretName)
	if err != nil {
		return nil, fmt.Errorf("key %s was not found in secret \"%s\": %w", dynatraceAPITokenSecretName, secretName, err)
	}

	return NewDynatraceCredentials(tenant, apiToken)
}

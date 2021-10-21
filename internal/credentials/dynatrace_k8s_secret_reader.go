package credentials

import (
	"fmt"
)

const dynatraceSecretName = "dynatrace"

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
	SecretReader *K8sSecretReader
}

func NewDynatraceK8sSecretReader(sr *K8sSecretReader) *DynatraceK8sSecretReader {
	return &DynatraceK8sSecretReader{SecretReader: sr}
}

func NewDefaultDynatraceK8sSecretReader() (*DynatraceK8sSecretReader, error) {
	sr, err := NewDefaultK8sSecretReader()
	if err != nil {
		return nil, fmt.Errorf("could not initialize DynatraceK8sSecretReader: %s", err.Error())
	}
	return &DynatraceK8sSecretReader{SecretReader: sr}, nil
}

func (cm *DynatraceK8sSecretReader) GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error) {
	dtTenant, err := cm.SecretReader.ReadSecret(secretName, dynatraceTenantSecretName)
	if err != nil {
		return nil, fmt.Errorf("key %s was not found in secret \"%s\"", dynatraceTenantSecretName, secretName)
	}

	dtAPIToken, err := cm.SecretReader.ReadSecret(secretName, dynatraceAPITokenSecretName)
	if err != nil {
		return nil, fmt.Errorf("key %s was not found in secret \"%s\"", dynatraceAPITokenSecretName, secretName)
	}

	return NewDynatraceCredentials(dtTenant, dtAPIToken)
}

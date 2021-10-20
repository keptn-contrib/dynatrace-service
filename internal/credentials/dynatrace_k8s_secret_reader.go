package credentials

import (
	"fmt"
	"strings"
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

func NewDynatraceK8sSecretReader(secretReader *K8sSecretReader) (*DynatraceK8sSecretReader, error) {
	cm := &DynatraceK8sSecretReader{}
	if secretReader == nil {
		sr, err := NewK8sSecretReader(nil)
		if err != nil {
			return nil, fmt.Errorf("could not initialize DynatraceK8sSecretReader: %s", err.Error())
		}
		secretReader = sr
	}
	cm.SecretReader = secretReader
	return cm, nil
}

func (cm *DynatraceK8sSecretReader) GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error) {
	dtTenant, err := cm.SecretReader.ReadSecret(secretName, namespace, dynatraceTenantSecretName)
	if err != nil {
		return nil, fmt.Errorf("key %s was not found in secret \"%s\"", dynatraceTenantSecretName, secretName)
	}

	dtAPIToken, err := cm.SecretReader.ReadSecret(secretName, namespace, dynatraceAPITokenSecretName)
	if err != nil {
		return nil, fmt.Errorf("key %s was not found in secret \"%s\"", dynatraceAPITokenSecretName, secretName)
	}

	dtAPIToken = strings.TrimSpace(dtAPIToken)
	return NewDynatraceCredentials(dtTenant, dtAPIToken)
}

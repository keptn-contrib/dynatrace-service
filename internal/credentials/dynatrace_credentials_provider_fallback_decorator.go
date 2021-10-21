package credentials

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type DynatraceCredentialsProviderFallbackDecorator struct {
	credentialsProvider DynatraceCredentialsProvider
	fallbackSecretNames []string
	secretName          string
}

func NewCredentialManagerFallbackDecorator(cm DynatraceCredentialsProvider, secretNames []string) *DynatraceCredentialsProviderFallbackDecorator {
	return &DynatraceCredentialsProviderFallbackDecorator{
		credentialsProvider: cm,
		fallbackSecretNames: secretNames,
	}
}

func NewCredentialManagerDefaultFallbackDecorator(cm DynatraceCredentialsProvider) *DynatraceCredentialsProviderFallbackDecorator {
	return NewCredentialManagerFallbackDecorator(cm, []string{"dynatrace"})
}

func NewCredentialManagerSLIServiceFallbackDecorator(cm DynatraceCredentialsProvider, project string) *DynatraceCredentialsProviderFallbackDecorator {
	return NewCredentialManagerFallbackDecorator(cm, []string{fmt.Sprintf("dynatrace-credentials-%s", project), "dynatrace-credentials", "dynatrace"})
}

func (cm *DynatraceCredentialsProviderFallbackDecorator) GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error) {
	secrets := []string{secretName}
	secrets = append(secrets, cm.fallbackSecretNames...)

	// let's see whether we are fine with the given secret name first, if not, we will try all our fallback secret names
	for _, secret := range secrets {
		if secret == "" {
			continue
		}

		dtCredentials, err := cm.credentialsProvider.GetDynatraceCredentials(secret)
		if err == nil && dtCredentials != nil {
			log.WithFields(
				log.Fields{
					"secret": secret,
					"tenant": dtCredentials.GetTenant(),
				}).Info("Found secret with credentials")
			cm.secretName = secret
			return dtCredentials, nil
		}
	}

	return nil, fmt.Errorf("could not find any Dynatrace specific secrets with the following names: %s", strings.Join(secrets, ","))
}

func (cm *DynatraceCredentialsProviderFallbackDecorator) GetSecretName() string {
	return cm.secretName
}

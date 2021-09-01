package credentials

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type CredentialManagerFallbackDecorator struct {
	credentialManager   CredentialManagerInterface
	fallbackSecretNames []string
	secretName          string
}

func NewCredentialManagerFallbackDecorator(cm CredentialManagerInterface, secretNames []string) *CredentialManagerFallbackDecorator {
	return &CredentialManagerFallbackDecorator{
		credentialManager:   cm,
		fallbackSecretNames: secretNames,
	}
}

func NewCredentialManagerDefaultFallbackDecorator(cm CredentialManagerInterface) *CredentialManagerFallbackDecorator {
	return NewCredentialManagerFallbackDecorator(cm, []string{"dynatrace"})
}

func NewCredentialManagerSLIServiceFallbackDecorator(cm CredentialManagerInterface, project string) *CredentialManagerFallbackDecorator {
	return NewCredentialManagerFallbackDecorator(cm, []string{fmt.Sprintf("dynatrace-credentials-%s", project), "dynatrace-credentials", "dynatrace"})
}

func (cm *CredentialManagerFallbackDecorator) GetDynatraceCredentials(secretName string) (*DTCredentials, error) {
	secrets := []string{secretName}
	secrets = append(secrets, cm.fallbackSecretNames...)

	// let's see whether we are fine with the given secret name first, if not, we will try all our fallback secret names
	for _, secret := range secrets {
		if secret == "" {
			continue
		}

		dtCredentials, err := cm.credentialManager.GetDynatraceCredentials(secret)
		if err == nil && dtCredentials != nil {
			log.WithFields(
				log.Fields{
					"secret": secret,
					"tenant": dtCredentials.Tenant,
				}).Info("Found secret with credentials")
			cm.secretName = secret
			return dtCredentials, nil
		}
	}

	return nil, fmt.Errorf("could not find any Dynatrace specific secrets with the following names: %s", strings.Join(secrets, ","))
}

func (cm *CredentialManagerFallbackDecorator) GetKeptnAPICredentials() (*KeptnAPICredentials, error) {
	return cm.credentialManager.GetKeptnAPICredentials()
}

func (cm *CredentialManagerFallbackDecorator) GetSecretName() string {
	return cm.secretName
}

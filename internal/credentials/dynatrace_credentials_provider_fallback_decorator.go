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

func NewCredentialsProviderFallbackDecorator(cp DynatraceCredentialsProvider, secretNames []string) *DynatraceCredentialsProviderFallbackDecorator {
	return &DynatraceCredentialsProviderFallbackDecorator{
		credentialsProvider: cp,
		fallbackSecretNames: secretNames,
	}
}

func NewDefaultCredentialsProviderFallbackDecorator(cp DynatraceCredentialsProvider) *DynatraceCredentialsProviderFallbackDecorator {
	return NewCredentialsProviderFallbackDecorator(cp, []string{"dynatrace"})
}

func NewCredentialsProviderSLIServiceFallbackDecorator(cp DynatraceCredentialsProvider, project string) *DynatraceCredentialsProviderFallbackDecorator {
	return NewCredentialsProviderFallbackDecorator(cp, []string{fmt.Sprintf("dynatrace-credentials-%s", project), "dynatrace-credentials", "dynatrace"})
}

func (cp *DynatraceCredentialsProviderFallbackDecorator) GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error) {
	secrets := []string{secretName}
	secrets = append(secrets, cp.fallbackSecretNames...)

	// let's see whether we are fine with the given secret name first, if not, we will try all our fallback secret names
	for _, secret := range secrets {
		if secret == "" {
			continue
		}

		dynatraceCredentials, err := cp.credentialsProvider.GetDynatraceCredentials(secret)
		if err == nil && dynatraceCredentials != nil {
			log.WithFields(
				log.Fields{
					"secret": secret,
					"tenant": dynatraceCredentials.GetTenant(),
				}).Info("Found secret with credentials")
			cp.secretName = secret
			return dynatraceCredentials, nil
		}
	}

	return nil, fmt.Errorf("could not find any Dynatrace specific secrets with the following names: %s", strings.Join(secrets, ","))
}

func (cp *DynatraceCredentialsProviderFallbackDecorator) GetSecretName() string {
	return cp.secretName
}

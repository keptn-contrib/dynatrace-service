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
	// if provided, try the provided secret name first, otherwise try the fallback secret names in order
	secretNames := []string{}
	if secretName != "" {
		secretNames = append(secretNames, secretName)
	}
	secretNames = append(secretNames, cp.fallbackSecretNames...)

	var secretsErrorMessageBuilder strings.Builder
	for i, sn := range secretNames {
		dynatraceCredentials, err := cp.credentialsProvider.GetDynatraceCredentials(sn)
		if err == nil {
			log.WithFields(
				log.Fields{
					"secret": sn,
					"tenant": dynatraceCredentials.GetTenant(),
				}).Info("Found secret with credentials")
			cp.secretName = sn
			return dynatraceCredentials, nil
		} else {
			if i > 0 {
				secretsErrorMessageBuilder.WriteString(", ")
			}
			secretsErrorMessageBuilder.WriteString(sn)
			secretsErrorMessageBuilder.WriteString(" (")
			secretsErrorMessageBuilder.WriteString(err.Error())
			secretsErrorMessageBuilder.WriteString(")")
		}
	}

	cp.secretName = ""
	return nil, fmt.Errorf("could not read the following Dynatrace secrets: %s", secretsErrorMessageBuilder.String())
}

func (cp *DynatraceCredentialsProviderFallbackDecorator) GetSecretName() string {
	return cp.secretName
}

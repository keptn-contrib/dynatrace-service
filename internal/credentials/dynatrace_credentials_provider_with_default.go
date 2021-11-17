package credentials

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type DynatraceCredentialsProviderWithDefault struct {
	credentialsProvider DynatraceCredentialsProvider
	secretName          string
}

func NewDynatraceCredentialsProviderWithDefault(cp DynatraceCredentialsProvider) *DynatraceCredentialsProviderWithDefault {
	return &DynatraceCredentialsProviderWithDefault{
		credentialsProvider: cp,
	}
}

// GetDynatraceCredentials tries to get the dynatrace credentials using the specified secret name. If empty, a default secret name "dynatrace" is used.
func (cp *DynatraceCredentialsProviderWithDefault) GetDynatraceCredentials(secretName string) (*DynatraceCredentials, error) {
	if secretName == "" {
		secretName = "dynatrace"
	}
	dynatraceCredentials, err := cp.credentialsProvider.GetDynatraceCredentials(secretName)
	if err != nil {
		cp.secretName = ""
		return nil, fmt.Errorf("could not get credentials from secret \"%s\": %w", secretName, err)
	}

	log.WithFields(
		log.Fields{
			"secret": secretName,
			"tenant": dynatraceCredentials.GetTenant(),
		}).Info("Found secret with credentials")
	cp.secretName = secretName
	return dynatraceCredentials, nil
}

// GetSecretName returns the name of the secret used for the getting the credentials. If an error occured this will be empty.
func (cp *DynatraceCredentialsProviderWithDefault) GetSecretName() string {
	return cp.secretName
}

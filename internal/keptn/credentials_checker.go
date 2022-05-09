package keptn

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	api "github.com/keptn/go-utils/pkg/api/utils"
)

// CredentialsCheckerInterface validates Keptn credentials.
type CredentialsCheckerInterface interface {
	// CheckCredentials checks if the provided credentials are valid.
	CheckCredentials(keptnCredentials credentials.KeptnCredentials) error
}

// CredentialsChecker is an implementation of ConnectionCheckerInterface.
type CredentialsChecker struct {
}

// NewDefaultCredentialsChecker creates a new CredentialsChecker.
func NewDefaultCredentialsChecker() *CredentialsChecker {
	return &CredentialsChecker{}
}

// CheckCredentials checks the provided credentials and returns an error if they are invalid.
func (c *CredentialsChecker) CheckCredentials(keptnCredentials credentials.KeptnCredentials) error {
	apiSet, err := api.New(keptnCredentials.GetAPIURL(), api.WithAuthToken(keptnCredentials.GetAPIToken()))
	if err != nil {
		return fmt.Errorf("error creating Keptn API set: %w", err)
	}

	_, mErr := apiSet.AuthV1().Authenticate()
	if mErr != nil {
		return fmt.Errorf("error authenticating Keptn connection: %s", mErr.GetMessage())
	}

	return nil
}

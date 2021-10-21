package credentials

import (
	"errors"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/url"
)

type KeptnCredentials struct {
	apiURL   string
	apiToken string
}

func NewKeptnCredentials(apiURL string, apiToken string) (*KeptnCredentials, error) {
	apiURL, err := url.MakeCleanURL(apiURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create keptn credentials: %v", err)
	}

	err = validateKeptnAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("cannot create keptn credentials: %v", err)
	}

	return &KeptnCredentials{apiURL: apiURL, apiToken: apiToken}, nil
}

func (c KeptnCredentials) GetAPIURL() string {
	return c.apiURL
}

func (c KeptnCredentials) GetAPIToken() string {
	return c.apiToken
}

func validateKeptnAPIToken(apiToken string) error {
	if apiToken == "" {
		return errors.New("Keptn API token cannot be empty")
	}
	return nil
}

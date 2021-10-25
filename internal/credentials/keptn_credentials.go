package credentials

import (
	"errors"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/url"
)

type KeptnCredentials struct {
	apiURL   string
	apiToken string
}

func NewKeptnCredentials(apiURL string, apiToken string) (*KeptnCredentials, error) {
	apiURL, err := url.CleanURL(apiURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create Keptn credentials: %v", err)
	}

	apiToken, err = cleanKeptnAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("cannot create Keptn credentials: %v", err)
	}

	return &KeptnCredentials{apiURL: apiURL, apiToken: apiToken}, nil
}

func (c *KeptnCredentials) GetAPIURL() string {
	return c.apiURL
}

func (c *KeptnCredentials) GetAPIToken() string {
	return c.apiToken
}

func cleanKeptnAPIToken(apiToken string) (string, error) {
	apiToken = strings.TrimSpace(apiToken)

	if apiToken == "" {
		return "", errors.New("Keptn API token cannot be empty")
	}
	return apiToken, nil
}

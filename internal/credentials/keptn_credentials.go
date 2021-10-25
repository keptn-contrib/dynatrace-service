package credentials

import (
	"errors"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/url"
)

type KeptnCredentials struct {
	apiURL    string
	apiToken  string
	bridgeURL string
}

func NewKeptnCredentials(apiURL string, apiToken string, bridgeURL string) (*KeptnCredentials, error) {
	apiURL, err := url.CleanURL(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Keptn API URL: %v", err)
	}

	apiToken, err = cleanKeptnAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("invalid Keptn API token: %v", err)
	}

	if bridgeURL != "" {
		bridgeURL, err = url.CleanURL(bridgeURL)
		if err != nil {
			return nil, fmt.Errorf("invalid Keptn bridge URL: %v", err)
		}
	}

	return &KeptnCredentials{apiURL: apiURL, apiToken: apiToken, bridgeURL: bridgeURL}, nil
}

func (c *KeptnCredentials) GetAPIURL() string {
	return c.apiURL
}

func (c *KeptnCredentials) GetAPIToken() string {
	return c.apiToken
}

func (c *KeptnCredentials) GetBridgeURL() string {
	return c.bridgeURL
}

func cleanKeptnAPIToken(apiToken string) (string, error) {
	apiToken = strings.TrimSpace(apiToken)

	if apiToken == "" {
		return "", errors.New("token cannot be empty")
	}
	return apiToken, nil
}

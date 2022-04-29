package credentials

import (
	"errors"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/url"
)

// KeptnCredentials represents Keptn credentials.
type KeptnCredentials struct {
	apiURL    string
	apiToken  string
	bridgeURL string
}

// NewKeptnCredentials creates new Keptn credentials using the specified API URL and token, and bridge URL or returns an error.
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

// GetAPIURL gets the API URL.
func (c *KeptnCredentials) GetAPIURL() string {
	return c.apiURL
}

// GetAPIToken gets the API token.
func (c *KeptnCredentials) GetAPIToken() string {
	return c.apiToken
}

// GetBridgeURL gets the bridge URL.
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

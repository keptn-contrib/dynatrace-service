package common

import (
	"crypto/tls"
	"errors"
	"fmt"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/http"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// RunLocal is true if the "ENV"-environment variable is set to local
var RunLocal = os.Getenv("ENV") == "local"

// RunLocalTest is true if the "ENV"-environment variable is set to localtest
var RunLocalTest = os.Getenv("ENV") == "localtest"

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	if RunLocal || RunLocalTest {
		return nil, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

/**
 * Returns the endpoint to the configuration-service
 */
func GetConfigurationServiceURL() string {
	if os.Getenv("CONFIGURATION_SERVICE_URL") != "" {
		return os.Getenv("CONFIGURATION_SERVICE_URL")
	}
	return "configuration-service:8080"
}

// KeptnCredentials contains the credentials for the Keptn API
type KeptnCredentials struct {
	ApiURL   string
	ApiToken string
}

// GetKeptnCredentials generates the Keptn Credentials from the environment variables KEPTN_API_URL and KEPTN_API_TOKEN
func GetKeptnCredentials() (*KeptnCredentials, error) {

	keptnCreds := &KeptnCredentials{}

	keptnCreds.ApiURL = os.Getenv("KEPTN_API_URL")
	keptnCreds.ApiToken = os.Getenv("KEPTN_API_TOKEN")

	if keptnCreds.ApiURL == "" || keptnCreds.ApiToken == "" {
		return nil, errors.New("no Keptn API credentials available. please provide them in the KEPTN_API_URL and KEPTN_API_TOKEN environment variables")
	}

	if strings.HasPrefix(keptnCreds.ApiURL, "http://") {
		return keptnCreds, nil
	}

	// ensure that apiURL uses https if no other protocol has explicitly been specified
	keptnCreds.ApiURL = strings.TrimPrefix(keptnCreds.ApiURL, "https://")
	keptnCreds.ApiURL = "https://" + keptnCreds.ApiURL

	return keptnCreds, nil
}

// CheckKeptnConnection verifies wether a connection to the Keptn API can be established
func CheckKeptnConnection(keptnCredentials *KeptnCredentials) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest(http.MethodGet, keptnCredentials.ApiURL+"/v1/auth", nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", keptnCredentials.ApiToken)

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not authenticate at Keptn API: " + err.Error())
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid Keptn API Token: received 401 - Unauthorized from " + keptnCredentials.ApiURL + "/v1/auth")
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("received unexpected response from "+keptnCredentials.ApiURL+"/v1/auth: %d", resp.StatusCode))
	}
	return nil
}

// GetKeptnBridgeURL returns the bridge URL
func GetKeptnBridgeURL() (string, error) {
	url := os.Getenv("KEPTN_BRIDGE_URL")

	if url == "" {
		return "", errors.New("no bridge URL specified in KEPTN_BRIDGE_URL env var")
	}

	if strings.HasPrefix(url, "http://") {
		return url, nil
	}

	// ensure that apiURL uses https if no other protocol has explicitly been specified
	url = strings.TrimPrefix(url, "https://")
	url = "https://" + url

	return url, nil
}

// DecodeKeptnEventData decodes a Keptn events to the specified object
func DecodeKeptnEventData(in, out interface{}) error {
	return keptnv2.Decode(in, out)
}

package common

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

// This is the label name for the Problem URL label
const PROBLEMURL_LABEL = "Problem URL"
const KEPTNSBRIDGE_LABEL = "Keptns Bridge"

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

/**
 * Finds the Problem ID that is associated with this Keptn Workflow
 * It first parses it from Problem URL label - if it cant be found there it will look for the Initial Problem Open Event and gets the ID from there!
 */
func FindProblemIDForEvent(keptnHandler *keptn.Keptn, labels map[string]string) (string, error) {

	// Step 1 - see if we have a Problem Url in the labels
	// iterate through the labels and find Problem URL
	for labelName, labelValue := range labels {
		if labelName == PROBLEMURL_LABEL {
			// the value should be of form https://dynatracetenant/#problems/problemdetails;pid=8485558334848276629_1604413609638V2
			// so - lets get the last part after pid=

			ix := strings.LastIndex(labelValue, ";pid=")
			if ix > 0 {
				return labelValue[ix+5:], nil
			}
		}
	}

	// Step 2 - lets see if we have a ProblemOpenEvent for this KeptnContext - if so - we try to extract the Problem ID
	eventHandler := keptnapi.NewEventHandler(os.Getenv("DATASTORE"))

	events, errObj := eventHandler.GetEvents(&keptnapi.EventFilter{
		Project:      keptnHandler.KeptnBase.Project,
		EventType:    keptn.ProblemOpenEventType,
		KeptnContext: keptnHandler.KeptnContext,
	})

	if errObj != nil {
		msg := "cannot send DT problem comment: Could not retrieve problem.open event for incoming event: " + *errObj.Message
		return "", errors.New(msg)
	}

	if len(events) == 0 {
		msg := "cannot send DT problem comment: Could not retrieve problem.open event for incoming event: no events returned"
		return "", errors.New(msg)
	}

	problemOpenEvent := &keptn.ProblemEventData{}
	err := mapstructure.Decode(events[0].Data, problemOpenEvent)

	if err != nil {
		msg := "could not decode problem.open event: " + err.Error()
		return "", errors.New(msg)
	}

	if problemOpenEvent.PID == "" {
		return "", errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	return problemOpenEvent.PID, nil
}

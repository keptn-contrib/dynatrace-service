package common

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/http"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// This is the label name for the Problem URL label
const PROBLEMURL_LABEL = "Problem URL"
const KEPTNSBRIDGE_LABEL = "Keptns Bridge"

const shipyardController = "SHIPYARD_CONTROLLER"
const configurationService = "CONFIGURATION_SERVICE"
const defaultShipyardControllerURL = "http://shipyard-controller:8080"
const defaultConfigurationServiceURL = "http://configuration-service:8080"

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

// GetConfigurationServiceURL Returns the endpoint to the configuration-service
func GetConfigurationServiceURL() string {
	return getKeptnServiceURL(configurationService, defaultConfigurationServiceURL)
}

// GetShipyardControllerURL Returns the endpoint to the shipyard-controller
func GetShipyardControllerURL() string {
	return getKeptnServiceURL(shipyardController, defaultShipyardControllerURL)
}

func getKeptnServiceURL(servicename, defaultURL string) string {
	var baseURL string
	url, err := keptncommon.GetServiceEndpoint(servicename)
	if err == nil {
		baseURL = url.String()
	} else {
		baseURL = defaultURL
	}
	return baseURL
}

// KeptnCredentials contains the credentials for the Keptn API
type KeptnCredentials struct {
	ApiURL   string
	ApiToken string
}

// CheckKeptnConnection verifies wether a connection to the Keptn API can be established
func CheckKeptnConnection(keptnCredentials *credentials.KeptnAPICredentials) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest(http.MethodGet, keptnCredentials.APIURL+"/v1/auth", nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", keptnCredentials.APIToken)

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not authenticate at Keptn API: " + err.Error())
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid Keptn API Token: received 401 - Unauthorized from " + keptnCredentials.APIURL + "/v1/auth")
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("received unexpected response from "+keptnCredentials.APIURL+"/v1/auth: %d", resp.StatusCode))
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
func FindProblemIDForEvent(keptnHandler *keptnv2.Keptn, labels map[string]string) (string, error) {

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
		Project:      keptnHandler.KeptnBase.Event.GetProject(),
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
	err := keptnv2.Decode(events[0].Data, problemOpenEvent)

	if err != nil {
		msg := "could not decode problem.open event: " + err.Error()
		return "", errors.New(msg)
	}

	if problemOpenEvent.PID == "" {
		return "", errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	return problemOpenEvent.PID, nil
}

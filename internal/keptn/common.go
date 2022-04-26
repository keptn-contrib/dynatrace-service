package keptn

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn/go-utils/pkg/lib/keptn"
)

const shipyardControllerURLEnvironmentVariableName = "SHIPYARD_CONTROLLER"
const configurationServiceEnvironmentVariableName = "CONFIGURATION_SERVICE"
const datastoreEnvironmentVariableName = "DATASTORE"

const defaultShipyardControllerURL = "http://shipyard-controller:8080"

// getConfigurationServiceURL Returns the endpoint to the configuration-service.
func getConfigurationServiceURL() string {
	return getKeptnServiceURL(configurationServiceEnvironmentVariableName, keptn.ConfigurationServiceURL)
}

// getDatastoreURL Returns the endpoint to the datastore.
func getDatastoreURL() string {
	return getKeptnServiceURL(datastoreEnvironmentVariableName, keptn.DatastoreURL)
}

// getShipyardControllerURL Returns the endpoint to the shipyard-controller.
func getShipyardControllerURL() string {
	return getKeptnServiceURL(shipyardControllerURLEnvironmentVariableName, defaultShipyardControllerURL)
}

func getKeptnServiceURL(serviceName, defaultURL string) string {
	url, err := keptn.GetServiceEndpoint(serviceName)
	if err != nil {
		return defaultURL
	}
	return url.String()
}

// TryGetBridgeURLForKeptnContext gets a backlink to the Keptn Bridge if available or returns "".
func TryGetBridgeURLForKeptnContext(event adapter.EventContentAdapter) string {
	credentials, err := credentials.GetKeptnCredentials()
	if err != nil {
		return ""
	}

	keptnBridgeURL := credentials.GetBridgeURL()
	if keptnBridgeURL == "" {
		return ""
	}

	return keptnBridgeURL + "/trace/" + event.GetShKeptnContext()
}

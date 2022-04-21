package keptn

import "github.com/keptn/go-utils/pkg/lib/keptn"

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

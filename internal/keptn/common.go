package keptn

import "github.com/keptn/go-utils/pkg/lib/keptn"

const shipyardControllerURLEnvironmentVariableName = "SHIPYARD_CONTROLLER"
const configurationServiceEnvironmentVariableName = "CONFIGURATION_SERVICE"
const datastoreEnvironmentVariableName = "DATASTORE"

const defaultShipyardControllerURL = "http://shipyard-controller:8080"

// GetConfigurationServiceURL Returns the endpoint to the configuration-service
func GetConfigurationServiceURL() string {
	return getKeptnServiceURL(configurationServiceEnvironmentVariableName, keptn.ConfigurationServiceURL)
}

// GetDatastoreURL Returns the endpoint to the datastore
func GetDatastoreURL() string {
	return getKeptnServiceURL(datastoreEnvironmentVariableName, keptn.DatastoreURL)
}

// GetShipyardControllerURL Returns the endpoint to the shipyard-controller
func GetShipyardControllerURL() string {
	return getKeptnServiceURL(shipyardControllerURLEnvironmentVariableName, defaultShipyardControllerURL)
}

func getKeptnServiceURL(serviceName, defaultURL string) string {
	url, err := keptn.GetServiceEndpoint(serviceName)
	if err != nil {
		return defaultURL
	}
	return url.String()
}

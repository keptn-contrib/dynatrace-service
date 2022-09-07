package keptn

import (
	"github.com/keptn-contrib/dynatrace-service/internal/env"

	api "github.com/keptn/go-utils/pkg/api/utils"
	v2 "github.com/keptn/go-utils/pkg/api/utils/v2"
)

// GetV1InClusterAPIMappings returns the InClusterAPIMappings.
func GetV1InClusterAPIMappings() api.InClusterAPIMappings {
	mappings := api.InClusterAPIMappings{}
	for k, v := range api.DefaultInClusterAPIMappings {
		mappings[k] = v
	}

	shipyardController := env.GetShipyardController()
	if shipyardController != "" {
		mappings[api.ShipyardController] = shipyardController
	}

	configurationService := env.GetConfigurationService()
	if configurationService != "" {
		mappings[api.ConfigurationService] = configurationService
	}

	datastore := env.GetDatastore()
	if datastore != "" {
		mappings[api.MongoDBDatastore] = datastore
	}

	apiService := env.GetAPIService()
	if apiService != "" {
		mappings[api.ApiService] = apiService
	}

	return mappings
}

// GetV2InClusterAPIMappings returns the InClusterAPIMappings.
func GetV2InClusterAPIMappings() v2.InClusterAPIMappings {
	mappings := v2.InClusterAPIMappings{}
	for k, v := range v2.DefaultInClusterAPIMappings {
		mappings[k] = v
	}

	shipyardController := env.GetShipyardController()
	if shipyardController != "" {
		mappings[v2.ShipyardController] = shipyardController
	}

	configurationService := env.GetConfigurationService()
	if configurationService != "" {
		mappings[v2.ConfigurationService] = configurationService
	}

	datastore := env.GetDatastore()
	if datastore != "" {
		mappings[v2.MongoDBDatastore] = datastore
	}

	apiService := env.GetAPIService()
	if apiService != "" {
		mappings[v2.ApiService] = apiService
	}

	return mappings
}

package keptn

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	api "github.com/keptn/keptn/cp-common/api"
	v2 "github.com/keptn/keptn/cp-common/v2"
)

// TryGetBridgeURLForKeptnContext gets a backlink to the Keptn Bridge if available or returns "".
func TryGetBridgeURLForKeptnContext(ctx context.Context, event adapter.EventContentAdapter) string {
	credentials, err := credentials.GetKeptnCredentials(ctx)
	if err != nil {
		return ""
	}

	keptnBridgeURL := credentials.GetBridgeURL()
	if keptnBridgeURL == "" {
		return ""
	}

	return keptnBridgeURL + "/trace/" + event.GetShKeptnContext()
}

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

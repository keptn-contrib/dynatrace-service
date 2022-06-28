package keptn

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	api "github.com/keptn/go-utils/pkg/api/utils"
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

// GetInClusterAPIMappings returns the InClusterAPIMappings.
func GetInClusterAPIMappings() api.InClusterAPIMappings {
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

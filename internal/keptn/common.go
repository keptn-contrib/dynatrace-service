package keptn

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/env"

	api "github.com/keptn/go-utils/pkg/api/utils"
	v2 "github.com/keptn/go-utils/pkg/api/utils/v2"
)

// TryGetBridgeURLForKeptnContext gets a backlink to the Keptn Bridge if available or returns empty string.
func TryGetBridgeURLForKeptnContext(ctx context.Context, event adapter.EventContentAdapter) string {
	keptnBridgeURL := tryGetBridgeURL(ctx)
	if keptnBridgeURL == "" {
		return ""
	}

	return keptnBridgeURL + "/trace/" + event.GetShKeptnContext()
}

// tryGetBridgeURL gets the Keptn Bridge URL if available or returns empty string.
func tryGetBridgeURL(ctx context.Context) string {
	creds, err := credentials.GetKeptnCredentials(ctx)
	if err != nil {
		return ""
	}

	return creds.GetBridgeURL()
}

// TryGetBridgeURLForEvaluation gets a backlink to the evaluation in Keptn Bridge if available or returns empty string.
func TryGetBridgeURLForEvaluation(ctx context.Context, event adapter.EventContentAdapter) string {
	keptnBridgeURL := tryGetBridgeURL(ctx)
	if keptnBridgeURL == "" {
		return ""
	}

	return fmt.Sprintf("%s/evaluation/%s/%s", keptnBridgeURL, event.GetShKeptnContext(), event.GetStage())
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

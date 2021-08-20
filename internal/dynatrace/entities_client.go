package dynatrace

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const entitiesPath = "/api/v2/entities"

// EntitiesResponse represents the response from Dynatrace entities endpoints
type EntitiesResponse struct {
	TotalCount  int      `json:"totalCount"`
	PageSize    int      `json:"pageSize"`
	NextPageKey string   `json:"nextPageKey"`
	Entities    []Entity `json:"entities"`
}

// Tag represents a tag applied to a Dynatrace entity
type Tag struct {
	Context              string `json:"context"`
	Key                  string `json:"key"`
	StringRepresentation string `json:"stringRepresentation"`
	Value                string `json:"value,omitempty"`
}

// Entity represents a Dynatrace entity
type Entity struct {
	EntityID    string `json:"entityId"`
	DisplayName string `json:"displayName"`
	Tags        []Tag  `json:"tags"`
}

// EntitiesClient is a client for interacting with the Dynatrace entities endpoints
type EntitiesClient struct {
	Client *DynatraceHelper
}

// NewEntitiesClient creates a new EntitiesClient
func NewEntitiesClient(client *DynatraceHelper) *EntitiesClient {
	return &EntitiesClient{
		Client: client,
	}
}

// GetKeptnManagedServices gets all service entities with a keptn_managed and keptn_service tag
func (ec *EntitiesClient) GetKeptnManagedServices() ([]Entity, error) {
	entities := []Entity{}
	nextPageKey := ""

	// TODO 2021-08-20: Investigate if pageSize should be optimized or removed
	pageSize := 50
	for {
		var response string
		var err error

		if nextPageKey == "" {
			response, err = ec.Client.Get(entitiesPath + "entitySelector=type(\"SERVICE\")%20AND%20tag(\"keptn_managed\",\"[Environment]keptn_managed\")%20AND%20tag(\"keptn_service\",\"[Environment]keptn_service\")&fields=+tags&pageSize=" + strconv.FormatInt(int64(pageSize), 10))
		} else {
			response, err = ec.Client.Get(entitiesPath + "?nextPageKey=" + nextPageKey)
		}
		if err != nil {
			return nil, err
		}

		entitiesResponse := &EntitiesResponse{}
		err = json.Unmarshal([]byte(response), entitiesResponse)
		if err != nil {
			return nil, fmt.Errorf("could not deserialize EntitiesResponse: %v", err)
		}

		entities = append(entities, entitiesResponse.Entities...)
		if entitiesResponse.NextPageKey == "" {
			break
		}
		nextPageKey = entitiesResponse.NextPageKey
	}
	return entities, nil
}

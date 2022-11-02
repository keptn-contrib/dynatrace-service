package dynatrace

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
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
	Client ClientInterface
}

// NewEntitiesClient creates a new EntitiesClient
func NewEntitiesClient(client ClientInterface) *EntitiesClient {
	return &EntitiesClient{
		Client: client,
	}
}

// GetKeptnManagedServices gets all service entities with a keptn_managed and keptn_service tag.
func (ec *EntitiesClient) GetKeptnManagedServices(ctx context.Context) ([]Entity, error) {
	entities := []Entity{}
	nextPageKey := ""

	// TODO 2021-08-20: Investigate if pageSize should be optimized or removed
	pageSize := 50
	for {
		var response []byte
		var err error

		if nextPageKey == "" {
			response, err = ec.Client.Get(ctx, entitiesPath+"?"+buildKeptnManagedServicesQueryParams(pageSize))
		} else {
			response, err = ec.Client.Get(ctx, entitiesPath+"?nextPageKey="+nextPageKey)
		}
		if err != nil {
			return nil, err
		}

		entitiesResponse := &EntitiesResponse{}
		err = json.Unmarshal(response, entitiesResponse)
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

// buildKeptnManagedServicesQueryParams generates url encoded query parameters for retrieving services tagged with "keptn_managed" and "keptn_service".
func buildKeptnManagedServicesQueryParams(pageSize int) string {
	query := newQueryParameters()

	query.add("entitySelector", "type(\"SERVICE\") AND tag(\"keptn_managed\",\"[Environment]keptn_managed\") AND tag(\"keptn_service\",\"[Environment]keptn_service\")")
	query.add("fields", "+tags")
	query.add("pageSize", strconv.FormatInt(int64(pageSize), 10))

	return query.encode()
}

type PGIQueryConfig struct {
	Project string
	Stage   string
	Service string
	Version string
	From    time.Time
	To      time.Time
}

// GetAllPGIsForKeptnServices returns all PGIs that belong to a SERVICE entity with tags for `keptn_project`, `keptn_stage` and `keptn_service`
func (ec *EntitiesClient) GetAllPGIsForKeptnServices(ctx context.Context, cfg PGIQueryConfig) ([]string, error) {

	query := newQueryParameters()
	query.add("entitySelector", fmt.Sprintf("type(\"process_group_instance\"),toRelationship.runsOnProcessGroupInstance(type(SERVICE),tag(\"keptn_project:%s\"),tag(\"keptn_stage:%s\"),tag(\"keptn_service:%s\")),releasesVersion(\"%s\")", cfg.Project, cfg.Stage, cfg.Service, cfg.Version))
	query.add("from", common.TimestampToUnixMillisecondsString(cfg.From))
	query.add("to", common.TimestampToUnixMillisecondsString(cfg.To))

	response, err := ec.Client.Get(ctx, entitiesPath+"?"+query.encode())
	if err != nil {
		return nil, err
	}

	entitiesResponse := &EntitiesResponse{}
	err = json.Unmarshal(response, entitiesResponse)
	if err != nil {
		return nil, common.NewUnmarshalJSONError("monitored entities", err)
	}

	var pgis []string
	for _, entity := range entitiesResponse.Entities {
		pgis = append(pgis, entity.EntityID)
	}

	return pgis, nil
}

package dynatrace

import (
	"context"
	"encoding/json"
	"fmt"
)

type ManagementZone struct {
	Name  string    `json:"name"`
	Rules []MZRules `json:"rules"`
}

type MZKey struct {
	Attribute string `json:"attribute"`
}
type MZValue struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}
type MZComparisonInfo struct {
	Type     string  `json:"type"`
	Operator string  `json:"operator"`
	Value    MZValue `json:"value"`
	Negate   bool    `json:"negate"`
}
type MZConditions struct {
	Key            MZKey            `json:"key"`
	ComparisonInfo MZComparisonInfo `json:"comparisonInfo"`
}
type MZRules struct {
	Type             string         `json:"type"`
	Enabled          bool           `json:"enabled"`
	PropagationTypes []string       `json:"propagationTypes"`
	Conditions       []MZConditions `json:"conditions"`
}

const managementZonesPath = "/api/config/v1/managementZones"

type ManagementZones struct {
	values map[string]values
}

func (mz *ManagementZones) GetByName(name string) (values, bool) {
	value, exists := mz.values[name]
	return value, exists
}

func (mz *ManagementZones) Contains(name string) bool {
	_, exists := mz.GetByName(name)
	return exists
}

type ManagementZonesClient struct {
	client ClientInterface
}

func NewManagementZonesClient(client ClientInterface) *ManagementZonesClient {
	return &ManagementZonesClient{
		client: client,
	}
}

// GetAll gets all management zones.
func (mzc *ManagementZonesClient) GetAll(ctx context.Context) (*ManagementZones, error) {
	response, err := mzc.client.Get(ctx, managementZonesPath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve management zones: %v", err)
	}

	mzs := &listResponse{}
	err = json.Unmarshal(response, mzs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse management zones list: %v", err)
	}

	return transformToManagementZones(mzs), nil
}

func transformToManagementZones(response *listResponse) *ManagementZones {
	managementZones := &ManagementZones{
		values: make(map[string]values, len(response.Values)),
	}
	for _, value := range response.Values {
		managementZones.values[value.Name] = value
	}

	return managementZones
}

// Create creates a management zone.
func (mzc *ManagementZonesClient) Create(ctx context.Context, managementZone *ManagementZone) error {
	mzPayload, err := json.Marshal(managementZone)
	if err != nil {
		return fmt.Errorf("failed to marshal management zone for project: %v", err)
	}

	_, err = mzc.client.Post(ctx, managementZonesPath, mzPayload)
	if err != nil {
		return fmt.Errorf("failed to create management zone: %v", err)
	}

	return nil
}

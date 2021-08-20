package dynatrace

import (
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
	values map[string]Values
}

func (mz *ManagementZones) GetByName(name string) (Values, bool) {
	value, exists := mz.values[name]
	return value, exists
}

func (mz *ManagementZones) Contains(name string) bool {
	_, exists := mz.GetByName(name)
	return exists
}

type ManagementZonesClient struct {
	client *DynatraceHelper
}

func NewManagementZonesClient(client *DynatraceHelper) *ManagementZonesClient {
	return &ManagementZonesClient{
		client: client,
	}
}

func (mzc *ManagementZonesClient) GetAll() (*ManagementZones, error) {
	response, err := mzc.client.Get(managementZonesPath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve management zones: %v", err)
	}

	mzs := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(response), mzs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse management zones list: %v", err)
	}

	return transformToManagementZones(mzs), nil
}

func transformToManagementZones(response *DTAPIListResponse) *ManagementZones {
	managementZones := &ManagementZones{
		values: make(map[string]Values, len(response.Values)),
	}
	for _, value := range response.Values {
		managementZones.values[value.Name] = value
	}

	return managementZones
}

func (mzc *ManagementZonesClient) Create(managementZone *ManagementZone) (string, error) {
	mzPayload, err := json.Marshal(managementZone)
	if err != nil {
		return "", fmt.Errorf("failed to marshal management zone for project: %v", err)
	}

	res, err := mzc.client.Post(managementZonesPath, mzPayload)
	if err != nil {
		return "", fmt.Errorf("failed to create management zone: %v", err)
	}

	return res, nil
}

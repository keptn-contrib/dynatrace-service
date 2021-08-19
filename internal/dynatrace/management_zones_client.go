package dynatrace

import (
	"encoding/json"
	"fmt"
)

const managementZonesPath = "/api/config/v1/managementZones"

type ManagementZones struct {
	values map[string]Values
}

func (mz *ManagementZones) GetBy(name string) (Values, bool) {
	value, exists := mz.values[name]
	return value, exists
}

func (mz *ManagementZones) Contains(name string) bool {
	_, exists := mz.GetBy(name)
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

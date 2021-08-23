package dynatrace

import (
	"encoding/json"
	"fmt"
)

const dashboardsPath = "/api/config/v1/dashboards"

type Dashboards struct {
	Dashboards []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Owner string `json:"owner"`
	} `json:"dashboards"`
}

type DynatraceDashboard struct {
	DashboardMetadata DashboardMetadata `json:"dashboardMetadata"`
	Tiles             []Tiles           `json:"tiles"`
}
type SharingDetails struct {
	LinkShared bool `json:"linkShared"`
	Published  bool `json:"published"`
}
type DashboardFilter struct {
	Timeframe      string      `json:"timeframe"`
	ManagementZone interface{} `json:"managementZone"`
}
type DashboardMetadata struct {
	Name            string          `json:"name"`
	Shared          bool            `json:"shared"`
	Owner           string          `json:"owner"`
	SharingDetails  SharingDetails  `json:"sharingDetails"`
	DashboardFilter DashboardFilter `json:"dashboardFilter"`
}
type Bounds struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
type TileFilter struct {
	Timeframe      interface{} `json:"timeframe"`
	ManagementZone interface{} `json:"managementZone"`
}
type ResultMetadata struct {
}
type ChartConfig struct {
	Type           string         `json:"type"`
	Series         []Series       `json:"series"`
	ResultMetadata ResultMetadata `json:"resultMetadata"`
}
type Series struct {
	Metric          string       `json:"metric"`
	Aggregation     string       `json:"aggregation"`
	Percentile      interface{}  `json:"percentile"`
	Type            string       `json:"type"`
	EntityType      string       `json:"entityType"`
	Dimensions      []Dimensions `json:"dimensions"`
	SortAscending   bool         `json:"sortAscending"`
	SortColumn      bool         `json:"sortColumn"`
	AggregationRate string       `json:"aggregationRate"`
}
type Dimensions struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Values          []string `json:"values"`
	EntityDimension bool     `json:"entityDimension"`
}
type FiltersPerEntityType struct {
	Service *EntityFilter `json:"SERVICE,omitempty"`
}
type EntityFilter struct {
	AutoTags []string `json:"AUTO_TAGS,omitempty"`
}
type FilterConfig struct {
	Type                 string               `json:"type"`
	CustomName           string               `json:"customName"`
	DefaultName          string               `json:"defaultName"`
	ChartConfig          ChartConfig          `json:"chartConfig"`
	FiltersPerEntityType FiltersPerEntityType `json:"filtersPerEntityType"`
}
type Tiles struct {
	Name                      string        `json:"name"`
	TileType                  string        `json:"tileType"`
	Configured                bool          `json:"configured"`
	Bounds                    Bounds        `json:"bounds"`
	TileFilter                TileFilter    `json:"tileFilter"`
	FilterConfig              *FilterConfig `json:"filterConfig,omitempty"`
	ChartVisible              bool          `json:"chartVisible,omitempty"`
	AssignedEntities          []string      `json:"assignedEntities,omitempty"`
	ExcludeMaintenanceWindows bool          `json:"excludeMaintenanceWindows,omitempty"`
	Markdown                  string        `json:"markdown,omitempty"`
}

type DashboardsClient struct {
	client *DynatraceHelper
}

func NewDashboardsClient(client *DynatraceHelper) *DashboardsClient {
	return &DashboardsClient{
		client: client,
	}
}

func (dc *DashboardsClient) GetAll() (*Dashboards, error) {
	res, err := dc.client.Get(dashboardsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve list of existing Dynatrace dashboards: %v", err)
	}

	dashboards := &Dashboards{}
	err = json.Unmarshal([]byte(res), dashboards)
	if err != nil {
		err = CheckForUnexpectedHTMLResponseError(err)
		return nil, fmt.Errorf("failed to unmarshal list of existing Dynatrace dashboards: %v", err)
	}

	return dashboards, nil
}

func (dc *DashboardsClient) Create(dashboard *DynatraceDashboard) (string, error) {
	dashboardPayload, err := json.Marshal(dashboard)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal Dynatrace dashboards: %v", err)
	}

	res, err := dc.client.Post(dashboardsPath, dashboardPayload)
	if err != nil {
		return "", fmt.Errorf("failed to create Dynatrace dashboards: %v", err)
	}

	return res, nil
}

func (dc *DashboardsClient) Delete(dashboardID string) (string, error) {
	res, err := dc.client.Delete(dashboardsPath + "/" + dashboardID)
	if err != nil {
		return "", err
	}

	return res, nil
}

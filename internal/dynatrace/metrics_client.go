package dynatrace

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
)

const metricsPath = "/api/v2/metrics"

// MetricDefinition defines the output of /metrics/<metricID>
type MetricDefinition struct {
	MetricID           string   `json:"metricId"`
	DisplayName        string   `json:"displayName"`
	Description        string   `json:"description"`
	Unit               string   `json:"unit"`
	AggregationTypes   []string `json:"aggregationTypes"`
	Transformations    []string `json:"transformations"`
	DefaultAggregation struct {
		Type string `json:"type"`
	} `json:"defaultAggregation"`
	DimensionDefinitions []struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Key         string `json:"key"`
		DisplayName string `json:"displayName"`
	} `json:"dimensionDefinitions"`
	EntityType []string `json:"entityType"`
}

// MetricsClient is a client for interacting with the Dynatrace problems endpoints
type MetricsClient struct {
	client *Client
}

// NewMetricsClient creates a new MetricsClient
func NewMetricsClient(client *Client) *MetricsClient {
	return &MetricsClient{
		client: client,
	}
}

// GetByID calls the Dynatrace API to retrieve MetricDefinition details.
func (mc *MetricsClient) GetByID(metricID string) (*MetricDefinition, error) {
	body, err := mc.client.Get(metricsPath + "/" + metricID)
	if err != nil {
		return nil, err
	}

	var result MetricDefinition
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetByQuery executes the passed Metrics API Call, validates that the call returns data and returns the data set
func (mc *MetricsClient) GetByQuery(metricsQuery string) (*MetricsQueryResult, error) {
	path := metricsPath + "/query?" + metricsQuery
	log.WithField("query", mc.client.DynatraceCreds.Tenant+path).Debug("Final Query")

	body, err := mc.client.Get(path)
	if err != nil {
		return nil, err
	}

	var result MetricsQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Result) == 0 {
		// there are no data points - try again?
		return nil, errors.New("dynatrace Metrics API returned no DataPoints")
	}

	return &result, nil
}

package dynatrace

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
)

// MetricsPath is the base endpoint for Metrics API v2
const MetricsPath = "/api/v2/metrics"

// MetricsQueryPath is the query endpoint for Metrics API v2
const MetricsQueryPath = MetricsPath + "/query"

const (
	fromKey           = "from"
	toKey             = "to"
	metricSelectorKey = "metricSelector"
	resolutionKey     = "resolution"
	entitySelectorKey = "entitySelector"
)

// MetricsClientQueryParameters encapsulates the query parameters for the MetricsClient's GetByQuery method.
type MetricsClientQueryParameters struct {
	query metrics.Query
	from  time.Time
	to    time.Time
}

// NewMetricsClientQueryParameters creates new MetricsClientQueryParameters.
func NewMetricsClientQueryParameters(query metrics.Query, from time.Time, to time.Time) MetricsClientQueryParameters {
	return MetricsClientQueryParameters{
		query: query,
		from:  from,
		to:    to,
	}
}

// encode encodes MetricsClientQueryParameters into a URL-encoded string.
func (q *MetricsClientQueryParameters) encode() string {
	queryParameters := newQueryParameters()
	queryParameters.add(metricSelectorKey, q.query.GetMetricSelector())
	queryParameters.add(fromKey, common.TimestampToUnixMillisecondsString(q.from))
	queryParameters.add(toKey, common.TimestampToUnixMillisecondsString(q.to))
	queryParameters.add(resolutionKey, "Inf")
	if q.query.GetEntitySelector() != "" {
		queryParameters.add(entitySelectorKey, q.query.GetEntitySelector())
	}
	return queryParameters.encode()
}

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
	DimensionDefinitions []DimensionDefinition `json:"dimensionDefinitions"`
	EntityType           []string              `json:"entityType"`
}

type DimensionDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
}

// MetricsQueryResult is struct for /metrics/query
type MetricsQueryResult struct {
	Result []MetricQueryResultValues `json:"result"`
}

type MetricQueryResultValues struct {
	MetricID string                     `json:"metricId"`
	Data     []MetricQueryResultNumbers `json:"data"`
	Warnings []string                   `json:"warnings,omitempty"`
}

type MetricQueryResultNumbers struct {
	Dimensions   []string          `json:"dimensions"`
	DimensionMap map[string]string `json:"dimensionMap,omitempty"`
	Timestamps   []int64           `json:"timestamps"`
	Values       []float64         `json:"values"`
}

// MetricsClient is a client for interacting with the Dynatrace problems endpoints
type MetricsClient struct {
	client ClientInterface
}

// NewMetricsClient creates a new MetricsClient
func NewMetricsClient(client ClientInterface) *MetricsClient {
	return &MetricsClient{
		client: client,
	}
}

// GetByID calls the Dynatrace API to retrieve MetricDefinition details.
func (mc *MetricsClient) GetByID(metricID string) (*MetricDefinition, error) {
	body, err := mc.client.Get(MetricsPath + "/" + metricID)
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
func (mc *MetricsClient) GetByQuery(parameters MetricsClientQueryParameters) (*MetricsQueryResult, error) {
	body, err := mc.client.Get(MetricsQueryPath + "?" + parameters.encode())
	if err != nil {
		return nil, err
	}

	var result MetricsQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Result) == 0 {
		return nil, errors.New("Dynatrace Metrics API returned no datapoints")
	}

	return &result, nil
}

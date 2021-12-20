package dynatrace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const sloPath = "/api/v2/slo"

type SLOResult struct {
	ID                  string  `json:"id"`
	Enabled             bool    `json:"enabled"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	EvaluatedPercentage float64 `json:"evaluatedPercentage"`
	ErrorBudget         float64 `json:"errorBudget"`
	Status              string  `json:"status"`
	Error               string  `json:"error"`
	UseRateMetric       bool    `json:"useRateMetric"`
	MetricRate          string  `json:"metricRate"`
	MetricNumerator     string  `json:"metricNumerator"`
	MetricDenominator   string  `json:"metricDenominator"`
	Target              float64 `json:"target"`
	Warning             float64 `json:"warning"`
	EvaluationType      string  `json:"evaluationType"`
	TimeWindow          string  `json:"timeWindow"`
	Filter              string  `json:"filter"`
}

type SLOClient struct {
	client ClientInterface
}

func NewSLOClient(client ClientInterface) *SLOClient {
	return &SLOClient{
		client: client,
	}
}

// Get calls Dynatrace API to retrieve the values of the Dynatrace SLO for that timeframe
// It returns a SLOResult object on success, an error otherwise
func (c *SLOClient) Get(sloID string, startUnix time.Time, endUnix time.Time) (*SLOResult, error) {
	body, err := c.client.Get(
		fmt.Sprintf("%s/%s?from=%s&to=%s",
			sloPath,
			sloID,
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix)))
	if err != nil {
		return nil, err
	}

	var result SLOResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// for SLO - its also possible that there is an HTTP 200 but there is an error text in the error property!
	// Since Sprint 206 the error property is always there - but - will have the value "NONE" in case there is no actual error retrieving the value
	if result.Error != "NONE" {
		return nil, fmt.Errorf("dynatrace API returned an error: %s", result.Error)
	}

	return &result, nil
}

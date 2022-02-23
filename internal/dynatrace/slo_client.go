package dynatrace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const SLOPath = "/api/v2/slo"

const (
	timeFrameKey = "timeFrame"
)

// SLOClientGetParameters encapsulates the parameters for the SLOClient's Get method.
type SLOClientGetParameters struct {
	sloID string
	from  time.Time
	to    time.Time
}

// NewSLOClientGetParameters creates new SLOClientGetParameters.
func NewSLOClientGetParameters(sloID string, from time.Time, to time.Time) SLOClientGetParameters {
	return SLOClientGetParameters{
		sloID: sloID,
		from:  from,
		to:    to,
	}
}

// encode encodes MetricsClientQueryParameters into a URL-encoded string.
func (q *SLOClientGetParameters) encode() string {

	// TODO:  2022-01-26: Fix string composition and think about a better struct for REST parameters for all Dynatrace clients
	queryParameters := newQueryParameters()
	queryParameters.add(fromKey, common.TimestampToUnixMillisecondsString(q.from))
	queryParameters.add(toKey, common.TimestampToUnixMillisecondsString(q.to))
	queryParameters.add(timeFrameKey, "GTF")
	return q.sloID + "?" + queryParameters.encode()
}

type SLOResult struct {
	Name                string  `json:"name"`
	EvaluatedPercentage float64 `json:"evaluatedPercentage"`
	Error               string  `json:"error"`
	Target              float64 `json:"target"`
	Warning             float64 `json:"warning"`
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
func (c *SLOClient) Get(parameters SLOClientGetParameters) (*SLOResult, error) {
	body, err := c.client.Get(SLOPath + "/" + parameters.encode())
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

package dynatrace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const SLOPath = "/api/v2/slo"

// SLORequiredDelay is delay required between the end of a timeframe and an SLO API request using it.
const SLORequiredDelay = 2 * time.Minute

// SLOMaximumWait is maximum acceptable wait time between the end of a timeframe and an SLO API request using it.
const SLOMaximumWait = 4 * time.Minute

const (
	timeFrameKey = "timeFrame"
)

// SLOClientGetRequest encapsulates the request for the SLOClient's Get method.
type SLOClientGetRequest struct {
	sloID     string
	timeframe common.Timeframe
}

// NewSLOClientGetRequest creates new SLOClientGetRequest.
func NewSLOClientGetRequest(sloID string, timeframe common.Timeframe) SLOClientGetRequest {
	return SLOClientGetRequest{
		sloID:     sloID,
		timeframe: timeframe,
	}
}

// RequestString encodes SLOClientGetRequest into a request string.
func (q *SLOClientGetRequest) RequestString() string {
	queryParameters := newQueryParameters()
	queryParameters.add(fromKey, common.TimestampToUnixMillisecondsString(q.timeframe.Start()))
	queryParameters.add(toKey, common.TimestampToUnixMillisecondsString(q.timeframe.End()))
	queryParameters.add(timeFrameKey, "GTF")

	return SLOPath + "/" + url.PathEscape(q.sloID) + "?" + queryParameters.encode()
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

// Get calls Dynatrace API to retrieve the values of the Dynatrace SLO for that timeframe.
// It returns a SLOResult object on success, an error otherwise.
func (c *SLOClient) Get(ctx context.Context, request SLOClientGetRequest) (*SLOResult, error) {
	err := NewTimeframeDelay(request.timeframe, SLORequiredDelay, SLOMaximumWait).Wait(ctx)
	if err != nil {
		return nil, err
	}

	body, err := c.client.Get(ctx, request.RequestString())
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
		return nil, fmt.Errorf("Dynatrace API returned an error: %s", result.Error)
	}

	return &result, nil
}

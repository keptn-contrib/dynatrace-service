package dynatrace

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
)

const USQLPath = "/api/v1/userSessionQueryLanguage/table"

// USQLRequiredDelay is delay required between the end of a timeframe and an USQL API request using it.
const USQLRequiredDelay = 6 * time.Minute

// USQLMaximumWait is maximum acceptable wait time between the end of a timeframe and an USQL API request using it.
const USQLMaximumWait = 8 * time.Minute

const (
	queryKey             = "query"
	explainKey           = "explain"
	addDeepLinkFieldsKey = "addDeepLinkFields"
	startTimestampKey    = "startTimestamp"
	endTimestampKey      = "endTimestamp"
)

// USQLClientQueryRequest encapsulates the request for the USQLClient's GetByQuery method.
type USQLClientQueryRequest struct {
	query     usql.Query
	timeframe common.Timeframe
}

// NewUSQLClientQueryRequest creates new USQLClientQueryRequest.
func NewUSQLClientQueryRequest(query usql.Query, timeframe common.Timeframe) USQLClientQueryRequest {
	return USQLClientQueryRequest{
		query:     query,
		timeframe: timeframe,
	}
}

// RequestString encodes USQLClientQueryRequest into a request string.
func (q *USQLClientQueryRequest) RequestString() string {
	queryParameters := newQueryParameters()
	queryParameters.add(queryKey, q.query.GetQuery())
	queryParameters.add(explainKey, "false")
	queryParameters.add(addDeepLinkFieldsKey, "false")
	queryParameters.add(startTimestampKey, common.TimestampToUnixMillisecondsString(q.timeframe.Start()))
	queryParameters.add(endTimestampKey, common.TimestampToUnixMillisecondsString(q.timeframe.End()))

	return USQLPath + "?" + queryParameters.encode()
}

// DTUSQLResult struct
type DTUSQLResult struct {
	ExtrapolationLevel int             `json:"extrapolationLevel"`
	ColumnNames        []string        `json:"columnNames"`
	Values             [][]interface{} `json:"values"`
}

type USQLClient struct {
	client ClientInterface
}

func NewUSQLClient(client ClientInterface) *USQLClient {
	return &USQLClient{
		client: client,
	}
}

// GetByQuery executes the passed USQL API query, validates that the call returns data and returns the data set.
func (uc *USQLClient) GetByQuery(ctx context.Context, request USQLClientQueryRequest) (*DTUSQLResult, error) {
	err := NewTimeframeDelay(request.timeframe, USQLRequiredDelay, USQLMaximumWait).Wait(ctx)
	if err != nil {
		return nil, err
	}

	body, err := uc.client.Get(ctx, request.RequestString())
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DTUSQLResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

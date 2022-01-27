package dynatrace

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
)

const USQLPath = "/api/v1/userSessionQueryLanguage/table"

const (
	queryKey             = "query"
	explainKey           = "explain"
	addDeepLinkFieldsKey = "addDeepLinkFields"
	startTimestampKey    = "startTimestamp"
	endTimestampKey      = "endTimestamp"
)

// USQLClientQueryParameters encapsulates the query parameters for the USQLClient's GetByQuery method.
type USQLClientQueryParameters struct {
	query          usql.Query
	startTimestamp time.Time
	endTimestamp   time.Time
}

// NewUSQLClientQueryParameters creates new USQLClientQueryParameters.
func NewUSQLClientQueryParameters(query usql.Query, startTimestamp time.Time, endTimestamp time.Time) USQLClientQueryParameters {
	return USQLClientQueryParameters{
		query:          query,
		startTimestamp: startTimestamp,
		endTimestamp:   endTimestamp,
	}
}

// encode encodes USQLClientQueryParameters into a URL-encoded string.
func (q *USQLClientQueryParameters) encode() string {
	queryParameters := newQueryParameters()
	queryParameters.add(queryKey, q.query.GetQuery())
	queryParameters.add(explainKey, "false")
	queryParameters.add(addDeepLinkFieldsKey, "false")
	queryParameters.add(startTimestampKey, common.TimestampToString(q.startTimestamp))
	queryParameters.add(endTimestampKey, common.TimestampToString(q.endTimestamp))
	return queryParameters.encode()
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

// GetByQuery executes the passed USQL API query, validates that the call returns data and returns the data set
func (uc *USQLClient) GetByQuery(parameters USQLClientQueryParameters) (*DTUSQLResult, error) {
	body, err := uc.client.Get(USQLPath + "?" + parameters.encode())
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DTUSQLResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// if no data comes back
	if len(result.Values) == 0 {
		// there are no data points - try again?
		return nil, errors.New("Dynatrace USQL API returned zero data points")
	}

	return &result, nil
}

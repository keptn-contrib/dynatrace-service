package dynatrace

import (
	"encoding/json"
	"errors"
)

const usqlPath = "/api/v1/userSessionQueryLanguage/table"

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
func (uc *USQLClient) GetByQuery(usql string) (*DTUSQLResult, error) {
	body, err := uc.client.Get(usqlPath + "?" + usql)
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
		return nil, errors.New("dynatrace USQL Query didnt return any DataPoints")
	}

	return &result, nil
}

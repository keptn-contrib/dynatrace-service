package dynatrace

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
)

const usqlPath = "/api/v1/userSessionQueryLanguage/table"

// DTUSQLResult struct
type DTUSQLResult struct {
	ExtrapolationLevel int             `json:"extrapolationLevel"`
	ColumnNames        []string        `json:"columnNames"`
	Values             [][]interface{} `json:"values"`
}

type USQLClient struct {
	client *Client
}

func NewUSQLClient(client *Client) *USQLClient {
	return &USQLClient{
		client: client,
	}
}

// GetByQuery executes the passed USQL API query, validates that the call returns data and returns the data set
func (uc *USQLClient) GetByQuery(usql string) (*DTUSQLResult, error) {
	path := usqlPath + "?" + usql
	log.WithField("query", uc.client.DynatraceCreds.Tenant+path).Debug("Final USQL Query")

	body, err := uc.client.Get(path)
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

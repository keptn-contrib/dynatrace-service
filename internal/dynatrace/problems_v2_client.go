package dynatrace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const problemsV2Path = "/api/v2/problems"

// ProblemQueryResult result of query to /api/v2/problems
// Here only totalCount is considered as that is the only field that is used
type problemQueryResult struct {
	TotalCount int `json:"totalCount"`
}

// Problem problem details returned by /api/v2/problems/{PROBLEM-ID}
// Here only status is considered as that is the only field that is used
type problem struct {
	Status string `json:"status"`
}

// ProblemsV2Client is a client for interacting with the Dynatrace problems endpoints
type ProblemsV2Client struct {
	client ClientInterface
}

// NewProblemsV2Client creates a new ProblemsV2Client
func NewProblemsV2Client(client ClientInterface) *ProblemsV2Client {
	return &ProblemsV2Client{
		client: client,
	}
}

// GetTotalCountByQuery calls the Dynatrace V2 API to retrieve the total count of problems for a given query and timeframe
func (pc *ProblemsV2Client) GetTotalCountByQuery(problemQuery string, startUnix time.Time, endUnix time.Time) (int, error) {
	body, err := pc.client.Get(
		fmt.Sprintf("%s?from=%s&to=%s&%s",
			problemsV2Path,
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix),
			problemQuery))
	if err != nil {
		return 0, err
	}

	var result problemQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, err
	}

	return result.TotalCount, nil
}

// GetStatusById calls the Dynatrace API to retrieve the status of a given problemID
func (pc *ProblemsV2Client) GetStatusById(problemID string) (string, error) {
	body, err := pc.client.Get(problemsV2Path + "/" + problemID)
	if err != nil {
		return "", err
	}

	// parse response json
	var result problem
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Status, nil
}

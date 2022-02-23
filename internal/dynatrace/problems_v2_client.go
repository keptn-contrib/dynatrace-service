package dynatrace

import (
	"encoding/json"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
)

// ProblemStatusOpen is the status of an open problem
const ProblemStatusOpen = "OPEN"

// ProblemsV2Path is the base endpoint for Problems API v2
const ProblemsV2Path = "/api/v2/problems"

const (
	problemSelectorKey = "problemSelector"
)

// ProblemsV2ClientQueryParameters encapsulates the query parameters for the ProblemsV2Client's GetTotalCountByQuery method.
type ProblemsV2ClientQueryParameters struct {
	query problems.Query
	from  time.Time
	to    time.Time
}

// NewProblemsV2ClientQueryParameters creates new ProblemsV2ClientQueryParameters.
func NewProblemsV2ClientQueryParameters(query problems.Query, from time.Time, to time.Time) ProblemsV2ClientQueryParameters {
	return ProblemsV2ClientQueryParameters{
		query: query,
		from:  from,
		to:    to,
	}
}

// encode encodes ProblemsV2ClientQueryParameters into a URL-encoded string.
func (q *ProblemsV2ClientQueryParameters) encode() string {
	queryParameters := newQueryParameters()
	if q.query.GetProblemSelector() != "" {
		queryParameters.add(problemSelectorKey, q.query.GetProblemSelector())
	}
	if q.query.GetEntitySelector() != "" {
		queryParameters.add(entitySelectorKey, q.query.GetEntitySelector())
	}

	queryParameters.add(fromKey, common.TimestampToUnixMillisecondsString(q.from))
	queryParameters.add(toKey, common.TimestampToUnixMillisecondsString(q.to))
	return queryParameters.encode()
}

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
func (pc *ProblemsV2Client) GetTotalCountByQuery(parameters ProblemsV2ClientQueryParameters) (int, error) {
	body, err := pc.client.Get(ProblemsV2Path + "?" + parameters.encode())
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

// GetStatusByID calls the Dynatrace API to retrieve the status of a given problemID
func (pc *ProblemsV2Client) GetStatusByID(problemID string) (string, error) {
	body, err := pc.client.Get(ProblemsV2Path + "/" + problemID)
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

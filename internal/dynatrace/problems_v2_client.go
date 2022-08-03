package dynatrace

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
)

// ProblemStatusOpen is the status of an open problem
const ProblemStatusOpen = "OPEN"

// ProblemsV2Path is the base endpoint for Problems API v2
const ProblemsV2Path = "/api/v2/problems"

// ProblemsV2RequiredDelay is delay required between the end of a timeframe and an PV2 API request using it.
const ProblemsV2RequiredDelay = 2 * time.Minute

// ProblemsV2MaximumWait is maximum acceptable wait time between the end of a timeframe and an PV2 API request using it.
const ProblemsV2MaximumWait = 4 * time.Minute

const (
	problemSelectorKey = "problemSelector"
)

// ProblemsV2ClientQueryRequest encapsulates the request for the ProblemsV2Client's GetTotalCountByQuery method.
type ProblemsV2ClientQueryRequest struct {
	query     problems.Query
	timeframe common.Timeframe
}

// NewProblemsV2ClientQueryRequest creates new ProblemsV2ClientQueryRequest.
func NewProblemsV2ClientQueryRequest(query problems.Query, timeframe common.Timeframe) ProblemsV2ClientQueryRequest {
	return ProblemsV2ClientQueryRequest{
		query:     query,
		timeframe: timeframe,
	}
}

// RequestString encodes ProblemsV2ClientQueryRequest into a request string.
func (q *ProblemsV2ClientQueryRequest) RequestString() string {
	queryParameters := newQueryParameters()
	if q.query.GetProblemSelector() != "" {
		queryParameters.add(problemSelectorKey, q.query.GetProblemSelector())
	}
	if q.query.GetEntitySelector() != "" {
		queryParameters.add(entitySelectorKey, q.query.GetEntitySelector())
	}

	queryParameters.add(fromKey, common.TimestampToUnixMillisecondsString(q.timeframe.Start()))
	queryParameters.add(toKey, common.TimestampToUnixMillisecondsString(q.timeframe.End()))

	return ProblemsV2Path + "?" + queryParameters.encode()
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

// GetTotalCountByQuery calls the Dynatrace V2 API to retrieve the total count of problems for a given query and timeframe.
func (pc *ProblemsV2Client) GetTotalCountByQuery(ctx context.Context, request ProblemsV2ClientQueryRequest) (int, error) {
	err := NewTimeframeDelay(request.timeframe, ProblemsV2RequiredDelay, ProblemsV2MaximumWait).Wait(ctx)
	if err != nil {
		return 0, err
	}

	body, err := pc.client.Get(ctx, request.RequestString())
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

// GetStatusByID calls the Dynatrace API to retrieve the status of a given problemID.
func (pc *ProblemsV2Client) GetStatusByID(ctx context.Context, problemID string) (string, error) {
	body, err := pc.client.Get(ctx, ProblemsV2Path+"/"+problemID)
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

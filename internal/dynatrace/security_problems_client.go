package dynatrace

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/secpv2"
)

// SecurityProblemsPath is the base endpoint for Security Problems API v2
const SecurityProblemsPath = "/api/v2/securityProblems"

// SecurityProblemsV2RequiredDelay is delay required between the end of a timeframe and an SECPV2 API request using it.
const SecurityProblemsV2RequiredDelay = 2 * time.Minute

// SecurityProblemsV2MaximumWait is maximum acceptable wait time between the end of a timeframe and an SECPV2 API request using it.
const SecurityProblemsV2MaximumWait = 4 * time.Minute

const (
	securityProblemSelectorKey = "securityProblemSelector"
)

// SecurityProblemsV2ClientQueryRequest encapsulates the request for the SecurityProblemsClient's GetTotalCountByQuery method.
type SecurityProblemsV2ClientQueryRequest struct {
	query     secpv2.Query
	timeframe common.Timeframe
}

// NewSecurityProblemsClientQueryRequest creates new SecurityProblemsV2ClientQueryRequest.
func NewSecurityProblemsClientQueryRequest(query secpv2.Query, timeframe common.Timeframe) SecurityProblemsV2ClientQueryRequest {
	return SecurityProblemsV2ClientQueryRequest{
		query:     query,
		timeframe: timeframe,
	}
}

// RequestString encodes SecurityProblemsV2ClientQueryRequest into a request string.
func (q *SecurityProblemsV2ClientQueryRequest) RequestString() string {
	queryParameters := newQueryParameters()
	if q.query.GetSecurityProblemSelector() != "" {
		queryParameters.add(securityProblemSelectorKey, q.query.GetSecurityProblemSelector())
	}
	queryParameters.add(fromKey, common.TimestampToUnixMillisecondsString(q.timeframe.Start()))
	queryParameters.add(toKey, common.TimestampToUnixMillisecondsString(q.timeframe.End()))

	return SecurityProblemsPath + "?" + queryParameters.encode()
}

type securityProblemQueryResult struct {
	TotalCount int `json:"totalCount"`
}

// SecurityProblemsClient is a client for interacting with the Dynatrace security problems endpoints
type SecurityProblemsClient struct {
	client ClientInterface
}

// NewSecurityProblemsClient creates a new SecurityProblemsClient
func NewSecurityProblemsClient(client ClientInterface) *SecurityProblemsClient {
	return &SecurityProblemsClient{
		client: client,
	}
}

// GetTotalCountByQuery calls the Dynatrace API to retrieve the total count of security problems for the given query and timeframe.
func (sc *SecurityProblemsClient) GetTotalCountByQuery(ctx context.Context, request SecurityProblemsV2ClientQueryRequest) (int, error) {
	err := NewTimeframeDelay(request.timeframe, SecurityProblemsV2RequiredDelay, SecurityProblemsV2MaximumWait).Wait(ctx)
	if err != nil {
		return 0, err
	}

	body, err := sc.client.Get(ctx, request.RequestString())
	if err != nil {
		return 0, err
	}

	var result securityProblemQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, err
	}

	return result.TotalCount, nil
}

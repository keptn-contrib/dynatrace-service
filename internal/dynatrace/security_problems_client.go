package dynatrace

import (
	"encoding/json"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/secpv2"
)

// SecurityProblemsPath is the base endpoint for Security Problems API v2
const SecurityProblemsPath = "/api/v2/securityProblems"

const (
	securityProblemSelectorKey = "securityProblemSelector"
)

// SecurityProblemsV2ClientQueryParameters encapsulates the query parameters for the SecurityProblemsClient's GetTotalCountByQuery method.
type SecurityProblemsV2ClientQueryParameters struct {
	query secpv2.Query
	from  time.Time
	to    time.Time
}

// NewSecurityProblemsV2ClientQueryParameters creates new SecurityProblemsV2ClientQueryParameters.
func NewSecurityProblemsV2ClientQueryParameters(query secpv2.Query, from time.Time, to time.Time) SecurityProblemsV2ClientQueryParameters {
	return SecurityProblemsV2ClientQueryParameters{
		query: query,
		from:  from,
		to:    to,
	}
}

// encode encodes SecurityProblemsV2ClientQueryParameters into a URL-encoded string.
func (q *SecurityProblemsV2ClientQueryParameters) encode() string {
	queryParameters := newQueryParameters()
	if q.query.GetSecurityProblemSelector() != "" {
		queryParameters.add(securityProblemSelectorKey, q.query.GetSecurityProblemSelector())
	}

	queryParameters.add(fromKey, common.TimestampToUnixMillisecondsString(q.from))
	queryParameters.add(toKey, common.TimestampToUnixMillisecondsString(q.to))
	return queryParameters.encode()
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

// GetTotalCountByQuery calls the Dynatrace API to retrieve the total count of security problems for the given query and timeframe
func (sc *SecurityProblemsClient) GetTotalCountByQuery(parameters SecurityProblemsV2ClientQueryParameters) (int, error) {
	body, err := sc.client.Get(SecurityProblemsPath + "?" + parameters.encode())
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

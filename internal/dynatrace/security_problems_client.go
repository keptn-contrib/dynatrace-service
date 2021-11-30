package dynatrace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const securityProblemsPath = "/api/v2/securityProblems"

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
func (sc *SecurityProblemsClient) GetTotalCountByQuery(securityProblemQuery string, startUnix time.Time, endUnix time.Time) (int, error) {
	body, err := sc.client.Get(
		fmt.Sprintf("%s?from=%s&to=%s&%s",
			securityProblemsPath,
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix),
			securityProblemQuery))
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

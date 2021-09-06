package dynatrace

import (
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"time"
)

const securityProblemsPath = "/api/v2/securityProblems"

// SecurityProblemQueryResult Result of/api/v2/securityProblems
type SecurityProblemQueryResult struct {
	TotalCount       int               `json:"totalCount"`
	PageSize         int               `json:"pageSize"`
	NextPageKey      string            `json:"nextPageKey"`
	SecurityProblems []SecurityProblem `json:"securityProblems"`
}

// SecurityProblem Problem Detail returned by /api/v2/securityProblems
type SecurityProblem struct {
	SecurityProblemID    string `json:"securityProblemId"`
	DisplayID            int    `json:"displayId"`
	State                string `json:"state"`
	VulnerabilityID      string `json:"vulnerabilityId"`
	VulnerabilityType    string `json:"vulnerabilityType"`
	FirstSeenTimestamp   int    `json:"firstSeenTimestamp"`
	LastUpdatedTimestamp int    `json:"lastUpdatedTimestamp"`
	RiskAssessment       struct {
		RiskCategory string `json:"riskCategory"`
		RiskScore    struct {
			Value int `json:"value"`
		} `json:"riskScore"`
		Exposed                bool `json:"exposed"`
		SensitiveDataAffected  bool `json:"sensitiveDataAffected"`
		PublicExploitAvailable bool `json:"publicExploitAvailable"`
	} `json:"riskAssessment"`
	ManagementZones      []string `json:"managementZones"`
	VulnerableComponents []struct {
		ID                          string   `json:"id"`
		DisplayName                 string   `json:"displayName"`
		FileName                    string   `json:"fileName"`
		NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
		VulnerableProcesses         []string `json:"vulnerableProcesses"`
	} `json:"vulnerableComponents"`
	VulnerableEntities  []string `json:"vulnerableEntities"`
	ExposedEntities     []string `json:"exposedEntities"`
	SensitiveDataAssets []string `json:"sensitiveDataAssets"`
	AffectedEntities    struct {
		Applications []struct {
			ID                          string   `json:"id"`
			NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
			VulnerableProcesses         []string `json:"vulnerableProcesses"`
		} `json:"applications"`
		Services []struct {
			ID                          string   `json:"id"`
			NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
			VulnerableProcesses         []string `json:"vulnerableProcesses"`
		} `json:"services"`
		Hosts []struct {
			ID                          string   `json:"id"`
			NumberOfVulnerableProcesses int      `json:"numberOfVulnerableProcesses"`
			VulnerableProcesses         []string `json:"vulnerableProcesses"`
		} `json:"hosts"`
		Databases []string `json:"databases"`
	} `json:"affectedEntities"`
}

// SecurityProblemsClient is a client for interacting with the Dynatrace security problems endpoints
type SecurityProblemsClient struct {
	client *Client
}

// NewSecurityProblemsClient creates a new SecurityProblemsClient
func NewSecurityProblemsClient(client *Client) *SecurityProblemsClient {
	return &SecurityProblemsClient{
		client: client,
	}
}

// GetByQuery Calls the Dynatrace API to retrieve the list of security problems for that timeframe.
// It returns a SecurityProblemQueryResult object on success, an error otherwise.
func (sc *SecurityProblemsClient) GetByQuery(problemQuery string, startUnix time.Time, endUnix time.Time) (*SecurityProblemQueryResult, error) {
	body, err := sc.client.Get(
		fmt.Sprintf("%s?from=%s&to=%s&%s",
			securityProblemsPath,
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix),
			problemQuery))
	if err != nil {
		return nil, err
	}

	var result SecurityProblemQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

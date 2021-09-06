package dynatrace

import (
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"time"
)

const problemsV2Path = "/api/v2/problems"

// ProblemQueryResult Result of /api/v2/problems
type ProblemQueryResult struct {
	TotalCount int       `json:"totalCount"`
	PageSize   int       `json:"pageSize"`
	Problems   []Problem `json:"problems"`
}

// Problem problem details returned by /api/v2/problems
type Problem struct {
	ProblemID        string `json:"problemId"`
	DisplayID        string `json:"displayId"`
	Title            string `json:"title"`
	ImpactLevel      string `json:"impactLevel"`
	SeverityLevel    string `json:"severityLevel"`
	Status           string `json:"status"`
	AffectedEntities []struct {
		EntityID struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"entityId"`
		Name string `json:"name"`
	} `json:"affectedEntities"`
	ImpactedEntities []struct {
		EntityID struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"entityId"`
		Name string `json:"name"`
	} `json:"impactedEntities"`
	RootCauseEntity struct {
		EntityID struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"entityId"`
		Name string `json:"name"`
	} `json:"rootCauseEntity"`
	ManagementZones []interface{} `json:"managementZones"`
	EntityTags      []struct {
		Context              string `json:"context"`
		Key                  string `json:"key"`
		Value                string `json:"value"`
		StringRepresentation string `json:"stringRepresentation"`
	} `json:"entityTags"`
	ProblemFilters []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"problemFilters"`
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

// ProblemsV2Client is a client for interacting with the Dynatrace problems endpoints
type ProblemsV2Client struct {
	client *Client
}

// NewProblemsV2Client creates a new ProblemsV2Client
func NewProblemsV2Client(client *Client) *ProblemsV2Client {
	return &ProblemsV2Client{
		client: client,
	}
}

// GetByQuery Calls the Dynatrace V2 API to retrieve the the list of problems for that timeframe
// It returns a ProblemQueryResult object on success, an error otherwise
func (pc *ProblemsV2Client) GetByQuery(problemQuery string, startUnix time.Time, endUnix time.Time) (*ProblemQueryResult, error) {
	body, err := pc.client.Get(
		fmt.Sprintf("%s?from=%s&to=%s&%s",
			problemsV2Path,
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix),
			problemQuery))
	if err != nil {
		return nil, err
	}

	var result ProblemQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetById Calls the Dynatrace API to retrieve Problem Details for a given problemID
// It returns a Problem object on success, an error otherwise
func (pc *ProblemsV2Client) GetById(problemID string) (*Problem, error) {
	body, err := pc.client.Get(problemsV2Path + "/" + problemID)
	if err != nil {
		return nil, err
	}

	// parse response json
	var result Problem
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

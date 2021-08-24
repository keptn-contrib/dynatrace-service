package dynatrace

import (
	"encoding/json"
)

const problemDetailsPath = "/api/v1/problem/details"

// ProblemsClient is a client for interacting with the Dynatrace problems endpoints
type ProblemsClient struct {
	client *DynatraceHelper
}

// NewProblemsClient creates a new ProblemsClient
func NewProblemsClient(client *DynatraceHelper) *ProblemsClient {
	return &ProblemsClient{
		client: client,
	}
}

// AddProblemComment sends a comment on a DT problem
func (pc *ProblemsClient) AddProblemComment(problemID string, comment string) (string, error) {
	payload := map[string]string{"comment": comment, "user": "keptn", "context": "keptn-remediation"}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return pc.client.Post(problemDetailsPath+"/"+problemID+"/comments", []byte(jsonPayload))
}

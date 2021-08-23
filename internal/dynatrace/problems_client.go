package dynatrace

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
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

// SendProblemComment sends a comment on a DT problem
func (pc *ProblemsClient) SendProblemComment(problemID string, comment string) error {
	payload := map[string]string{"comment": comment, "user": "keptn", "context": "keptn-remediation"}
	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	log.WithField("jsonPayload", jsonPayload).Info("Sending problem event")

	resp, err := pc.client.Post(problemDetailsPath+"/"+problemID+"/comments", []byte(jsonPayload))

	log.WithField("response", resp).Info("Received response from Dynatrace API")
	if err != nil {
		return err
	}
	return nil
}

package dynatrace

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

const problemDetailsPath = "/api/v1/problem/details"

// ProblemsClient is a client for interacting with the Dynatrace problems endpoints
type ProblemsClient struct {
	client ClientInterface
}

// NewProblemsClient creates a new ProblemsClient
func NewProblemsClient(client ClientInterface) *ProblemsClient {
	return &ProblemsClient{
		client: client,
	}
}

// addProblemComment sends a comment on a DT problem.
func (pc *ProblemsClient) addProblemComment(ctx context.Context, problemID string, comment string) ([]byte, error) {
	payload := map[string]string{"comment": comment, "user": "keptn", "context": "keptn-remediation"}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return pc.client.Post(ctx, problemDetailsPath+"/"+problemID+"/comments", jsonPayload)
}

// AddProblemComment sends a comment on a DT problem and logs errors if necessary.
func (pc *ProblemsClient) AddProblemComment(ctx context.Context, pid string, comment string) {
	log.WithField("comment", comment).Info("Adding problem comment")
	response, err := pc.addProblemComment(ctx, pid, comment)
	if err != nil {
		log.WithError(err).Error("Error adding problem comment")
		return
	}

	log.WithField("response", string(response)).Info("Received problem comment response")
}

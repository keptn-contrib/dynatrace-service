package lib

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// SendProblemComment sends a commont on a DT problem
func (dt *DynatraceHelper) SendProblemComment(problemID string, comment string) error {
	dtCommentPayload := map[string]string{"comment": comment, "user": "keptn", "context": "keptn-remediation"}
	jsonPayload, err := json.Marshal(dtCommentPayload)

	if err != nil {
		return err
	}

	log.WithField("jsonPayload", jsonPayload).Info("Sending problem event")

	resp, err := dt.sendDynatraceAPIRequest("/api/v1/problem/details/"+problemID+"/comments", "POST", jsonPayload)

	log.WithField("response", resp).Info("Received response from Dynatrace API")
	if err != nil {
		return err
	}
	return nil
}

package lib

import "encoding/json"

// SendProblemComment sends a commont on a DT problem
func (dt *DynatraceHelper) SendProblemComment(problemID string, comment string) error {
	dtCommentPayload := map[string]string{"comment": comment, "user": "keptn", "context": "keptn-remediation"}
	jsonPayload, err := json.Marshal(dtCommentPayload)

	if err != nil {
		return err
	}

	dt.Logger.Info("Sending problem event: " + string(jsonPayload))

	resp, err := dt.sendDynatraceAPIRequest("/api/v1/problem/details/"+problemID+"/comments", "POST", jsonPayload)

	dt.Logger.Info("Received response from Dynatrace API: " + resp)
	if err != nil {
		return err
	}
	return nil
}

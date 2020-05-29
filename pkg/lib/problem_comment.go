package lib

import "encoding/json"

func (dt *DynatraceHelper) SendProblemComment(problemID string, comment string, dynatraceSecretName string) error {
	dtCommentPayload := map[string]string{"comment": comment, "user": "keptn", "context": "keptn-remediation"}
	jsonPayload, _ := json.Marshal(dtCommentPayload)

	dt.Logger.Info("Sending problem event: " + string(jsonPayload))

	resp, err := dt.sendDynatraceAPIRequest(dynatraceSecretName, "/api/v1/problem/details/"+problemID+"/comments", "POST", string(jsonPayload))

	dt.Logger.Info("Received response from Dynatrace API: " + resp)
	if err != nil {
		return err
	}
	return nil
}

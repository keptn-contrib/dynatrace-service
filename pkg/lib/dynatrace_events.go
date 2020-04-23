package lib

import (
	"encoding/json"
)

// Sends an event to the Dynatrace events API
func (dt *DynatraceHelper) SendEvent(dtEvent interface{}, dtCreds string) {
	dt.Logger.Info("Sending event to Dynatrace API")

	jsonString, err := json.Marshal(dtEvent)

	if err != nil {
		dt.Logger.Error("Error while generating Dynatrace API Request payload.")
		return
	}

	body, err := dt.sendDynatraceAPIRequest(dtCreds, "/api/v1/events", "POST", string(jsonString))

	if err != nil {
		dt.Logger.Error("Failed sending Dynatrace API request: " + err.Error())
		dt.Logger.Error("Response Body:" + body)
	}

	dt.Logger.Debug("Dynatrace API has accepted the event")
	dt.Logger.Debug("Response Body:" + body)
}

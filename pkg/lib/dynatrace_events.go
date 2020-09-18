package lib

import (
	"encoding/json"
	"fmt"
)

// Sends an event to the Dynatrace events API
func (dt *DynatraceHelper) SendEvent(dtEvent interface{}) {
	dt.Logger.Info("Sending event to Dynatrace API")

	jsonString, err := json.Marshal(dtEvent)

	if err != nil {
		dt.Logger.Error("Error while generating Dynatrace API Request payload.")
		return
	}

	body, err := dt.sendDynatraceAPIRequest("/api/v1/events", "POST", jsonString)
	if err != nil {
		dt.Logger.Error(fmt.Sprintf("failed sending Dynatrace API request: %v", err))
	} else {
		dt.Logger.Debug(fmt.Sprintf("Dynatrace API has accepted the event. Response: %s", body))
	}

}

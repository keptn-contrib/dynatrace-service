package dynatrace

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// SendEvent sends an event to the Dynatrace events API
func (dt *DynatraceHelper) SendEvent(dtEvent interface{}) {
	log.Info("Sending event to Dynatrace API")

	jsonString, err := json.Marshal(dtEvent)

	if err != nil {
		log.WithError(err).Error("Error while generating Dynatrace API Request payload.")
		return
	}

	body, err := dt.SendDynatraceAPIRequest("/api/v1/events", "POST", jsonString)
	if err != nil {
		log.WithError(err).Error("Failed sending Dynatrace API request")
	} else {
		log.WithField("body", body).Debug("Dynatrace API has accepted the event")
	}

}

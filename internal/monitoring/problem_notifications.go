package monitoring

import (
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"

	log "github.com/sirupsen/logrus"
)

type ProblemNotificationCreation struct {
	client *lib.DynatraceHelper
}

func NewProblemNotificationCreation(client *lib.DynatraceHelper) *ProblemNotificationCreation {
	return &ProblemNotificationCreation{
		client: client,
	}
}

// Create sets up/updates the DT problem notification and returns it
func (pn *ProblemNotificationCreation) Create() lib.ConfigResult {
	if !lib.IsProblemNotificationsGenerationEnabled() {
		return lib.ConfigResult{}
	}

	log.Info("Setting up problem notifications in Dynatrace Tenant")

	alertingProfileId, err := pn.setupAlertingProfile()
	if err != nil {
		log.WithError(err).Error("Failed to set up problem notification")
		return lib.ConfigResult{
			Success: false,
			Message: "failed to set up problem notification: " + err.Error(),
		}
	}

	response, err := pn.client.SendDynatraceAPIRequest("/api/config/v1/notifications", "GET", nil)
	existingNotifications := lib.DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), &existingNotifications)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal notifications")
	}

	for _, notification := range existingNotifications.Values {
		if notification.Name == "Keptn Problem Notification" {
			_, err = pn.client.SendDynatraceAPIRequest("/api/config/v1/notifications/"+notification.ID, "DELETE", nil)
			if err != nil {
				// Error occurred but continue
				log.WithError(err).WithField("notificationId", notification.ID).Error("Failed to delete notification")
			}
		}
	}

	keptnCredentials, err := credentials.GetKeptnCredentials()
	if err != nil {
		log.WithError(err).Error("Failed to retrieve Keptn API credentials")
		return lib.ConfigResult{
			Success: false,
			Message: "failed to retrieve Keptn API credentials: " + err.Error(),
		}
	}

	problemNotification := lib.PROBLEM_NOTIFICATION_PAYLOAD
	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_DNS", keptnCredentials.APIURL)
	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_TOKEN", keptnCredentials.APIToken)
	problemNotification = strings.ReplaceAll(problemNotification, "$ALERTING_PROFILE_ID", alertingProfileId)

	_, err = pn.client.SendDynatraceAPIRequest("/api/config/v1/notifications", "POST", []byte(problemNotification))
	if err != nil {
		log.WithError(err).Error("Failed to set up problem notification")
		return lib.ConfigResult{
			Success: false,
			Message: "failed to set up problem notification: " + err.Error(),
		}
	}

	return lib.ConfigResult{
		Success: true,
		Message: "Successfully set up Keptn Alerting Profile and Problem Notifications",
	}
}

func (pn *ProblemNotificationCreation) setupAlertingProfile() (string, error) {
	log.Info("Checking Keptn alerting profile availability")
	response, err := pn.client.SendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "GET", nil)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).Debug("Could not get alerting profiles")
	} else {
		existingAlertingProfiles := lib.DTAPIListResponse{}

		err = json.Unmarshal([]byte(response), &existingAlertingProfiles)
		if err != nil {
			// Error occurred but continue
			log.WithError(err).Error("Failed to unmarshal alerting profiles")
		}
		for _, ap := range existingAlertingProfiles.Values {
			if ap.Name == "Keptn" {
				log.Info("Keptn alerting profile available")
				return ap.ID, nil
			}
		}
	}

	log.Info("Creating Keptn alerting profile.")
	alertingProfile := lib.CreateKeptnAlertingProfile()
	alertingProfilePayload, err := json.Marshal(alertingProfile)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alerting profile: %v", err)
	}

	response, err = pn.client.SendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "POST", alertingProfilePayload)
	if err != nil {
		return "", fmt.Errorf("failed to setup alerting profile: %v", err)
	}

	createdItem := &lib.Values{}

	err = json.Unmarshal([]byte(response), createdItem)
	if err != nil {
		err = checkForUnexpectedHTMLResponseError(err)
		return "", fmt.Errorf("failed to unmarshal alerting profile: %v", err)
	}
	log.Info("Alerting profile created successfully.")
	return createdItem.ID, nil
}

package lib

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"

	log "github.com/sirupsen/logrus"
)

// EnsureProblemNotificationsAreSetUp sets up/updates the DT problem notification
func (dt *DynatraceHelper) EnsureProblemNotificationsAreSetUp() {
	if !IsProblemNotificationsGenerationEnabled() {
		return
	}

	log.Info("Setting up problem notifications in Dynatrace Tenant")

	alertingProfileId, err := dt.setupAlertingProfile()
	if err != nil {
		log.WithError(err).Error("Failed to set up problem notification")
		dt.configuredEntities.ProblemNotifications.Success = false
		dt.configuredEntities.ProblemNotifications.Message = "failed to set up problem notification: " + err.Error()
		return
	}

	response, err := dt.SendDynatraceAPIRequest("/api/config/v1/notifications", "GET", nil)
	existingNotifications := DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), &existingNotifications)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal notifications")
	}

	for _, notification := range existingNotifications.Values {
		if notification.Name == "Keptn Problem Notification" {
			_, err = dt.SendDynatraceAPIRequest("/api/config/v1/notifications/"+notification.ID, "DELETE", nil)
			if err != nil {
				// Error occurred but continue
				log.WithError(err).WithField("notificationId", notification.ID).Error("Failed to delete notification")
			}
		}
	}
	problemNotification := PROBLEM_NOTIFICATION_PAYLOAD

	keptnCredentials, err := credentials.GetKeptnCredentials()
	if err != nil {
		log.WithError(err).Error("Failed to retrieve Keptn API credentials")
		dt.configuredEntities.ProblemNotifications.Success = false
		dt.configuredEntities.ProblemNotifications.Message = "failed to retrieve Keptn API credentials: " + err.Error()
		return
	}

	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_DNS", keptnCredentials.APIURL)
	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_TOKEN", keptnCredentials.APIToken)
	problemNotification = strings.ReplaceAll(problemNotification, "$ALERTING_PROFILE_ID", alertingProfileId)

	_, err = dt.SendDynatraceAPIRequest("/api/config/v1/notifications", "POST", []byte(problemNotification))
	if err != nil {
		log.WithError(err).Error("Failed to set up problem notification")
		dt.configuredEntities.ProblemNotifications.Success = false
		dt.configuredEntities.ProblemNotifications.Message = "failed to set up problem notification: " + err.Error()
	}
	dt.configuredEntities.ProblemNotifications.Success = true
	dt.configuredEntities.ProblemNotifications.Message = "Successfully set up Keptn Alerting Profile and Problem Notifications"
}

func (dt *DynatraceHelper) setupAlertingProfile() (string, error) {
	log.Info("Checking Keptn alerting profile availability")
	response, err := dt.SendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "GET", nil)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).Debug("Could not get alerting profiles")
	} else {
		existingAlertingProfiles := DTAPIListResponse{}

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
	alertingProfile := CreateKeptnAlertingProfile()
	alertingProfilePayload, err := json.Marshal(alertingProfile)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alerting profile: %v", err)
	}

	response, err = dt.SendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "POST", alertingProfilePayload)
	if err != nil {
		return "", fmt.Errorf("failed to setup alerting profile: %v", err)
	}

	createdItem := &Values{}

	err = json.Unmarshal([]byte(response), createdItem)
	if err != nil {
		err = checkForUnexpectedHTMLResponseError(err)
		return "", fmt.Errorf("failed to unmarshal alerting profile: %v", err)
	}
	log.Info("Alerting profile created successfully.")
	return createdItem.ID, nil
}

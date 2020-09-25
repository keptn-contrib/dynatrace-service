package lib

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
)

// EnsureProblemNotificationsAreSetUp sets up/updates the DT problem notification
func (dt *DynatraceHelper) EnsureProblemNotificationsAreSetUp() {
	if !IsProblemNotificationsGenerationEnabled() {
		return
	}

	dt.Logger.Info("Setting up problem notifications in Dynatrace Tenant")

	alertingProfileId, err := dt.setupAlertingProfile()
	if err != nil {
		dt.Logger.Error("failed to set up problem notification: " + err.Error())
		return
	}

	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/notifications", "GET", nil)
	existingNotifications := DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), &existingNotifications)
	if err != nil {
		dt.Logger.Error(fmt.Sprintf("failed to unmarshal notifications: %v", err))
	}

	for _, notification := range existingNotifications.Values {
		if notification.Name == "Keptn Problem Notification" {
			_, err = dt.sendDynatraceAPIRequest("/api/config/v1/notifications/"+notification.ID, "DELETE", nil)
			if err != nil {
				// Error occurred but continue
				dt.Logger.Error(fmt.Sprintf("failed to delete notification with ID %s: %v", notification.ID, err))
			}
		}
	}
	problemNotification := PROBLEM_NOTIFICATION_PAYLOAD

	keptnCredentials, err := common.GetKeptnCredentials()
	if err != nil {
		dt.Logger.Error("failed to retrieve Keptn API credentials: " + err.Error())
		return
	}

	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_DNS", keptnCredentials.ApiURL)
	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_TOKEN", keptnCredentials.ApiToken)
	problemNotification = strings.ReplaceAll(problemNotification, "$ALERTING_PROFILE_ID", alertingProfileId)

	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/notifications", "POST", []byte(problemNotification))
	if err != nil {
		dt.Logger.Error("failed to set up problem notification: " + err.Error())
	}
}

func (dt *DynatraceHelper) setupAlertingProfile() (string, error) {
	dt.Logger.Info("Checking Keptn alerting profile availability")
	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "GET", nil)
	if err != nil {
		// Error occurred but continue
		dt.Logger.Debug("could not get alerting profiles: " + err.Error())
	} else {
		existingAlertingProfiles := DTAPIListResponse{}

		err = json.Unmarshal([]byte(response), &existingAlertingProfiles)
		if err != nil {
			// Error occurred but continue
			dt.Logger.Error(fmt.Sprintf("failed to unmarshal alerting profiles: %v", err))
		}
		for _, ap := range existingAlertingProfiles.Values {
			if ap.Name == "Keptn" {
				dt.Logger.Info("Keptn alerting profile available")
				return ap.ID, nil
			}
		}
	}

	dt.Logger.Info("Creating Keptn alerting profile.")
	alertingProfile := CreateKeptnAlertingProfile()
	alertingProfilePayload, err := json.Marshal(alertingProfile)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alerting profile: %v", err)
	}

	response, err = dt.sendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "POST", alertingProfilePayload)
	if err != nil {
		return "", fmt.Errorf("failed to setup alerting profile: %v", err)
	}

	createdItem := &Values{}

	err = json.Unmarshal([]byte(response), createdItem)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal alerting profile: %v", err)
	}
	dt.Logger.Info("Alerting profile created successfully.")
	return createdItem.ID, nil
}

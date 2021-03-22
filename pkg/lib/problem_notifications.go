package lib

import (
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	"strings"
)

// EnsureProblemNotificationsAreSetUp sets up/updates the DT problem notification
func (dt *DynatraceHelper) EnsureProblemNotificationsAreSetUp() {
	if !IsProblemNotificationsGenerationEnabled() {
		return
	}

	dt.Logger.Info("Setting up problem notifications in Dynatrace Tenant")

	alertingProfileId, err := dt.setupAlertingProfile()
	if err != nil {
		msg := "failed to set up problem notification: " + err.Error()
		dt.Logger.Error(msg)
		dt.configuredEntities.ProblemNotifications.Success = false
		dt.configuredEntities.ProblemNotifications.Message = msg
		return
	}

	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/notifications", "GET", nil)
	existingNotifications := DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), &existingNotifications)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal notifications: %v", err)
		dt.Logger.Error(msg)
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

	keptnCredentials, err := credentials.GetKeptnCredentials()
	if err != nil {
		msg := "failed to retrieve Keptn API credentials: " + err.Error()
		dt.Logger.Error(msg)
		dt.configuredEntities.ProblemNotifications.Success = false
		dt.configuredEntities.ProblemNotifications.Message = msg
		return
	}

	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_DNS", keptnCredentials.APIURL)
	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_TOKEN", keptnCredentials.APIToken)
	problemNotification = strings.ReplaceAll(problemNotification, "$ALERTING_PROFILE_ID", alertingProfileId)

	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/notifications", "POST", []byte(problemNotification))
	if err != nil {
		msg := "failed to set up problem notification: " + err.Error()
		dt.Logger.Error(msg)
		dt.configuredEntities.ProblemNotifications.Success = false
		dt.configuredEntities.ProblemNotifications.Message = msg
	}
	dt.configuredEntities.ProblemNotifications.Success = true
	dt.configuredEntities.ProblemNotifications.Message = "Successfully set up Keptn Alerting Profile and Problem Notifications"
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
		err = checkForUnexpectedHTMLResponseError(err)
		return "", fmt.Errorf("failed to unmarshal alerting profile: %v", err)
	}
	dt.Logger.Info("Alerting profile created successfully.")
	return createdItem.ID, nil
}

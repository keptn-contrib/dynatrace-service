package lib

import (
	"encoding/json"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
)

func (dt *DynatraceHelper) EnsureProblemNotificationsAreSetUp() error {
	dt.Logger.Info("Setting up problem notifications in Dynatrace Tenant")

	alertingProfileId, err := dt.setupAlertingProfile()
	if err != nil {
		dt.Logger.Error("could not set up problem notification: " + err.Error())
		return err
	}

	response, err := dt.sendDynatraceAPIRequest("", "/api/config/v1/notifications", "GET", "")

	existingNotifications := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), existingNotifications)
	if err != nil {
		dt.Logger.Info("No existing Dynatrace problem notification rules found. Creating new notification.")
	}

	for _, notification := range existingNotifications.Values {
		if notification.Name == "Keptn Problem Notification" {
			_, _ = dt.sendDynatraceAPIRequest("", "/api/config/v1/notifications/"+notification.ID, "DELETE", "")

		}
	}
	problemNotification := PROBLEM_NOTIFICATION_PAYLOAD

	keptnCredentials, err := common.GetKeptnCredentials()

	if err != nil {
		dt.Logger.Error("Could not retrieve Keptn API credentials: " + err.Error())
		return err
	}

	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_DNS", keptnCredentials.ApiURL)
	problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_TOKEN", keptnCredentials.ApiToken)
	problemNotification = strings.ReplaceAll(problemNotification, "$ALERTING_PROFILE_ID", alertingProfileId)

	_, err = dt.sendDynatraceAPIRequest("", "/api/config/v1/notifications", "POST", problemNotification)
	if err != nil {
		dt.Logger.Error("could not set up problem notification: " + err.Error())
		return err
	}
	return nil
}

func (dt *DynatraceHelper) setupAlertingProfile() (string, error) {
	dt.Logger.Info("Checking Keptn alerting profile availability")
	response, err := dt.sendDynatraceAPIRequest("", "/api/config/v1/alertingProfiles", "GET", "")

	existingAlertingProfiles := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), existingAlertingProfiles)
	if err != nil {
		dt.Logger.Info("No existing alerting profiles found.")
	}

	for _, ap := range existingAlertingProfiles.Values {
		if ap.Name == "Keptn" {
			dt.Logger.Info("Keptn alerting profile available")
			return ap.ID, nil
		}
	}

	dt.Logger.Info("Creating Keptn alerting profile.")

	alertingProfile := CreateKeptnAlertingProfile()

	alertingProfilePayload, _ := json.Marshal(alertingProfile)

	response, err = dt.sendDynatraceAPIRequest("", "/api/config/v1/alertingProfiles", "POST", string(alertingProfilePayload))

	if err != nil {
		return "", err
	}

	createdItem := &Values{}

	err = json.Unmarshal([]byte(response), createdItem)
	if err != nil {
		dt.Logger.Error("Could not create alerting profile: " + err.Error())
		return "", err
	}
	dt.Logger.Info("Alerting profile created successfully.")
	return createdItem.ID, nil
}

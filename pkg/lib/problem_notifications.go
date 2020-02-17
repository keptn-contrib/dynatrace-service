package lib

import (
	"encoding/json"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (dt *DynatraceHelper) EnsureProblemNotificationsAreSetUp() error {
	dt.Logger.Info("Setting up problem notifications in Dynatrace Tenant")

	alertingProfileId, err := dt.setupAlertingProfile()
	if err != nil {
		dt.Logger.Error("could not set up problem notification: " + err.Error())
		return err
	}

	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/notifications", "GET", "")

	existingNotifications := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), existingNotifications)
	if err != nil {
		dt.Logger.Info("No existing Dynatrace problem notifications rules found. Creating new notification.")
	}

	found := false
	for _, notification := range existingNotifications.Values {
		if notification.Name == "Keptn Problem Notification" {
			found = true
		}
	}
	if !found {
		problemNotification := PROBLEM_NOTIFICATION_PAYLOAD
		keptnDomainCM, err := dt.KubeApi.CoreV1().ConfigMaps("keptn").Get("keptn-domain", metav1.GetOptions{})
		if err != nil {
			dt.Logger.Error("Could not retrieve keptn-domain ConfigMap: " + err.Error())
		}

		keptnDomain := keptnDomainCM.Data["app_domain"]

		problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_DNS", "https://api.keptn."+keptnDomain)

		keptnSecret, err := dt.KubeApi.CoreV1().Secrets("keptn").Get("keptn-api-token", metav1.GetOptions{})
		if err != nil {
			dt.Logger.Error("Could not retrieve keptn-api-token: " + err.Error())
		}

		apiToken := keptnSecret.Data["keptn-api-token"]

		problemNotification = strings.ReplaceAll(problemNotification, "$KEPTN_TOKEN", string(apiToken))

		problemNotification = strings.ReplaceAll(problemNotification, "$ALERTING_PROFILE_ID", alertingProfileId)

		_, err = dt.sendDynatraceAPIRequest("/api/config/v1/notifications", "POST", problemNotification)
		if err != nil {
			dt.Logger.Error("could not set up problem notification: " + err.Error())
			return err
		}
	}
	return nil
}

func (dt *DynatraceHelper) setupAlertingProfile() (string, error) {
	dt.Logger.Info("Checking Keptn alerting profile availability")
	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "GET", "")

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

	response, err = dt.sendDynatraceAPIRequest("/api/config/v1/alertingProfiles", "POST", string(alertingProfilePayload))

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

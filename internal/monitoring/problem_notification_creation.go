package monitoring

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/env"

	log "github.com/sirupsen/logrus"
)

type ProblemNotificationCreation struct {
	client *dynatrace.Client
}

func NewProblemNotificationCreation(client *dynatrace.Client) *ProblemNotificationCreation {
	return &ProblemNotificationCreation{
		client: client,
	}
}

// Create sets up/updates the DT problem notification and returns it
func (pn *ProblemNotificationCreation) Create() ConfigResult {
	if !env.IsProblemNotificationsGenerationEnabled() {
		return ConfigResult{}
	}

	log.Info("Setting up problem notifications in Dynatrace Tenant")

	alertingProfileId, err := getOrCreateKeptnAlertingProfile(
		dynatrace.NewAlertingProfilesClient(pn.client))
	if err != nil {
		log.WithError(err).Error("Failed to set up problem notification")
		return ConfigResult{
			Success: false,
			Message: "failed to set up problem notification: " + err.Error(),
		}
	}

	notificationsClient := dynatrace.NewNotificationsClient(pn.client)
	err = notificationsClient.DeleteExistingKeptnProblemNotifications()
	if err != nil {
		log.WithError(err).Error("failed to delete existing notifications")
	}

	keptnCredentials, err := credentials.GetKeptnCredentials()
	if err != nil {
		log.WithError(err).Error("Failed to retrieve Keptn API credentials")
		return ConfigResult{
			Success: false,
			Message: "failed to retrieve Keptn API credentials: " + err.Error(),
		}
	}

	err = notificationsClient.Create(keptnCredentials, alertingProfileId)
	if err != nil {
		log.WithError(err).Error("Failed to create problem notification")
		return ConfigResult{
			Success: false,
			Message: "failed to set up problem notification: " + err.Error(),
		}
	}

	return ConfigResult{
		Success: true,
		Message: "Successfully set up Keptn Alerting Profile and Problem Notifications",
	}
}

func getOrCreateKeptnAlertingProfile(alertingProfilesClient *dynatrace.AlertingProfilesClient) (string, error) {
	log.Info("Checking Keptn alerting profile availability")
	alertingProfileID, err := alertingProfilesClient.GetProfileID("Keptn")
	if err != nil {
		log.WithError(err).Error("Could not get alerting profiles")
	}
	if alertingProfileID != "" {
		log.Info("Keptn alerting profile available")
		return alertingProfileID, nil
	}

	log.Info("Creating Keptn alerting profile.")
	alertingProfile := createKeptnAlertingProfile()
	profileID, err := alertingProfilesClient.Create(alertingProfile)
	if err != nil {
		return "", fmt.Errorf("failed to create Keptn alerting profile: %v", err)
	}

	log.Info("Alerting profile created successfully.")
	return profileID, nil
}

func createKeptnAlertingProfile() *dynatrace.AlertingProfile {
	return &dynatrace.AlertingProfile{
		Metadata:    dynatrace.AlertingProfileMetadata{},
		DisplayName: "Keptn",
		Rules: []dynatrace.AlertingProfileRules{
			createAlertingProfileRule("AVAILABILITY"),
			createAlertingProfileRule("ERROR"),
			createAlertingProfileRule("PERFORMANCE"),
			createAlertingProfileRule("RESOURCE_CONTENTION"),
			createAlertingProfileRule("CUSTOM_ALERT"),
			createAlertingProfileRule("MONITORING_UNAVAILABLE"),
		},
		ManagementZoneID: nil,
	}
}

func createAlertingProfileRule(severityLevel string) dynatrace.AlertingProfileRules {
	return dynatrace.AlertingProfileRules{
		SeverityLevel: severityLevel,
		TagFilter: dynatrace.AlertingProfileTagFilter{
			IncludeMode: "NONE",
			TagFilters:  nil,
		},
		DelayInMinutes: 0,
	}
}

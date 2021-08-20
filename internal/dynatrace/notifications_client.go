package dynatrace

import (
	"encoding/json"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"strings"
)

const keptnProblemNotificationName = "Keptn Problem Notification"

const problemNotificationPayload string = `{ 
      "type": "WEBHOOK", 
      "name": "$KEPTN_PROBLEM_NOTIFICATION_NAME", 
      "alertingProfile": "$ALERTING_PROFILE_ID", 
      "active": true, 
      "url": "$KEPTN_DNS/v1/event", 
      "acceptAnyCertificate": true, 
      "headers": [ 
        { "name": "x-token", "value": "$KEPTN_TOKEN" },
        { "name": "Content-Type", "value": "application/cloudevents+json" }
      ],
      "payload": "{\n    \"specversion\":\"1.0\",\n    \"type\":\"sh.keptn.events.problem\",\n    \"shkeptncontext\":\"{PID}\",\n    \"source\":\"dynatrace\",\n    \"id\":\"{PID}\",\n    \"time\":\"\",\n    \"contenttype\":\"application/json\",\n    \"data\": {\n        \"State\":\"{State}\",\n        \"ProblemID\":\"{ProblemID}\",\n        \"PID\":\"{PID}\",\n        \"ProblemTitle\":\"{ProblemTitle}\",\n        \"ProblemURL\":\"{ProblemURL}\",\n        \"ProblemDetails\":{ProblemDetailsJSON},\n        \"Tags\":\"{Tags}\",\n        \"ImpactedEntities\":{ImpactedEntities},\n        \"ImpactedEntity\":\"{ImpactedEntity}\"\n    }\n}\n" 

      }`

type NotificationsError struct {
	errors []error
}

func (ne *NotificationsError) HasErrors() bool {
	return len(ne.errors) > 0
}

func (ne *NotificationsError) Error() string {
	sb := strings.Builder{}
	for i, err := range ne.errors {
		sb.WriteString(err.Error())
		if i < len(ne.errors)-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}

const notificationsPath = "/api/config/v1/notifications"

type NotificationsClient struct {
	client *DynatraceHelper
}

func NewNotificationsClient(client *DynatraceHelper) *NotificationsClient {
	return &NotificationsClient{
		client: client,
	}
}

func (nc *NotificationsClient) getAll() (*DTAPIListResponse, error) {
	response, err := nc.client.Get(notificationsPath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve notifications: %v", err)
	}

	existingNotifications := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(response), existingNotifications)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal notifications: %v", err)
	}

	return existingNotifications, nil
}

func (nc *NotificationsClient) DeleteExistingKeptnProblemNotifications() error {
	existingNotifications, err := nc.getAll()
	if err != nil {
		return fmt.Errorf("failed to retrieve notifications: %v", err)
	}

	notificationError := &NotificationsError{}
	for _, notification := range existingNotifications.Values {
		if notification.Name == keptnProblemNotificationName {
			_, err := nc.deleteBy(notification.ID)
			if err != nil {
				// Error occurred but continue
				notificationError.errors = append(
					notificationError.errors,
					fmt.Errorf("failed to delete notification with ID: %s", notification.ID))
			}
		}
	}

	if notificationError.HasErrors() {
		return notificationError
	}

	return nil
}

// Create creates a new default notification for the given KeptnAPICredentials and the alertingProfileID
func (nc *NotificationsClient) Create(credentials *credentials.KeptnAPICredentials, alertingProfileID string) (string, error) {
	notification := problemNotificationPayload
	notification = strings.ReplaceAll(notification, "$KEPTN_DNS", credentials.APIURL)
	notification = strings.ReplaceAll(notification, "$KEPTN_TOKEN", credentials.APIToken)
	notification = strings.ReplaceAll(notification, "$ALERTING_PROFILE_ID", alertingProfileID)
	notification = strings.ReplaceAll(notification, "$KEPTN_PROBLEM_NOTIFICATION_NAME", keptnProblemNotificationName)

	res, err := nc.client.Post(notificationsPath, []byte(notification))
	if err != nil {
		return "", err
	}

	return res, nil
}

func (nc *NotificationsClient) deleteBy(id string) (string, error) {
	res, err := nc.client.Delete(notificationsPath + "/" + id)
	if err != nil {
		return "", nil
	}

	return res, nil
}

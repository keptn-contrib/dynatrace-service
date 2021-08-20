package dynatrace

import (
	"encoding/json"
	"fmt"
)

const alertingProfilesPath = "/api/config/v1/alertingProfiles"

type AlertingProfile struct {
	Metadata         AlertingProfileMetadata           `json:"metadata"`
	ID               string                            `json:"id"`
	DisplayName      string                            `json:"displayName"`
	Rules            []AlertingProfileRules            `json:"rules"`
	ManagementZoneID interface{}                       `json:"managementZoneId"`
	EventTypeFilters []*AlertingProfileEventTypeFilter `json:"eventTypeFilters,omitempty"`
}
type AlertingProfileMetadata struct {
	ConfigurationVersions []int  `json:"configurationVersions"`
	ClusterVersion        string `json:"clusterVersion"`
}
type AlertingProfileTagFilter struct {
	IncludeMode string   `json:"includeMode"`
	TagFilters  []string `json:"tagFilters"`
}
type AlertingProfileRules struct {
	SeverityLevel  string                   `json:"severityLevel"`
	TagFilter      AlertingProfileTagFilter `json:"tagFilter"`
	DelayInMinutes int                      `json:"delayInMinutes"`
}

type AlertingProfileEventTypeFilter struct {
	CustomEventFilter CustomEventFilter `json:"customEventFilter"`
}
type CustomTitleFilter struct {
	Enabled         bool   `json:"enabled"`
	Value           string `json:"value"`
	Operator        string `json:"operator"`
	Negate          bool   `json:"negate"`
	CaseInsensitive bool   `json:"caseInsensitive"`
}
type CustomEventFilter struct {
	CustomTitleFilter CustomTitleFilter `json:"customTitleFilter"`
}

type AlertingProfilesClient struct {
	client *DynatraceHelper
}

func NewAlertingProfilesClient(client *DynatraceHelper) *AlertingProfilesClient {
	return &AlertingProfilesClient{
		client: client,
	}
}

func (apc *AlertingProfilesClient) getAll() (*DTAPIListResponse, error) {
	response, err := apc.client.Get(notificationsPath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve alerting profiles: %v", err)
	}

	alertingProfiles := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(response), alertingProfiles)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal alerting profiles: %v", err)
	}

	return alertingProfiles, nil
}

// GetProfileIDFor returns the profile ID for the given profileName if found, an empty string otherwise
func (apc *AlertingProfilesClient) GetProfileIDFor(profileName string) (string, error) {
	res, err := apc.getAll()
	if err != nil {
		return "", err
	}

	for _, ap := range res.Values {
		if ap.Name == profileName {
			return ap.ID, nil
		}
	}

	return "", nil
}

func (apc *AlertingProfilesClient) Create(alertingProfile *AlertingProfile) (string, error) {
	alertingProfilePayload, err := json.Marshal(alertingProfile)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alerting profile: %v", err)
	}

	response, err := apc.client.Post(alertingProfilesPath, alertingProfilePayload)
	if err != nil {
		return "", fmt.Errorf("failed to setup alerting profile: %v", err)
	}

	createdItem := &Values{}
	err = json.Unmarshal([]byte(response), createdItem)
	if err != nil {
		err = CheckForUnexpectedHTMLResponseError(err)
		return "", fmt.Errorf("failed to unmarshal alerting profile: %v", err)
	}

	return createdItem.ID, nil
}

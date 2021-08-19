package dynatrace

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

const autoTagsPath = "/api/config/v1/autoTags"

type AutoTagsClient struct {
	client *DynatraceHelper
}

func NewAutoTagClient(client *DynatraceHelper) *AutoTagsClient {
	return &AutoTagsClient{
		client: client,
	}
}

func (atc *AutoTagsClient) Create(rule *DTTaggingRule) (string, error) {
	log.WithField("name", rule.Name).Info("Creating DT tagging rule")
	payload, err := json.Marshal(rule)
	if err != nil {
		return "", err
	}
	return atc.client.Post(autoTagsPath, payload)
}

func (atc *AutoTagsClient) Get() (*DTAPIListResponse, error) {
	response, err := atc.client.Get(autoTagsPath)
	if err != nil {
		log.WithError(err).Error("Could not get existing tagging rules")
		return nil, err
	}

	existingDTRules := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(response), existingDTRules)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal Dynatrace tagging rules")
		return nil, err
	}

	return existingDTRules, nil
}

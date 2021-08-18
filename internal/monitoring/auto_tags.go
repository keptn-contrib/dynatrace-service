package monitoring

import (
	"encoding/json"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"

	log "github.com/sirupsen/logrus"
)

type AutoTagCreation struct {
	client *dynatrace.DynatraceHelper
}

func NewAutoTagCreation(client *dynatrace.DynatraceHelper) *AutoTagCreation {
	return &AutoTagCreation{
		client: client,
	}
}

// Create creates auto-tags in Dynatrace and returns the tagging rules
func (at *AutoTagCreation) Create() []dynatrace.ConfigResult {
	var taggingRules []dynatrace.ConfigResult
	if !lib.IsTaggingRulesGenerationEnabled() {
		return taggingRules
	}

	log.Info("Setting up auto-tagging rules in Dynatrace Tenant")

	response, err := at.client.SendDynatraceAPIRequest("/api/config/v1/autoTags", "GET", nil)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).Error("Could not get existing tagging rules")
	}

	existingDTRules := &dynatrace.DTAPIListResponse{}
	err = json.Unmarshal([]byte(response), existingDTRules)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).Error("Failed to unmarshal Dynatrace tagging rules")
	}

	for _, ruleName := range []string{"keptn_service", "keptn_stage", "keptn_project", "keptn_deployment"} {
		if !taggingRuleExists(ruleName, existingDTRules) {
			rule := createAutoTaggingRule(ruleName)
			err = at.createDTTaggingRule(rule)
			if err != nil {
				// Error occurred but continue
				taggingRules = append(
					taggingRules,
					dynatrace.ConfigResult{
						Name:    ruleName,
						Success: false,
						Message: "Could not create auto tagging rule: " + err.Error(),
					})
				log.WithError(err).Error("Could not create auto tagging rule")
			} else {
				taggingRules = append(
					taggingRules,
					dynatrace.ConfigResult{
						Name:    ruleName,
						Success: true,
					})
			}
		} else {
			log.WithField("ruleName", ruleName).Info("Tagging rule already exists")
			taggingRules = append(
				taggingRules,
				dynatrace.ConfigResult{
					Name:    ruleName,
					Message: "Tagging rule " + ruleName + " already exists",
					Success: true,
				})
		}
	}
	return taggingRules
}

func (at *AutoTagCreation) createDTTaggingRule(rule *dynatrace.DTTaggingRule) error {
	log.WithField("name", rule.Name).Info("Creating DT tagging rule")
	payload, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	_, err = at.client.SendDynatraceAPIRequest("/api/config/v1/autoTags", "POST", payload)
	return err
}

func taggingRuleExists(ruleName string, existingRules *dynatrace.DTAPIListResponse) bool {
	for _, rule := range existingRules.Values {
		if rule.Name == ruleName {
			return true
		}
	}
	return false
}

func createAutoTaggingRule(ruleName string) *dynatrace.DTTaggingRule {
	return &dynatrace.DTTaggingRule{
		Name: ruleName,
		Rules: []dynatrace.Rules{
			{
				Type:             "SERVICE",
				Enabled:          true,
				ValueFormat:      "{ProcessGroup:Environment:" + ruleName + "}",
				PropagationTypes: []string{"SERVICE_TO_PROCESS_GROUP_LIKE"},
				Conditions: []dynatrace.Conditions{
					{
						Key: dynatrace.Key{
							Attribute: "PROCESS_GROUP_CUSTOM_METADATA",
							DynamicKey: dynatrace.DynamicKey{
								Source: "ENVIRONMENT",
								Key:    ruleName,
							},
							Type: "PROCESS_CUSTOM_METADATA_KEY",
						},
						ComparisonInfo: dynatrace.ComparisonInfo{
							Type:          "STRING",
							Operator:      "EXISTS",
							Value:         nil,
							Negate:        false,
							CaseSensitive: nil,
						},
					},
				},
			},
		},
	}
}

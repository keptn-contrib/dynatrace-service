package lib

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// EnsureDTTaggingRulesAreSetUp ensures that the tagging rules are set up
func (dt *DynatraceHelper) EnsureDTTaggingRulesAreSetUp() {
	if !IsTaggingRulesGenerationEnabled() {
		return
	}

	log.Info("Setting up auto-tagging rules in Dynatrace Tenant")

	response, err := dt.sendDynatraceAPIRequest("/api/config/v1/autoTags", "GET", nil)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).Error("Could not get existing tagging rules")
	}

	existingDTRules := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), existingDTRules)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).Error("Failed to unmarshal Dynatrace tagging rules")
	}

	for _, ruleName := range []string{"keptn_service", "keptn_stage", "keptn_project", "keptn_deployment"} {
		if !dt.taggingRuleExists(ruleName, existingDTRules) {
			rule := createAutoTaggingRule(ruleName)
			err = dt.createDTTaggingRule(rule)
			if err != nil {
				// Error occurred but continue
				dt.configuredEntities.TaggingRules = append(dt.configuredEntities.TaggingRules, ConfigResult{
					Name:    ruleName,
					Success: false,
					Message: "Could not create auto tagging rule: " + err.Error(),
				})
				log.WithError(err).Error("Could not create auto tagging rule")
			} else {
				dt.configuredEntities.TaggingRules = append(dt.configuredEntities.TaggingRules, ConfigResult{
					Name:    ruleName,
					Success: true,
				})
			}
		} else {
			log.WithField("ruleName", ruleName).Info("Tagging rule already exists")
			dt.configuredEntities.TaggingRules = append(dt.configuredEntities.TaggingRules, ConfigResult{
				Name:    ruleName,
				Message: "Tagging rule " + ruleName + " already exists",
				Success: true,
			})
		}
	}
	return
}

func (dt *DynatraceHelper) createDTTaggingRule(rule *DTTaggingRule) error {
	log.WithField("name", rule.Name).Info("Creating DT tagging rule")
	payload, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	_, err = dt.sendDynatraceAPIRequest("/api/config/v1/autoTags", "POST", payload)
	return err
}

func (dt *DynatraceHelper) taggingRuleExists(ruleName string, existingRules *DTAPIListResponse) bool {
	for _, rule := range existingRules.Values {
		if rule.Name == ruleName {
			return true
		}
	}
	return false
}

func createAutoTaggingRule(ruleName string) *DTTaggingRule {
	return &DTTaggingRule{
		Name: ruleName,
		Rules: []Rules{
			{
				Type:             "SERVICE",
				Enabled:          true,
				ValueFormat:      "{ProcessGroup:Environment:" + ruleName + "}",
				PropagationTypes: []string{"SERVICE_TO_PROCESS_GROUP_LIKE"},
				Conditions: []Conditions{
					{
						Key: Key{
							Attribute: "PROCESS_GROUP_CUSTOM_METADATA",
							DynamicKey: DynamicKey{
								Source: "ENVIRONMENT",
								Key:    ruleName,
							},
							Type: "PROCESS_CUSTOM_METADATA_KEY",
						},
						ComparisonInfo: ComparisonInfo{
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

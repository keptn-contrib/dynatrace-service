package lib

import "encoding/json"

func (dt *DynatraceHelper) EnsureDTTaggingRulesAreSetUp() error {
	if !GetTaggingRulesConfig() {
		return nil
	}

	dt.Logger.Info("Setting up auto-tagging rules in Dynatrace Tenant")

	response, err := dt.sendDynatraceAPIRequest("", "/api/config/v1/autoTags", "GET", "")

	existingDTRules := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(response), existingDTRules)
	if err != nil {
		dt.Logger.Info("No existing Dynatrace tagging rules found")
	}

	for _, ruleName := range []string{"keptn_service", "keptn_stage", "keptn_project", "keptn_deployment"} {
		if !dt.taggingRuleExists(ruleName, existingDTRules) {
			rule := createAutoTaggingRule(ruleName)
			err = dt.createDTTaggingRule(rule)
			if err != nil {
				dt.Logger.Error("Could not create auto tagging rule: " + err.Error())
			}
		} else {
			dt.Logger.Info("Tagging rule " + ruleName + " already exists")
		}
	}
	return nil
}

func (dt *DynatraceHelper) createDTTaggingRule(rule *DTTaggingRule) error {
	dt.Logger.Info("Creating DT tagging rule: " + rule.Name)
	payload, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	_, err = dt.sendDynatraceAPIRequest("", "/api/config/v1/autoTags", "POST", string(payload))
	return err
}

func (dt *DynatraceHelper) deleteExistingDTTaggingRule(ruleName string, existingRules *DTAPIListResponse) {
	dt.Logger.Info("Deleting rule " + ruleName)
	for _, rule := range existingRules.Values {
		if rule.Name == ruleName {
			_, err := dt.sendDynatraceAPIRequest("", "/api/config/v1/autoTags/"+rule.ID, "DELETE", "")
			if err != nil {
				dt.Logger.Info("Could not delete rule " + rule.ID + ": " + err.Error())
			}
		}
	}
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

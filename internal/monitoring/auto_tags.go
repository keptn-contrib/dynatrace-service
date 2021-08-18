package monitoring

import (
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

	autoTagsClient := dynatrace.NewAutoTagClient(at.client)
	existingDTRules, err := autoTagsClient.Get()
	if err != nil {
		// Error occurred but continue
		// TODO 2021-08-18: should this error just be ignored?
		log.WithError(err).Error("Failed retrieving Dynatrace tagging rules")
	}

	for _, ruleName := range []string{"keptn_service", "keptn_stage", "keptn_project", "keptn_deployment"} {
		if !taggingRuleExists(ruleName, existingDTRules) {
			rule := createAutoTaggingRule(ruleName)
			_, err = autoTagsClient.Create(rule)
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

func taggingRuleExists(ruleName string, existingRules *dynatrace.DTAPIListResponse) bool {
	if existingRules == nil {
		return false
	}

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

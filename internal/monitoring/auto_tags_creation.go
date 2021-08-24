package monitoring

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"

	log "github.com/sirupsen/logrus"
)

type AutoTagCreation struct {
	client *dynatrace.Client
}

func NewAutoTagCreation(client *dynatrace.Client) *AutoTagCreation {
	return &AutoTagCreation{
		client: client,
	}
}

// Create creates auto-tags in Dynatrace and returns the tagging rules
func (at *AutoTagCreation) Create() []dynatrace.ConfigResult {
	if !lib.IsTaggingRulesGenerationEnabled() {
		return nil
	}

	log.Info("Setting up auto-tagging rules in Dynatrace Tenant")

	autoTagsClient := dynatrace.NewAutoTagClient(at.client)
	existingDTRuleNames, err := autoTagsClient.GetAllTagNames()
	if err != nil {
		// Error occurred but continue
		// TODO 2021-08-18: should this error just be ignored?
		log.WithError(err).Error("Failed retrieving Dynatrace tagging rules")
	}

	var taggingRulesResults []dynatrace.ConfigResult
	for _, ruleName := range []string{"keptn_service", "keptn_stage", "keptn_project", "keptn_deployment"} {
		taggingRulesResults = append(
			taggingRulesResults,
			createAutoTaggingRuleForRuleName(autoTagsClient, existingDTRuleNames, ruleName))
	}
	return taggingRulesResults
}

func createAutoTaggingRuleForRuleName(client *dynatrace.AutoTagsClient, existingTagNames *dynatrace.TagNames, ruleName string) dynatrace.ConfigResult {
	if !existingTagNames.Contains(ruleName) {
		rule := createAutoTaggingRuleDTO(ruleName)

		_, err := client.Create(rule)
		if err != nil {
			// Error occurred but continue
			log.WithError(err).Error("Could not create auto tagging rule")
			return dynatrace.ConfigResult{
				Name:    ruleName,
				Success: false,
				Message: "Could not create auto tagging rule: " + err.Error(),
			}
		}

		return dynatrace.ConfigResult{
			Name:    ruleName,
			Success: true,
		}
	}

	log.WithField("ruleName", ruleName).Info("Tagging rule already exists")
	return dynatrace.ConfigResult{
		Name:    ruleName,
		Message: "Tagging rule " + ruleName + " already exists",
		Success: true,
	}
}

func createAutoTaggingRuleDTO(ruleName string) *dynatrace.DTTaggingRule {
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

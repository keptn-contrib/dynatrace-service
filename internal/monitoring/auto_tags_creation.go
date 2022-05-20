package monitoring

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	log "github.com/sirupsen/logrus"
)

type autoTagCreation struct {
	client dynatrace.ClientInterface
}

func newAutoTagCreation(client dynatrace.ClientInterface) *autoTagCreation {
	return &autoTagCreation{
		client: client,
	}
}

// create creates auto-tags in Dynatrace and returns the tagging rules.
func (at *autoTagCreation) create(ctx context.Context) []configResult {
	log.Info("Setting up auto-tagging rules in Dynatrace Tenant")

	autoTagsClient := dynatrace.NewAutoTagClient(at.client)
	existingDTRuleNames, err := autoTagsClient.GetAllTagNames(ctx)
	if err != nil {
		// Error occurred but continue
		// TODO 2021-08-18: should this error just be ignored?
		log.WithError(err).Error("Failed retrieving Dynatrace tagging rules")
	}

	var taggingRulesResults []configResult
	for _, ruleName := range []string{"keptn_service", "keptn_stage", "keptn_project", "keptn_deployment"} {
		taggingRulesResults = append(
			taggingRulesResults,
			createAutoTaggingRuleForRuleName(ctx, autoTagsClient, existingDTRuleNames, ruleName))
	}
	return taggingRulesResults
}

func createAutoTaggingRuleForRuleName(ctx context.Context, client *dynatrace.AutoTagsClient, existingTagNames *dynatrace.TagNames, ruleName string) configResult {
	if !existingTagNames.Contains(ruleName) {
		rule := createAutoTaggingRuleDTO(ruleName)

		err := client.Create(ctx, rule)
		if err != nil {
			// Error occurred but continue
			log.WithError(err).Error("Could not create auto tagging rule")
			return configResult{
				Name:    ruleName,
				Success: false,
				Message: "Could not create auto tagging rule: " + err.Error(),
			}
		}

		return configResult{
			Name:    ruleName,
			Success: true,
		}
	}

	log.WithField("ruleName", ruleName).Info("Tagging rule already exists")
	return configResult{
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

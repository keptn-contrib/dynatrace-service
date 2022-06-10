package config

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"

	"gopkg.in/yaml.v2"
)

type DynatraceConfigProvider interface {
	GetDynatraceConfig(ctx context.Context, event adapter.EventContentAdapter) (*DynatraceConfig, error)
}

type DynatraceConfigGetter struct {
	resourceClient keptn.DynatraceConfigReaderInterface
}

func NewDynatraceConfigGetter(client keptn.DynatraceConfigReaderInterface) *DynatraceConfigGetter {
	return &DynatraceConfigGetter{
		resourceClient: client,
	}
}

// GetDynatraceConfig loads the dynatrace.conf.yaml from the GIT repo
func (d *DynatraceConfigGetter) GetDynatraceConfig(ctx context.Context, event adapter.EventContentAdapter) (*DynatraceConfig, error) {

	fileContent, err := d.resourceClient.GetDynatraceConfig(ctx, event.GetProject(), event.GetStage(), event.GetService())
	if err != nil {
		return nil, err
	}

	// unmarshal the file
	dynatraceConfig, err := parseDynatraceConfigYAML(fileContent)
	if err != nil {
		return nil, err
	}

	dynatraceConfig = replacePlaceholdersInDynatraceConfig(dynatraceConfig, event)

	if dynatraceConfig.AttachRules == nil {
		dynatraceConfig.AttachRules = createDefaultAttachRules(event)
	}

	return dynatraceConfig, nil
}

func replacePlaceholdersInDynatraceConfig(dynatraceConfig *DynatraceConfig, event adapter.EventContentAdapter) *DynatraceConfig {
	return &DynatraceConfig{
		SpecVersion: dynatraceConfig.SpecVersion,
		DtCreds:     common.ReplaceKeptnPlaceholders(dynatraceConfig.DtCreds, event),
		Dashboard:   common.ReplaceKeptnPlaceholders(dynatraceConfig.Dashboard, event),
		AttachRules: replacePlaceholdersInAttachRules(dynatraceConfig.AttachRules, event),
	}
}

func replacePlaceholdersInAttachRules(attachRules *dynatrace.AttachRules, event adapter.EventContentAdapter) *dynatrace.AttachRules {
	if attachRules == nil {
		return nil
	}

	tagRulesWithReplacedPlaceholders := make([]dynatrace.TagRule, 0, len(attachRules.TagRule))
	for _, tagRule := range attachRules.TagRule {
		tagRulesWithReplacedPlaceholders = append(tagRulesWithReplacedPlaceholders, replacePlaceholdersInTagRule(tagRule, event))
	}

	return &dynatrace.AttachRules{
		TagRule: tagRulesWithReplacedPlaceholders,
	}
}

func replacePlaceholdersInTagRule(tagRule dynatrace.TagRule, event adapter.EventContentAdapter) dynatrace.TagRule {
	meTypesWithReplacedPlaceholders := make([]string, 0, len(tagRule.MeTypes))
	for _, meType := range tagRule.MeTypes {
		meTypesWithReplacedPlaceholders = append(meTypesWithReplacedPlaceholders, common.ReplaceKeptnPlaceholders(meType, event))
	}

	tagsWithReplacedPlaceholders := make([]dynatrace.TagEntry, 0, len(tagRule.Tags))
	for _, tag := range tagRule.Tags {
		tagsWithReplacedPlaceholders = append(tagsWithReplacedPlaceholders, replacePlaceholdersInTagEntry(tag, event))
	}

	return dynatrace.TagRule{
		MeTypes: meTypesWithReplacedPlaceholders,
		Tags:    tagsWithReplacedPlaceholders,
	}
}

func replacePlaceholdersInTagEntry(tag dynatrace.TagEntry, event adapter.EventContentAdapter) dynatrace.TagEntry {
	return dynatrace.TagEntry{
		Context: common.ReplaceKeptnPlaceholders(tag.Context, event),
		Key:     common.ReplaceKeptnPlaceholders(tag.Key, event),
		Value:   common.ReplaceKeptnPlaceholders(tag.Value, event),
	}
}

func parseDynatraceConfigYAML(input string) (*DynatraceConfig, error) {
	dynatraceConfig := NewDynatraceConfigWithDefaults()
	err := yaml.Unmarshal([]byte(input), dynatraceConfig)
	if err != nil {
		return nil, common.NewUnmarshalYAMLError("Dynatrace config", err)
	}

	return dynatraceConfig, nil
}

func createDefaultAttachRules(a adapter.EventContentAdapter) *dynatrace.AttachRules {
	return &dynatrace.AttachRules{
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   a.GetProject(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   a.GetStage(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   a.GetService(),
					},
				},
			},
		},
	}
}

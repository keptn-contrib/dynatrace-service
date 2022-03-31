package config

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"

	"gopkg.in/yaml.v2"
)

type DynatraceConfigProvider interface {
	GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfig, error)
}

type DynatraceConfigGetter struct {
	resourceClient keptn.DynatraceConfigResourceClientInterface
}

func NewDynatraceConfigGetter(client keptn.DynatraceConfigResourceClientInterface) *DynatraceConfigGetter {
	return &DynatraceConfigGetter{
		resourceClient: client,
	}
}

// GetDynatraceConfig loads the dynatrace.conf.yaml from the GIT repo
func (d *DynatraceConfigGetter) GetDynatraceConfig(event adapter.EventContentAdapter) (*DynatraceConfig, error) {

	fileContent, err := d.resourceClient.GetDynatraceConfig(event.GetProject(), event.GetStage(), event.GetService())
	if err != nil {
		return nil, err
	}

	// unmarshal the file
	dynatraceConfig, err := parseDynatraceConfigYAML(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dynatrace config file found for service %s in stage %s in project %s: %s", event.GetService(), event.GetStage(), event.GetProject(), err.Error())
	}

	return replacePlaceholdersInDynatraceConfig(dynatraceConfig, event), nil
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
		return nil, err
	}

	return dynatraceConfig, nil
}

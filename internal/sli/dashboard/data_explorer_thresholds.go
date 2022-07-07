package dashboard

import (
	"errors"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"golang.org/x/exp/slices"
)

type passAndWarningCriteria struct {
	pass    keptnapi.SLOCriteria
	warning keptnapi.SLOCriteria
}

var passColors = []string{
	"#006613",
	"#1f7e1e",
	"#5ead35",
	"#7dc540",
	"#9cd575",
	"#e8f9dc",
	"#048855",
	"#009e60",
	"#2ab06f",
	"#54c27d",
	"#99dea8",
	"#e1f7dc",
}

var warnColors = []string{
	"#ef651f",
	"#fd8232",
	"#ffa86c",
	"#ffd0ab",
	"#c9a000",
	"#e6be00",
	"#f5d30f",
	"#ffe11c",
	"#ffee7c",
	"#fff9d5",
}

var failColors = []string{
	"#93060e",
	"#ab0c17",
	"#c41425",
	"#dc172a",
	"#f28289",
	"#ffeaea",
}

func isPassColor(color string) bool {
	return slices.Contains(passColors, color)
}

func isPassRule(rule dynatrace.Rule) bool {
	return isPassColor(rule.Color) && rule.Value != nil
}

func isWarnColor(color string) bool {
	return slices.Contains(warnColors, color)
}

func isWarnRule(rule dynatrace.Rule) bool {
	return isWarnColor(rule.Color) && rule.Value != nil
}

func isFailColor(color string) bool {
	return slices.Contains(failColors, color)
}

func isFailRule(rule dynatrace.Rule) bool {
	return isFailColor(rule.Color) && rule.Value != nil
}

// tryGetThresholdPassAndWarningCriteria tries to get pass and warning criteria defined using the thresholds placed on a Data Explorer tile.
// It returns either the criteria and no error (conversion succeeded), nil for the criteria and no error (no threshold set), or nil for the criteria and an error (conversion failed).
func tryGetThresholdPassAndWarningCriteria(tile *dynatrace.Tile) (*passAndWarningCriteria, error) {
	if tile.VisualConfig == nil {
		return nil, nil
	}

	visualConfig := tile.VisualConfig
	if len(visualConfig.Thresholds) == 0 {
		return nil, nil
	}

	if len(visualConfig.Thresholds) > 1 {
		return nil, errors.New("Too many threshold configurations")
	}

	thresholdConfiguration := &visualConfig.Thresholds[0]
	if !areThresholdsEnabled(thresholdConfiguration) {
		return nil, nil
	}

	return parseThresholds(thresholdConfiguration)
}

// areThresholdsEnabled returns true if a user has set thresholds that will be displayed.
func areThresholdsEnabled(threshold *dynatrace.Threshold) bool {
	if !threshold.Visible {
		return false
	}

	for _, rule := range threshold.Rules {
		if rule.Value != nil {
			return true
		}
	}

	return false
}

// parseThresholds parses a dashboard threshold struct and returns pass and warning SLO criteria or an error.
func parseThresholds(threshold *dynatrace.Threshold) (*passAndWarningCriteria, error) {
	if !threshold.Visible {
		return nil, errors.New("threshold is not visible")
	}

	if len(threshold.Rules) != 3 {
		return nil, errors.New("expected 3 threshold rules")
	}

	for _, rule := range threshold.Rules {
		if rule.Value == nil {
			return nil, errors.New("missing threshold value")
		}

		if !(isPassColor(rule.Color) || isWarnColor(rule.Color) || isFailColor(rule.Color)) {
			return nil, fmt.Errorf("invalid threshold color: %s", rule.Color)
		}
	}

	if criteria := tryParsePassWarnFailThresholdRules(threshold.Rules); criteria != nil {
		return criteria, nil
	}

	if criteria := tryParseFailWarnPassThresholdRules(threshold.Rules); criteria != nil {
		return criteria, nil
	}

	return nil, errors.New("invalid threshold sequence")
}

// tryParsePassWarnFailThresholdRules tries to parse a pass-warn-fail dashboard threshold struct and returns pass and warning SLO criteria or nil.
func tryParsePassWarnFailThresholdRules(rules []dynatrace.Rule) *passAndWarningCriteria {
	if len(rules) != 3 {
		return nil
	}

	if !isPassRule(rules[0]) || !isWarnRule(rules[1]) || !isFailRule(rules[2]) {
		return nil
	}

	passThreshold := *rules[0].Value
	warnThreshold := *rules[1].Value
	failThreshold := *rules[2].Value

	if passThreshold >= warnThreshold {
		return nil
	}

	if warnThreshold >= failThreshold {
		return nil
	}

	return &passAndWarningCriteria{
		pass: keptnapi.SLOCriteria{
			Criteria: []string{
				fmt.Sprintf(">=%f", passThreshold),
				fmt.Sprintf("<%f", warnThreshold),
			},
		},
		warning: keptnapi.SLOCriteria{
			Criteria: []string{
				fmt.Sprintf("<%f", failThreshold),
			},
		},
	}
}

// tryParseFailWarnPassThresholdRules tries to parse a fail-warn-pass dashboard threshold struct and returns pass and warning SLO criteria or nil.
func tryParseFailWarnPassThresholdRules(rules []dynatrace.Rule) *passAndWarningCriteria {
	if len(rules) != 3 {
		return nil
	}

	if !isFailRule(rules[0]) || !isWarnRule(rules[1]) || !isPassRule(rules[2]) {
		return nil
	}

	failThreshold := *rules[0].Value
	warnThreshold := *rules[1].Value
	passThreshold := *rules[2].Value

	if failThreshold >= warnThreshold {
		return nil
	}

	if warnThreshold >= passThreshold {
		return nil
	}

	return &passAndWarningCriteria{
		pass: keptnapi.SLOCriteria{
			Criteria: []string{
				fmt.Sprintf(">=%f", passThreshold),
			},
		},
		warning: keptnapi.SLOCriteria{
			Criteria: []string{
				fmt.Sprintf(">=%f", warnThreshold),
			},
		},
	}
}

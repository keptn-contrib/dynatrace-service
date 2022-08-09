package dashboard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
)

type passAndWarningCriteria struct {
	pass    keptnapi.SLOCriteria
	warning keptnapi.SLOCriteria
}

type thresholdColor int

const (
	unknownThresholdColor thresholdColor = 0
	passThresholdColor    thresholdColor = 1
	warnThresholdColor    thresholdColor = 2
	failThresholdColor    thresholdColor = 3
)

type thresholdColorSequence int

const (
	unknownColorSequence      thresholdColorSequence = 0
	passWarnFailColorSequence thresholdColorSequence = 1
	failWarnPassColorSequence thresholdColorSequence = 2
)

var thresholdColors = map[string]thresholdColor{
	// pass colors
	"#006613": passThresholdColor,
	"#1f7e1e": passThresholdColor,
	"#5ead35": passThresholdColor,
	"#7dc540": passThresholdColor,
	"#9cd575": passThresholdColor,
	"#e8f9dc": passThresholdColor,
	"#048855": passThresholdColor,
	"#009e60": passThresholdColor,
	"#2ab06f": passThresholdColor,
	"#54c27d": passThresholdColor,
	"#99dea8": passThresholdColor,
	"#e1f7dc": passThresholdColor,

	// warn colors
	"#ef651f": warnThresholdColor,
	"#fd8232": warnThresholdColor,
	"#ffa86c": warnThresholdColor,
	"#ffd0ab": warnThresholdColor,
	"#c9a000": warnThresholdColor,
	"#e6be00": warnThresholdColor,
	"#f5d30f": warnThresholdColor,
	"#ffe11c": warnThresholdColor,
	"#ffee7c": warnThresholdColor,
	"#fff9d5": warnThresholdColor,

	// fail colors
	"#93060e": failThresholdColor,
	"#ab0c17": failThresholdColor,
	"#c41425": failThresholdColor,
	"#dc172a": failThresholdColor,
	"#f28289": failThresholdColor,
	"#ffeaea": failThresholdColor,
}

func getColorType(c string) thresholdColor {
	v, ok := thresholdColors[c]
	if !ok {
		return unknownThresholdColor
	}

	return v
}

func getColorTypeString(colorType thresholdColor) string {
	switch colorType {
	case passThresholdColor:
		return "pass"
	case warnThresholdColor:
		return "warn"
	case failThresholdColor:
		return "fail"
	}
	return "unknown"
}

type thresholdParsingErrors struct {
	errors []error
}

func (err *thresholdParsingErrors) Error() string {
	var errStrings = make([]string, len(err.errors))
	for i, e := range err.errors {
		errStrings[i] = e.Error()
	}
	return strings.Join(errStrings, "; ")
}

type incorrectThresholdRuleCountError struct {
	count int
}

func (err *incorrectThresholdRuleCountError) Error() string {
	return fmt.Sprintf("expected 3 rules rather than %d rules", err.count)
}

type invalidThresholdColorError struct {
	position int
	color    string
}

func (err *invalidThresholdColorError) Error() string {
	return fmt.Sprintf("invalid color %s at position %d ", err.color, err.position)
}

type missingThresholdValueError struct {
	position int
}

func (err *missingThresholdValueError) Error() string {
	return fmt.Sprintf("missing value at position %d ", err.position)
}

type strictlyMonotonicallyIncreasingConstraintError struct {
	value1 float64
	value2 float64
}

func (err *strictlyMonotonicallyIncreasingConstraintError) Error() string {
	return fmt.Sprintf("values (%f %f) must increase strictly monotonically", err.value1, err.value2)
}

type invalidThresholdColorSequenceError struct {
	colorType1 thresholdColor
	colorType2 thresholdColor
	colorType3 thresholdColor
}

func (err *invalidThresholdColorSequenceError) Error() string {
	return fmt.Sprintf("invalid color sequence: %s %s %s", getColorTypeString(err.colorType1), getColorTypeString(err.colorType2), getColorTypeString(err.colorType3))
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
		return nil, errors.New("too many threshold configurations")
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
		log.Error("parseThresholds should not be called for thresholds that are not visible")
		return nil, errors.New("threshold is not visible")
	}

	err := validateThresholdRules(threshold.Rules)
	if err != nil {
		return nil, err
	}

	return convertThresholdRulesToPassAndWarningCriteria(threshold.Rules)
}

// validateThresholdRules checks that the threshold rules are complete or returns an error.
func validateThresholdRules(rules []dynatrace.ThresholdRule) error {
	var errs []error

	if len(rules) != 3 {
		// log this error as it may mean something has changed on the Data Explorer side
		log.WithField("ruleCount", len(rules)).Error("Encountered unexpected number of threshold rules")

		errs = append(errs, &incorrectThresholdRuleCountError{count: len(rules)})
	}

	for i, rule := range rules {
		if rule.Value == nil {
			errs = append(errs, &missingThresholdValueError{position: i + 1})
		}

		if getColorType(rule.Color) == unknownThresholdColor {
			errs = append(errs, &invalidThresholdColorError{color: rule.Color, position: i + 1})
		}
	}

	if len(errs) > 0 {
		return &thresholdParsingErrors{errors: errs}
	}

	return nil
}

// convertThresholdRulesToPassAndWarningCriteria converts the threshold rules to SLO pass and warning criteria or returns an error.
// Note: assumes rules have passed validateThresholdRules
func convertThresholdRulesToPassAndWarningCriteria(rules []dynatrace.ThresholdRule) (*passAndWarningCriteria, error) {
	var errs []error

	v1 := *rules[0].Value
	v2 := *rules[1].Value
	v3 := *rules[2].Value

	if v1 >= v2 {
		errs = append(errs, &strictlyMonotonicallyIncreasingConstraintError{value1: v1, value2: v2})
	}

	if v2 >= v3 {
		errs = append(errs, &strictlyMonotonicallyIncreasingConstraintError{value1: v2, value2: v3})
	}

	colorSequence, err := getThresholdColorSequence(rules)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, &thresholdParsingErrors{errors: errs}
	}

	switch colorSequence {
	case passWarnFailColorSequence:
		return convertPassWarnFailThresholdsToSLOCriteria(rules), nil
	case failWarnPassColorSequence:
		return convertFailWarnPassThresholdsToSLOCriteria(rules), nil
	}

	// log this error as this should never occur
	log.Error("Encountered unexpected threshold color sequence")
	return nil, errors.New("unable to generate SLO pass and warning criteria for color sequence")
}

// getThresholdColorSequence returns the color sequence that the thresholds follow or an error.
// Note: assumes rules have passed validateThresholdRules
func getThresholdColorSequence(rules []dynatrace.ThresholdRule) (thresholdColorSequence, error) {
	colorType1 := getColorType(rules[0].Color)
	colorType2 := getColorType(rules[1].Color)
	colorType3 := getColorType(rules[2].Color)

	if (colorType1 == passThresholdColor) && (colorType2 == warnThresholdColor) && (colorType3 == failThresholdColor) {
		return passWarnFailColorSequence, nil
	}

	if (colorType1 == failThresholdColor) && (colorType2 == warnThresholdColor) && (colorType3 == passThresholdColor) {
		return failWarnPassColorSequence, nil
	}

	return unknownColorSequence, &invalidThresholdColorSequenceError{colorType1: colorType1, colorType2: colorType2, colorType3: colorType3}
}

func convertPassWarnFailThresholdsToSLOCriteria(rules []dynatrace.ThresholdRule) *passAndWarningCriteria {
	passThreshold := *rules[0].Value
	warnThreshold := *rules[1].Value
	failThreshold := *rules[2].Value

	return &passAndWarningCriteria{
		pass: keptnapi.SLOCriteria{
			Criteria: []string{
				fmt.Sprintf(">=%f", passThreshold),
				fmt.Sprintf("<%f", warnThreshold),
			},
		},
		warning: keptnapi.SLOCriteria{
			Criteria: []string{
				fmt.Sprintf(">=%f", passThreshold),
				fmt.Sprintf("<%f", failThreshold),
			},
		},
	}
}

func convertFailWarnPassThresholdsToSLOCriteria(rules []dynatrace.ThresholdRule) *passAndWarningCriteria {
	warnThreshold := *rules[1].Value
	passThreshold := *rules[2].Value

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

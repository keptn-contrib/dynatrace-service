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

type thresholdColorType int

const (
	unknownThresholdColorType thresholdColorType = 0
	passThresholdColorType    thresholdColorType = 1
	warnThresholdColorType    thresholdColorType = 2
	failThresholdColorType    thresholdColorType = 3
)

var thresholdColors = map[string]thresholdColorType{
	// pass colors
	"#006613": passThresholdColorType,
	"#1f7e1e": passThresholdColorType,
	"#5ead35": passThresholdColorType,
	"#7dc540": passThresholdColorType,
	"#9cd575": passThresholdColorType,
	"#e8f9dc": passThresholdColorType,
	"#048855": passThresholdColorType,
	"#009e60": passThresholdColorType,
	"#2ab06f": passThresholdColorType,
	"#54c27d": passThresholdColorType,
	"#99dea8": passThresholdColorType,
	"#e1f7dc": passThresholdColorType,

	// warn colors
	"#ef651f": warnThresholdColorType,
	"#fd8232": warnThresholdColorType,
	"#ffa86c": warnThresholdColorType,
	"#ffd0ab": warnThresholdColorType,
	"#c9a000": warnThresholdColorType,
	"#e6be00": warnThresholdColorType,
	"#f5d30f": warnThresholdColorType,
	"#ffe11c": warnThresholdColorType,
	"#ffee7c": warnThresholdColorType,
	"#fff9d5": warnThresholdColorType,

	// fail colors
	"#93060e": failThresholdColorType,
	"#ab0c17": failThresholdColorType,
	"#c41425": failThresholdColorType,
	"#dc172a": failThresholdColorType,
	"#f28289": failThresholdColorType,
	"#ffeaea": failThresholdColorType,
}

func getColorType(c string) thresholdColorType {
	v, ok := thresholdColors[c]
	if !ok {
		return unknownThresholdColorType
	}

	return v
}

func (colorType thresholdColorType) String() string {
	switch colorType {
	case passThresholdColorType:
		return "pass"
	case warnThresholdColorType:
		return "warn"
	case failThresholdColorType:
		return "fail"
	}
	return "unknown"
}

type thresholdConfiguration struct {
	thresholds [3]threshold
}

type threshold struct {
	colorType thresholdColorType
	value     float64
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
	colorType1 thresholdColorType
	colorType2 thresholdColorType
	colorType3 thresholdColorType
}

func (err *invalidThresholdColorSequenceError) Error() string {
	return fmt.Sprintf("invalid color sequence: %s %s %s", err.colorType1, err.colorType2, err.colorType3)
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

	t := &visualConfig.Thresholds[0]
	if !areThresholdsEnabled(t) {
		return nil, nil
	}

	thresholdConfiguration, err := convertThresholdRulesToThresholdConfiguration(t.Rules)
	if err != nil {
		return nil, err
	}

	return convertThresholdConfigurationToPassAndWarningCriteria(*thresholdConfiguration)
}

// areThresholdsEnabled returns true if a user has set thresholds that will be displayed, i.e. if thresholds are visible and at least one value has been set.
func areThresholdsEnabled(threshold *dynatrace.VisualizationThreshold) bool {
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

// convertThresholdRulesToThresholdConfiguration checks that the threshold rules are complete and returns them as a threshold configuration or returns an error.
func convertThresholdRulesToThresholdConfiguration(rules []dynatrace.VisualizationThresholdRule) (*thresholdConfiguration, error) {
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

		if getColorType(rule.Color) == unknownThresholdColorType {
			errs = append(errs, &invalidThresholdColorError{color: rule.Color, position: i + 1})
		}
	}

	if len(errs) > 0 {
		return nil, &thresholdParsingErrors{errors: errs}
	}

	return &thresholdConfiguration{
		thresholds: [3]threshold{
			{colorType: getColorType(rules[0].Color), value: *rules[0].Value},
			{colorType: getColorType(rules[1].Color), value: *rules[1].Value},
			{colorType: getColorType(rules[2].Color), value: *rules[2].Value}}}, nil
}

func convertThresholdConfigurationToPassAndWarningCriteria(t thresholdConfiguration) (*passAndWarningCriteria, error) {
	var errs []error

	v1 := t.thresholds[0].value
	v2 := t.thresholds[1].value
	v3 := t.thresholds[2].value

	if v1 >= v2 {
		errs = append(errs, &strictlyMonotonicallyIncreasingConstraintError{value1: v1, value2: v2})
	}

	if v2 >= v3 {
		errs = append(errs, &strictlyMonotonicallyIncreasingConstraintError{value1: v2, value2: v3})
	}

	sloCriteria, err := matchThresholdColorSequenceAndConvertToPassAndWarningCriteria(t)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, &thresholdParsingErrors{errors: errs}
	}

	return sloCriteria, nil
}

func matchThresholdColorSequenceAndConvertToPassAndWarningCriteria(t thresholdConfiguration) (*passAndWarningCriteria, error) {
	colorType1 := t.thresholds[0].colorType
	colorType2 := t.thresholds[1].colorType
	colorType3 := t.thresholds[2].colorType

	if (colorType1 == passThresholdColorType) && (colorType2 == warnThresholdColorType) && (colorType3 == failThresholdColorType) {
		return convertPassWarnFailThresholdsToPassAndWarningCriteria(t), nil
	}

	if (colorType1 == failThresholdColorType) && (colorType2 == warnThresholdColorType) && (colorType3 == passThresholdColorType) {
		return convertFailWarnPassThresholdsToPassAndWarningCriteria(t), nil
	}

	return nil, &invalidThresholdColorSequenceError{colorType1: colorType1, colorType2: colorType2, colorType3: colorType3}
}

func convertPassWarnFailThresholdsToPassAndWarningCriteria(t thresholdConfiguration) *passAndWarningCriteria {
	passThreshold := t.thresholds[0].value
	warnThreshold := t.thresholds[1].value
	failThreshold := t.thresholds[2].value

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

func convertFailWarnPassThresholdsToPassAndWarningCriteria(t thresholdConfiguration) *passAndWarningCriteria {
	warnThreshold := t.thresholds[1].value
	passThreshold := t.thresholds[2].value

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

package dashboard

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
)

// passAndWarningProvider provides pass and warning criteria.
type passAndWarningProvider interface {
	getPass() []*keptnapi.SLOCriteria
	getWarning() []*keptnapi.SLOCriteria
}

type strictMonotonicityType int

const (
	notStrictlyMonotonic strictMonotonicityType = iota
	strictlyIncreasingValues
	strictlyDecreasingValues
)

type thresholdType int

const (
	noThresholdType thresholdType = iota
	passThresholdType
	warnThresholdType
	failThresholdType
	unknownThresholdType
)

var thresholdColors = map[string]thresholdType{
	// pass colors
	"#006613": passThresholdType,
	"#1f7e1e": passThresholdType,
	"#5ead35": passThresholdType,
	"#7dc540": passThresholdType,
	"#9cd575": passThresholdType,
	"#e8f9dc": passThresholdType,
	"#048855": passThresholdType,
	"#009e60": passThresholdType,
	"#2ab06f": passThresholdType,
	"#54c27d": passThresholdType,
	"#99dea8": passThresholdType,
	"#e1f7dc": passThresholdType,

	// warn colors
	"#ef651f": warnThresholdType,
	"#fd8232": warnThresholdType,
	"#ffa86c": warnThresholdType,
	"#ffd0ab": warnThresholdType,
	"#c9a000": warnThresholdType,
	"#e6be00": warnThresholdType,
	"#f5d30f": warnThresholdType,
	"#ffe11c": warnThresholdType,
	"#ffee7c": warnThresholdType,
	"#fff9d5": warnThresholdType,

	// fail colors
	"#93060e": failThresholdType,
	"#ab0c17": failThresholdType,
	"#c41425": failThresholdType,
	"#dc172a": failThresholdType,
	"#f28289": failThresholdType,
	"#ffeaea": failThresholdType,
}

func getThresholdTypeByColor(c string) thresholdType {
	v, ok := thresholdColors[c]
	if !ok {
		return unknownThresholdType
	}

	return v
}

// GetThresholdTypeNameForColor returns a string with the threshold type name for the specified color. Convienence function for use in tests.
func GetThresholdTypeNameForColor(color string) string {
	return getThresholdTypeByColor(color).String()
}

func (t thresholdType) String() string {
	switch t {
	case noThresholdType:
		return "none"
	case passThresholdType:
		return "pass"
	case warnThresholdType:
		return "warn"
	case failThresholdType:
		return "fail"
	default:
		return "unknown"
	}
}

type thresholdTypeConfiguration struct {
	thresholdTypes [3]thresholdType
}

type thresholdConfiguration struct {
	thresholds [3]threshold
}

func (tc *thresholdConfiguration) thresholdTypeConfiguration() thresholdTypeConfiguration {
	return thresholdTypeConfiguration{
		thresholdTypes: [3]thresholdType{
			tc.thresholds[0].thresholdType,
			tc.thresholds[1].thresholdType,
			tc.thresholds[2].thresholdType,
		},
	}
}

func (tc *thresholdConfiguration) reverse() {
	t := tc.thresholds[2]
	tc.thresholds[2] = tc.thresholds[0]
	tc.thresholds[0] = t
}

func (tc *thresholdConfiguration) createPassAndWarningProvider() (passAndWarningProvider, error) {
	if (tc.thresholds[0].thresholdType == passThresholdType) && (tc.thresholds[1].thresholdType == warnThresholdType) && (tc.thresholds[2].thresholdType == failThresholdType) {
		return &passWarnFailThresholdConfiguration{passValue: tc.thresholds[0].value, warnValue: tc.thresholds[1].value, failValue: tc.thresholds[2].value}, nil
	}

	if (tc.thresholds[0].thresholdType == passThresholdType) && (tc.thresholds[1].thresholdType == noThresholdType) && (tc.thresholds[2].thresholdType == failThresholdType) {
		return &passFailThresholdConfiguration{passValue: tc.thresholds[0].value, failValue: tc.thresholds[2].value}, nil
	}

	if (tc.thresholds[0].thresholdType == failThresholdType) && (tc.thresholds[1].thresholdType == warnThresholdType) && (tc.thresholds[2].thresholdType == passThresholdType) {
		return &failWarnPassThresholdConfiguration{failValue: tc.thresholds[0].value, warnValue: tc.thresholds[1].value, passValue: tc.thresholds[2].value}, nil
	}

	if (tc.thresholds[0].thresholdType == failThresholdType) && (tc.thresholds[1].thresholdType == noThresholdType) && (tc.thresholds[2].thresholdType == passThresholdType) {
		return &failPassThresholdConfiguration{failValue: tc.thresholds[0].value, passValue: tc.thresholds[2].value}, nil
	}

	return nil, &invalidThresholdTypeSequenceError{thresholdTypes: tc.thresholdTypeConfiguration()}
}

type threshold struct {
	thresholdType thresholdType
	value         float64
}

type thresholdParsingError struct {
	errors []error
}

func (err *thresholdParsingError) Error() string {
	var errStrings = make([]string, len(err.errors))
	for i, e := range err.errors {
		errStrings[i] = e.Error()
	}
	return fmt.Sprintf("error parsing thresholds: %s", strings.Join(errStrings, "; "))
}

type incorrectThresholdRuleCountError struct {
	count int
}

func (err *incorrectThresholdRuleCountError) Error() string {
	return fmt.Sprintf("expected 3 rules rather than %d rules", err.count)
}

type valueSequenceError struct{}

func (err *valueSequenceError) Error() string {
	return "values must increase or decrease strictly monotonically"
}

type invalidThresholdColorError struct {
	index int
}

func (err *invalidThresholdColorError) Error() string {
	return fmt.Sprintf("invalid color at position %d", err.index+1)
}

type missingThresholdValueError struct {
	index int
}

func (err *missingThresholdValueError) Error() string {
	return fmt.Sprintf("missing value at position %d", err.index+1)
}

type invalidThresholdTypeSequenceError struct {
	thresholdTypes thresholdTypeConfiguration
}

func (err *invalidThresholdTypeSequenceError) Error() string {
	return fmt.Sprintf("invalid sequence: %s %s %s", err.thresholdTypes.thresholdTypes[0], err.thresholdTypes.thresholdTypes[1], err.thresholdTypes.thresholdTypes[2])
}

// tryGetThresholdPassAndWarningProvider tries to get pass and warning criteria defined using the thresholds placed on a Data Explorer tile.
// It returns either a pass and warning provider and no error (conversion succeeded), nil for the provider and no error (no thresholds set), or nil for the provider and an error (conversion failed).
func tryGetThresholdPassAndWarningProvider(tile *dynatrace.Tile) (passAndWarningProvider, error) {
	thresholdRules, err := getThresholdRulesFromTile(tile)
	if err != nil {
		return nil, err
	}

	if thresholdRules == nil {
		return nil, nil
	}

	thresholdConfiguration, err := convertThresholdRulesToThresholdConfiguration(thresholdRules)
	if err != nil {
		return nil, err
	}

	return thresholdConfiguration.createPassAndWarningProvider()
}

func getThresholdRulesFromTile(tile *dynatrace.Tile) ([]dynatrace.VisualizationThresholdRule, error) {
	if tile.VisualConfig == nil {
		return nil, nil
	}

	visibleThresholdRules := make([][]dynatrace.VisualizationThresholdRule, 0, len(tile.VisualConfig.Thresholds))
	for _, t := range tile.VisualConfig.Thresholds {
		if areThresholdsVisible(t) {
			visibleThresholdRules = append(visibleThresholdRules, t.Rules)
		}
	}

	if len(visibleThresholdRules) == 0 {
		return nil, nil
	}

	if len(visibleThresholdRules) > 1 {
		return nil, fmt.Errorf("Data explorer tile has %d visible thresholds but only one is supported", len(visibleThresholdRules))
	}

	return visibleThresholdRules[0], nil
}

// areThresholdsVisible returns true if a user has set thresholds that will be displayed, i.e. if thresholds are visible and at least one value has been set.
func areThresholdsVisible(threshold dynatrace.VisualizationThreshold) bool {
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

// convertThresholdRulesToThresholdConfiguration checks that the threshold rules are complete and returns them as a monotonically increasing threshold configuration or returns an error.
func convertThresholdRulesToThresholdConfiguration(rules []dynatrace.VisualizationThresholdRule) (*thresholdConfiguration, error) {
	var tc thresholdConfiguration
	thresholdTypes, errs := getThresholdTypeConfigurationFromRules(rules)

	v0 := tryGetValueFromThresholdRules(rules, 0)
	if v0 == nil {
		errs = append(errs, &missingThresholdValueError{index: 0})
	} else {
		tc.thresholds[0] = threshold{thresholdType: thresholdTypes.thresholdTypes[0], value: *v0}
	}

	v1 := tryGetValueFromThresholdRules(rules, 1)
	if v1 == nil {
		tc.thresholds[1] = threshold{thresholdType: noThresholdType}
	} else {
		tc.thresholds[1] = threshold{thresholdType: thresholdTypes.thresholdTypes[1], value: *v1}
	}

	v2 := tryGetValueFromThresholdRules(rules, 2)
	if v2 == nil {
		errs = append(errs, &missingThresholdValueError{index: 2})
	} else {
		tc.thresholds[2] = threshold{thresholdType: thresholdTypes.thresholdTypes[2], value: *v2}
	}

	monotonicity := getStrictMonotonicityOfThreeOptionalValues(v0, v1, v2)
	if monotonicity == notStrictlyMonotonic {
		errs = append(errs, &valueSequenceError{})
	}

	if len(errs) > 0 {
		return nil, &thresholdParsingError{errors: errs}
	}

	if monotonicity == strictlyDecreasingValues {
		tc.reverse()
	}

	return &tc, nil
}

func getThresholdTypeConfigurationFromRules(rules []dynatrace.VisualizationThresholdRule) (thresholdTypeConfiguration, []error) {
	var errs []error
	var thresholdTypes thresholdTypeConfiguration

	// check that colors are set correctly
	for i, rule := range rules {
		if (i == 0) || (i == 2) {
			tt := getThresholdTypeByColor(rule.Color)
			if (tt != passThresholdType) && (tt != failThresholdType) {
				errs = append(errs, &invalidThresholdColorError{index: i})
				continue
			}
			thresholdTypes.thresholdTypes[i] = tt
		}

		if i == 1 {
			tt := getThresholdTypeByColor(rule.Color)
			if tt != warnThresholdType {
				errs = append(errs, &invalidThresholdColorError{index: i})
				continue
			}
			thresholdTypes.thresholdTypes[i] = tt
		}
	}

	if len(rules) != 3 {
		// log this error as it may mean something has changed on the Data Explorer side
		log.WithField("ruleCount", len(rules)).Error("Encountered unexpected number of threshold rules")
		errs = append(errs, &incorrectThresholdRuleCountError{count: len(rules)})
	}

	if (len(rules) >= 3) && (thresholdTypes.thresholdTypes[0] == thresholdTypes.thresholdTypes[2]) {
		errs = append(errs, &invalidThresholdTypeSequenceError{thresholdTypes: thresholdTypes})
	}

	return thresholdTypes, errs
}

func tryGetValueFromThresholdRules(rules []dynatrace.VisualizationThresholdRule, index int) *float64 {
	if index >= len(rules) {
		return nil
	}

	return rules[index].Value
}

func getStrictMonotonicityOfThreeOptionalValues(v0, v1, v2 *float64) strictMonotonicityType {
	if (v0 != nil) && (v2 != nil) {
		monotonicity := getStrictMonotonicity(*v0, *v2)

		if v1 != nil {
			if (getStrictMonotonicity(*v0, *v1) != monotonicity) || (getStrictMonotonicity(*v1, *v2) != monotonicity) {
				return notStrictlyMonotonic
			}
		}
		return monotonicity
	}

	if (v0 != nil) && (v1 != nil) {
		return getStrictMonotonicity(*v0, *v1)
	}

	if (v1 != nil) && (v2 != nil) {
		return getStrictMonotonicity(*v1, *v2)
	}

	return notStrictlyMonotonic
}

func getStrictMonotonicity(v1 float64, v2 float64) strictMonotonicityType {
	if v1 == v2 {
		return notStrictlyMonotonic
	}

	if v1 < v2 {
		return strictlyIncreasingValues
	}

	return strictlyDecreasingValues
}

type passWarnFailThresholdConfiguration struct {
	passValue float64
	warnValue float64
	failValue float64
}

func (c *passWarnFailThresholdConfiguration) getPass() []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{
		{
			Criteria: []string{
				fmt.Sprintf(">=%f", c.passValue),
				fmt.Sprintf("<%f", c.warnValue),
			},
		},
	}
}

func (c *passWarnFailThresholdConfiguration) getWarning() []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{
		{
			Criteria: []string{
				fmt.Sprintf(">=%f", c.passValue),
				fmt.Sprintf("<%f", c.failValue),
			},
		},
	}
}

type passFailThresholdConfiguration struct {
	passValue float64
	failValue float64
}

func (c *passFailThresholdConfiguration) getPass() []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{
		{
			Criteria: []string{
				fmt.Sprintf(">=%f", c.passValue),
				fmt.Sprintf("<%f", c.failValue),
			},
		},
	}
}

func (c *passFailThresholdConfiguration) getWarning() []*keptnapi.SLOCriteria {
	return nil
}

type failWarnPassThresholdConfiguration struct {
	failValue float64
	warnValue float64
	passValue float64
}

func (c *failWarnPassThresholdConfiguration) getPass() []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{
		{
			Criteria: []string{
				fmt.Sprintf(">=%f", c.passValue),
			},
		},
	}
}

func (c *failWarnPassThresholdConfiguration) getWarning() []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{
		{
			Criteria: []string{
				fmt.Sprintf(">=%f", c.warnValue),
			},
		},
	}
}

type failPassThresholdConfiguration struct {
	failValue float64
	passValue float64
}

func (c *failPassThresholdConfiguration) getPass() []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{
		{
			Criteria: []string{
				fmt.Sprintf(">=%f", c.passValue),
			},
		},
	}
}

func (c *failPassThresholdConfiguration) getWarning() []*keptnapi.SLOCriteria {
	return nil
}

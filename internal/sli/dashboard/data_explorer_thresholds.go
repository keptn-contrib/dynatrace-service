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

type monotonicityType int

const (
	unknownMonotonicity      monotonicityType = 0
	strictlyIncreasingValues monotonicityType = 1
	strictlyDecreasingValues monotonicityType = 2
	constantValues           monotonicityType = 3
)

type thresholdType int

const (
	noThresholdType      thresholdType = 0
	passThresholdType    thresholdType = 1
	warnThresholdType    thresholdType = 2
	failThresholdType    thresholdType = 3
	unknownThresholdType thresholdType = 4
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

type thresholdConfiguration struct {
	thresholds [3]threshold
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
	thresholdType1 thresholdType
	thresholdType2 thresholdType
	thresholdType3 thresholdType
}

func (err *invalidThresholdTypeSequenceError) Error() string {
	return fmt.Sprintf("invalid sequence: %s %s %s", err.thresholdType1, err.thresholdType2, err.thresholdType3)
}

// tryGetThresholdPassAndWarning tries to get pass and warning criteria defined using the thresholds placed on a Data Explorer tile.
// It returns either the criteria and no error (conversion succeeded), nil for the criteria and no error (no threshold set), or nil for the criteria and an error (conversion failed).
func tryGetThresholdPassAndWarning(tile *dynatrace.Tile) (passAndWarningProvider, error) {
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

	return matchThresholdColorSequenceAndConvertToPassAndWarningCriteria(*thresholdConfiguration)
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

// convertThresholdRulesToThresholdConfiguration checks that the threshold rules are complete and returns them as a threshold configuration or returns an error.
func convertThresholdRulesToThresholdConfiguration(rules []dynatrace.VisualizationThresholdRule) (*thresholdConfiguration, error) {
	tc := thresholdConfiguration{}
	errs := checkThresholdColorsAndWriteThemToConfiguration(rules, &tc)
	monotonicity, vErrs := checkThresholdValuesAndWriteThemToConfiguration(rules, &tc)
	errs = append(errs, vErrs...)

	if len(errs) > 0 {
		return nil, &thresholdParsingError{errors: errs}
	}

	if monotonicity == strictlyDecreasingValues {
		return &thresholdConfiguration{
			thresholds: [3]threshold{tc.thresholds[2], tc.thresholds[1], tc.thresholds[0]},
		}, nil
	}

	return &tc, nil
}

func checkThresholdColorsAndWriteThemToConfiguration(rules []dynatrace.VisualizationThresholdRule, tc *thresholdConfiguration) []error {
	var errs []error

	// check that colors are set correctly
	for i, rule := range rules {
		switch i {
		case 0, 2:
			tt := getThresholdTypeByColor(rule.Color)
			if (tt != passThresholdType) && (tt != failThresholdType) {
				errs = append(errs, &invalidThresholdColorError{index: i})
				continue
			}
			tc.thresholds[i].thresholdType = tt

		case 1:
			tt := getThresholdTypeByColor(rule.Color)
			if tt != warnThresholdType {
				errs = append(errs, &invalidThresholdColorError{index: i})
				continue
			}
			tc.thresholds[i].thresholdType = tt
		}
	}

	if len(rules) != 3 {
		// log this error as it may mean something has changed on the Data Explorer side
		log.WithField("ruleCount", len(rules)).Error("Encountered unexpected number of threshold rules")
		errs = append(errs, &incorrectThresholdRuleCountError{count: len(rules)})
	}

	if (len(rules) >= 3) && (tc.thresholds[0].thresholdType == tc.thresholds[2].thresholdType) {
		errs = append(errs, &invalidThresholdTypeSequenceError{thresholdType1: tc.thresholds[0].thresholdType, thresholdType2: tc.thresholds[1].thresholdType, thresholdType3: tc.thresholds[2].thresholdType})
	}

	return errs
}

func checkThresholdValuesAndWriteThemToConfiguration(rules []dynatrace.VisualizationThresholdRule, tc *thresholdConfiguration) (monotonicityType, []error) {
	var errs []error

	v0 := tryGetValueFromThresholdRules(rules, 0)
	if v0 == nil {
		errs = append(errs, &missingThresholdValueError{index: 0})
	} else {
		tc.thresholds[0].value = *v0
	}

	v1 := tryGetValueFromThresholdRules(rules, 1)
	if v1 == nil {
		tc.thresholds[1].thresholdType = noThresholdType
	} else {
		tc.thresholds[1].value = *v1
	}

	v2 := tryGetValueFromThresholdRules(rules, 2)
	if v2 == nil {
		errs = append(errs, &missingThresholdValueError{index: 2})
	} else {
		tc.thresholds[2].value = *v2
	}

	monotonicity := getMonotonicityOfThreeOptionalValues(v0, v1, v2)
	if monotonicity == constantValues || monotonicity == unknownMonotonicity {
		errs = append(errs, &valueSequenceError{})
	}

	return monotonicity, errs
}

func tryGetValueFromThresholdRules(rules []dynatrace.VisualizationThresholdRule, index int) *float64 {
	if index >= len(rules) {
		return nil
	}

	return rules[index].Value
}

func getMonotonicityOfThreeOptionalValues(v0, v1, v2 *float64) monotonicityType {
	if (v0 != nil) && (v2 != nil) {
		monotonicity := getMonotonicity(*v0, *v2)

		if v1 != nil {
			if (getMonotonicity(*v0, *v1) != monotonicity) || (getMonotonicity(*v1, *v2) != monotonicity) {
				return unknownMonotonicity
			}
		}
		return monotonicity
	}

	if (v0 != nil) && (v1 != nil) {
		return getMonotonicity(*v0, *v1)
	}

	if (v1 != nil) && (v2 != nil) {
		return getMonotonicity(*v1, *v2)
	}

	return unknownMonotonicity
}

func getMonotonicity(v1 float64, v2 float64) monotonicityType {
	if v1 == v2 {
		return constantValues
	}

	if v1 < v2 {
		return strictlyIncreasingValues
	}

	return strictlyDecreasingValues
}

func matchThresholdColorSequenceAndConvertToPassAndWarningCriteria(t thresholdConfiguration) (passAndWarningProvider, error) {
	if (t.thresholds[0].thresholdType == passThresholdType) && (t.thresholds[1].thresholdType == warnThresholdType) && (t.thresholds[2].thresholdType == failThresholdType) {
		return &passWarnFailThresholdConfiguration{passValue: t.thresholds[0].value, warnValue: t.thresholds[1].value, failValue: t.thresholds[2].value}, nil
	}

	if (t.thresholds[0].thresholdType == passThresholdType) && (t.thresholds[1].thresholdType == noThresholdType) && (t.thresholds[2].thresholdType == failThresholdType) {
		return &passFailThresholdConfiguration{passValue: t.thresholds[0].value, failValue: t.thresholds[2].value}, nil
	}

	if (t.thresholds[0].thresholdType == failThresholdType) && (t.thresholds[1].thresholdType == warnThresholdType) && (t.thresholds[2].thresholdType == passThresholdType) {
		return &failWarnPassThresholdConfiguration{failValue: t.thresholds[0].value, warnValue: t.thresholds[1].value, passValue: t.thresholds[2].value}, nil
	}

	if (t.thresholds[0].thresholdType == failThresholdType) && (t.thresholds[1].thresholdType == noThresholdType) && (t.thresholds[2].thresholdType == passThresholdType) {
		return &failPassThresholdConfiguration{failValue: t.thresholds[0].value, passValue: t.thresholds[2].value}, nil
	}

	return nil, &invalidThresholdTypeSequenceError{thresholdType1: t.thresholds[0].thresholdType, thresholdType2: t.thresholds[1].thresholdType, thresholdType3: t.thresholds[2].thresholdType}
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

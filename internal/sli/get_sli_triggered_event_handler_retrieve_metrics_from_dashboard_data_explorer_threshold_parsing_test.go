package sli

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/dashboard"
)

const (
	invalidThresholdColor = "#14a8f5"
	passThresholdColor    = "#7dc540"
	warnThresholdColor    = "#f5d30f"
	failThresholdColor    = "#dc172a"
)

var (
	thresholdValue0   float64 = 0
	thresholdValue100 float64 = 100
	thresholdValue200 float64 = 200
)

const (
	invalidColorErrorSubstring  = "invalid color at position "
	invalidColorErrorSubstring1 = invalidColorErrorSubstring + "1"
	invalidColorErrorSubstring2 = invalidColorErrorSubstring + "2"
	invalidColorErrorSubstring3 = invalidColorErrorSubstring + "3"

	missingValueErrorSubstring  = "missing value at position "
	missingValueErrorSubstring1 = missingValueErrorSubstring + "1"
	missingValueErrorSubstring3 = missingValueErrorSubstring + "3"

	valueSequenceErrorSubstring               = "values must increase or decrease strictly monotonically"
	invalidSequenceErrorSubstring             = "invalid sequence: "
	incorrectThresholdRuleCountErrorSubstring = "expected 3 rules"
)

// colorTestComponent specifies if the given threshold colors should result in a parsing error and lists any error substrings they are expected to produce.
type colorTestComponent struct {
	colors                  []string
	expectError             bool
	expectedErrorSubstrings []string
}

// List all possible combinations of three theshold colors, where each one may be one of three valid colors (for pass, warn or fail) or an invalid color.
// Each combination is marked as valid or invalid depending on whether it would lead to an error and lists any error message substrings the combination would produce.
// The list is ordered sequentially, i.e. AAA, BAA, CAA, DAA, ABA, BBA, CBA...
var threeColorTestComponentVariants = []colorTestComponent{
	createInvalidColorTestComponent([]string{invalidThresholdColor, invalidThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, invalidThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, invalidThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, invalidThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),

	createInvalidColorTestComponent([]string{invalidThresholdColor, passThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, passThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, passThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, passThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),

	createInvalidColorTestComponent([]string{invalidThresholdColor, warnThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, warnThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, warnThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, warnThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring3),

	createInvalidColorTestComponent([]string{invalidThresholdColor, failThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, failThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, failThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, failThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),

	// ---

	createInvalidColorTestComponent([]string{invalidThresholdColor, invalidThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{passThresholdColor, invalidThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{warnThresholdColor, invalidThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{failThresholdColor, invalidThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),

	createInvalidColorTestComponent([]string{invalidThresholdColor, passThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{passThresholdColor, passThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{warnThresholdColor, passThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{failThresholdColor, passThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),

	createInvalidColorTestComponent([]string{invalidThresholdColor, warnThresholdColor, passThresholdColor}, invalidColorErrorSubstring1),
	createInvalidColorTestComponent([]string{passThresholdColor, warnThresholdColor, passThresholdColor}, invalidSequenceErrorSubstring+"pass warn pass"),
	createInvalidColorTestComponent([]string{warnThresholdColor, warnThresholdColor, passThresholdColor}, invalidColorErrorSubstring1),
	createValidColorTestComponent([]string{failThresholdColor, warnThresholdColor, passThresholdColor}),

	createInvalidColorTestComponent([]string{invalidThresholdColor, failThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{passThresholdColor, failThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{warnThresholdColor, failThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{failThresholdColor, failThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),

	// ---

	createInvalidColorTestComponent([]string{invalidThresholdColor, invalidThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, invalidThresholdColor, warnThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, invalidThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, invalidThresholdColor, warnThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),

	createInvalidColorTestComponent([]string{invalidThresholdColor, passThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, passThresholdColor, warnThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, passThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, passThresholdColor, warnThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),

	createInvalidColorTestComponent([]string{invalidThresholdColor, warnThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, warnThresholdColor, warnThresholdColor}, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, warnThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, warnThresholdColor, warnThresholdColor}, invalidColorErrorSubstring3),

	createInvalidColorTestComponent([]string{invalidThresholdColor, failThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{passThresholdColor, failThresholdColor, warnThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{warnThresholdColor, failThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2, invalidColorErrorSubstring3),
	createInvalidColorTestComponent([]string{failThresholdColor, failThresholdColor, warnThresholdColor}, invalidColorErrorSubstring2, invalidColorErrorSubstring3),

	// ---

	createInvalidColorTestComponent([]string{invalidThresholdColor, invalidThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{passThresholdColor, invalidThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{warnThresholdColor, invalidThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{failThresholdColor, invalidThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),

	createInvalidColorTestComponent([]string{invalidThresholdColor, passThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{passThresholdColor, passThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{warnThresholdColor, passThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{failThresholdColor, passThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),

	createInvalidColorTestComponent([]string{invalidThresholdColor, warnThresholdColor, failThresholdColor}, invalidColorErrorSubstring1),
	createValidColorTestComponent([]string{passThresholdColor, warnThresholdColor, failThresholdColor}),
	createInvalidColorTestComponent([]string{warnThresholdColor, warnThresholdColor, failThresholdColor}, invalidColorErrorSubstring1),
	createInvalidColorTestComponent([]string{failThresholdColor, warnThresholdColor, failThresholdColor}, invalidSequenceErrorSubstring+"fail warn fail"),

	createInvalidColorTestComponent([]string{invalidThresholdColor, failThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{passThresholdColor, failThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{warnThresholdColor, failThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
	createInvalidColorTestComponent([]string{failThresholdColor, failThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),
}

func createValidColorTestComponent(colors []string) colorTestComponent {
	return colorTestComponent{
		colors:      colors,
		expectError: false,
	}
}

func createInvalidColorTestComponent(colors []string, expectedErrorSubstrings ...string) colorTestComponent {
	return colorTestComponent{
		colors:                  colors,
		expectError:             true,
		expectedErrorSubstrings: expectedErrorSubstrings,
	}
}

// valueTestComponent specifies if the given threshold colors should result in a parsing error and lists any error substrings they are expected to produce.
type valueTestComponent struct {
	values                  []*float64
	expectError             bool
	expectedErrorSubstrings []string
}

// Sequentially list all possible combinations of three values, where each one may be one of three sample values as well as nil.
// Each combination is marked as valid or invalid depending on whether it would lead to an error and lists any error message substrings the combination would produce.
// The list is sequential, i.e. ordered AAA, BAA, CAA, DAA, ABA, BBA, CBA..., however redundant cases have been commented out to save testing effort.
var threeValueTestComponentVariants = []valueTestComponent{

	// valid, but indicates no thresholds are to be used: createValidValueTestComponent([]*float64{nil, nil, nil}),  // combination 1
	createInvalidValueTestComponent([]*float64{&thresholdValue0, nil, nil}, missingValueErrorSubstring3), // combination 2
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, nil, nil}, missingValueErrorSubstring3),  // combination 3: redudant due to combination 2
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, nil, nil}, missingValueErrorSubstring3), // combination 4: redudant due to combination 2

	createInvalidValueTestComponent([]*float64{nil, &thresholdValue0, nil}, missingValueErrorSubstring1, missingValueErrorSubstring3),              // combination  5
	createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue0, nil}, missingValueErrorSubstring3, valueSequenceErrorSubstring), // combination 6
	createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue0, nil}, missingValueErrorSubstring3),                            // combination 7
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue0, nil}, missingValueErrorSubstring3),  // combination 8: redudant due to combination 7

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue100, nil}, missingValueErrorSubstring1, missingValueErrorSubstring3),  // combination 9: redudant due to combination 5
	createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue100, nil}, missingValueErrorSubstring3), // combination 10
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue100, nil}, missingValueErrorSubstring3, valueSequenceErrorSubstring), // combination 11: redudant due to combination 6
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue100, nil}, missingValueErrorSubstring3), //combination 12: redudant due to combination 7

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue200, nil}, missingValueErrorSubstring1, missingValueErrorSubstring3), // combination 13: redudant due to combination 5
	// createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue200, nil}, missingValueErrorSubstring3), // combination 14: redudant due to combination 10
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue200, nil}, missingValueErrorSubstring3), // combination 15: redudant due to combination 10
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue200, nil}, missingValueErrorSubstring3, valueSequenceErrorSubstring), // combination 16: redudant due to combination 6

	// ---

	createInvalidValueTestComponent([]*float64{nil, nil, &thresholdValue0}, missingValueErrorSubstring1),              // combination 17
	createInvalidValueTestComponent([]*float64{&thresholdValue0, nil, &thresholdValue0}, valueSequenceErrorSubstring), // combination 18
	createValidValueTestComponent([]*float64{&thresholdValue100, nil, &thresholdValue0}),                              // combination 19
	// createValidValueTestComponent([]*float64{&thresholdValue200, nil, &thresholdValue0}), // combination 20: redudant due to combination 19

	createInvalidValueTestComponent([]*float64{nil, &thresholdValue0, &thresholdValue0}, missingValueErrorSubstring1, valueSequenceErrorSubstring), // combination 21
	createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue0, &thresholdValue0}, valueSequenceErrorSubstring),                 // combination 22
	createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue0, &thresholdValue0}, valueSequenceErrorSubstring),               // combination 23
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue0, &thresholdValue0}, valueSequenceErrorSubstring), // combination 24: redudant due to combination 23

	createInvalidValueTestComponent([]*float64{nil, &thresholdValue100, &thresholdValue0}, missingValueErrorSubstring1),                // combination 25
	createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue100, &thresholdValue0}, valueSequenceErrorSubstring),   // combination 26
	createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue100, &thresholdValue0}, valueSequenceErrorSubstring), // combination 27
	createValidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue100, &thresholdValue0}),                                // combination 28

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue200, &thresholdValue0}, missingValueErrorSubstring1), // combination 29: redudant due to combination 25
	// createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue200, &thresholdValue0}, valueSequenceErrorSubstring),  // combination 30: redudant due to combination 26
	createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue200, &thresholdValue0}, valueSequenceErrorSubstring), // combination 31
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue200, &thresholdValue0}, valueSequenceErrorSubstring), // combination 32: redudant due to combination 27

	//---

	// createInvalidValueTestComponent([]*float64{nil, nil, &thresholdValue100}, missingValueErrorSubstring1), // combination 33: redudant due to combination 17
	createValidValueTestComponent([]*float64{&thresholdValue0, nil, &thresholdValue100}), // combination 34
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, nil, &thresholdValue100}, valueSequenceErrorSubstring), // combination 35: redudant due to combination 18
	// createValidValueTestComponent([]*float64{&thresholdValue200, nil, &thresholdValue100}), // combination 36: redudant due to combination 19

	createInvalidValueTestComponent([]*float64{nil, &thresholdValue0, &thresholdValue100}, missingValueErrorSubstring1),                // combination 37
	createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue0, &thresholdValue100}, valueSequenceErrorSubstring),   // combination 38
	createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue0, &thresholdValue100}, valueSequenceErrorSubstring), // combination 39
	createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue0, &thresholdValue100}, valueSequenceErrorSubstring), // combination 40

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue100, &thresholdValue100}, missingValueErrorSubstring1, valueSequenceErrorSubstring), // combination 41: redudant due to combination 21
	createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue100, &thresholdValue100}, valueSequenceErrorSubstring), // combination 42
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue100, &thresholdValue100}, valueSequenceErrorSubstring), // combination 43: redudant due to combination 22
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue100, &thresholdValue100}, valueSequenceErrorSubstring), // combination 44: redudant due to combination 23

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue200, &thresholdValue100}, missingValueErrorSubstring1), // combination 45: redudant due to combination 25
	createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue200, &thresholdValue100}, valueSequenceErrorSubstring), // combination 46
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue200, &thresholdValue100}, valueSequenceErrorSubstring), // combination 47: redudant due to combination 26
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue200, &thresholdValue100}, valueSequenceErrorSubstring), // combination 48: redudant due to combination 27

	//---

	// createInvalidValueTestComponent([]*float64{nil, nil, &thresholdValue200}, missingValueErrorSubstring1), // combination 49: redudant due to combination 17
	// createValidValueTestComponent([]*float64{&thresholdValue0, nil, &thresholdValue200}), // combination 50: redudant due to combination  34
	// createValidValueTestComponent([]*float64{&thresholdValue100, nil, &thresholdValue200}), // combination 51: redudant due to combination 34
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, nil, &thresholdValue200}, valueSequenceErrorSubstring), // combination 52: redudant due to combination 18

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue0, &thresholdValue200}, missingValueErrorSubstring1), // combination 53: redudant due to combination 37
	// createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue0, &thresholdValue200}, missingValueErrorSubstring3, valueSequenceErrorSubstring), // combination 54: redudant due to combination 38
	createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue0, &thresholdValue200}, valueSequenceErrorSubstring), // combination 55
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue0, &thresholdValue200}, valueSequenceErrorSubstring), // combination 56: redudant due to combination 39

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue100, &thresholdValue200}, missingValueErrorSubstring1), // combination 57: redudant due to combination 37
	createValidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue100, &thresholdValue200}), // combination 58
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue100, &thresholdValue200}, valueSequenceErrorSubstring), // combination 59: redudant due to combination 38
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue100, &thresholdValue200}, valueSequenceErrorSubstring), // combination 60: redudant due to combination 39

	// createInvalidValueTestComponent([]*float64{nil, &thresholdValue200, &thresholdValue200}, missingValueErrorSubstring1, valueSequenceErrorSubstring), // combination 61: redudant due to combination 21
	// createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue200, &thresholdValue200}, missingValueErrorSubstring3, valueSequenceErrorSubstring), // combination 62: redudant due to combination 42
	// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue200, &thresholdValue200}, valueSequenceErrorSubstring), // combination 63: redudant due to combination 42
	// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue200, &thresholdValue200}, valueSequenceErrorSubstring), // combination 64: redudant due to combination 22
}

func createValidValueTestComponent(values []*float64) valueTestComponent {
	return valueTestComponent{
		values: values,
	}
}

func createInvalidValueTestComponent(values []*float64, expectedErrorSubstrings ...string) valueTestComponent {
	return valueTestComponent{
		values:                  values,
		expectError:             true,
		expectedErrorSubstrings: expectedErrorSubstrings,
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsing tests that parsing thresholds works or results in an error as expected.
// This test works by combining combinations of threshold colors with combinations of threshold values and checking the result.
func TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsing(t *testing.T) {
	for _, colorTestComponentVariant := range threeColorTestComponentVariants {
		for _, valueTestComponentVariant := range threeValueTestComponentVariants {
			t.Run(getThresholdParsingTestName(colorTestComponentVariant, valueTestComponentVariant), func(t *testing.T) {
				if colorTestComponentVariant.expectError || valueTestComponentVariant.expectError {
					runTileThresholdRuleParsingTestAndExpectError(t, colorTestComponentVariant, valueTestComponentVariant)
				} else {
					runTileThresholdRuleParsingTestAndExpectSuccess(t, colorTestComponentVariant, valueTestComponentVariant)
				}
			})
		}
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsingWithFourRules tests that parsing thresholds with an extra rule results in an error as expected.
// The Data explorer UI does not currently allow for this setup, however it is included for completeness in case this changes in the future.
func TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsingWithFourRules(t *testing.T) {
	for _, colorTestComponentVariant := range threeColorTestComponentVariants {
		for _, valueTestComponentVariant := range threeValueTestComponentVariants {
			colorTestComponentVariantWithExtraColor := addExtraInvalidColorToColorTestComponent(colorTestComponentVariant)
			valueTestComponentVariantWithExtraValue := addExtraNilValueToValueTestComponent(valueTestComponentVariant)
			t.Run(getThresholdParsingTestName(colorTestComponentVariantWithExtraColor, valueTestComponentVariantWithExtraValue), func(t *testing.T) {
				runTileThresholdRuleParsingTestAndExpectError(t, colorTestComponentVariantWithExtraColor, valueTestComponentVariantWithExtraValue)
			})
		}
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsingWithTwoRules tests that parsing thresholds with a missing rule results in an error as expected.
// The Data explorer UI does not currently allow for this setup, however it is included for completeness in case this changes in the future.
func TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsingWithTwoRules(t *testing.T) {

	// List all possible combinations of two theshold colors, where each one may be one of three valid colors (for pass, warn or fail) or an invalid color.
	var twoColorTestComponentVariants = []colorTestComponent{
		createInvalidColorTestComponent([]string{invalidThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{passThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{warnThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{failThresholdColor, invalidThresholdColor}, invalidColorErrorSubstring2),

		createInvalidColorTestComponent([]string{invalidThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{passThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{warnThresholdColor, passThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{failThresholdColor, passThresholdColor}, invalidColorErrorSubstring2),

		createInvalidColorTestComponent([]string{invalidThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1),
		createInvalidColorTestComponent([]string{passThresholdColor, warnThresholdColor}),
		createInvalidColorTestComponent([]string{warnThresholdColor, warnThresholdColor}, invalidColorErrorSubstring1),
		createInvalidColorTestComponent([]string{failThresholdColor, warnThresholdColor}),

		createInvalidColorTestComponent([]string{invalidThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{passThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{warnThresholdColor, failThresholdColor}, invalidColorErrorSubstring1, invalidColorErrorSubstring2),
		createInvalidColorTestComponent([]string{failThresholdColor, failThresholdColor}, invalidColorErrorSubstring2),
	}

	// List all possible combinations of two values, where each one may be one of three sample values as well as nil.
	// The list is sequential, i.e. ordered AA, AB. AC, AD, BA, BB, BC..., however Redundant cases have been commented out.
	var twoValueTestComponentVariants = []valueTestComponent{
		// would indicate no thresholds are to be used: //createInvalidValueTestComponent([]*float64{nil, nil}, incorrectThresholdRuleCountErrorSubstring), // combination 1
		createInvalidValueTestComponent([]*float64{&thresholdValue0, nil}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 2
		// createInvalidValueTestComponent([]*float64{&thresholdValue100, nil}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 3: redudant due to combination 2
		// createInvalidValueTestComponent([]*float64{&thresholdValue200, nil}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 4: redudant due to combination 2

		createInvalidValueTestComponent([]*float64{nil, &thresholdValue0}, missingValueErrorSubstring1, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring),              // combination 5
		createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue0}, missingValueErrorSubstring3, valueSequenceErrorSubstring, incorrectThresholdRuleCountErrorSubstring), // combination 6
		createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue0}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring),                            // combination 7
		// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue0}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 8: redudant due to combination 7

		// createInvalidValueTestComponent([]*float64{nil, &thresholdValue100}, missingValueErrorSubstring1, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 9: redudant due to combination 5
		createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue100}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 10
		// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue100}, missingValueErrorSubstring3, valueSequenceErrorSubstring, incorrectThresholdRuleCountErrorSubstring), // combination 11: redudant due to combination 6
		// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue100}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 12: redudant due to combination 7

		// createInvalidValueTestComponent([]*float64{nil, &thresholdValue200}, missingValueErrorSubstring1, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 13: redudant due to combination 5
		// createInvalidValueTestComponent([]*float64{&thresholdValue0, &thresholdValue200}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 14: redudant due to combination 10
		// createInvalidValueTestComponent([]*float64{&thresholdValue100, &thresholdValue200}, missingValueErrorSubstring3, incorrectThresholdRuleCountErrorSubstring), // combination 15: redudant due to combination 10
		// createInvalidValueTestComponent([]*float64{&thresholdValue200, &thresholdValue200}, missingValueErrorSubstring3, valueSequenceErrorSubstring, incorrectThresholdRuleCountErrorSubstring), // combination 16: redudant due to combination 6
	}

	for _, colorTestComponentVariant := range twoColorTestComponentVariants {
		for _, valueTestComponentVariant := range twoValueTestComponentVariants {
			t.Run(getThresholdParsingTestName(colorTestComponentVariant, valueTestComponentVariant), func(t *testing.T) {
				runTileThresholdRuleParsingTestAndExpectError(t, colorTestComponentVariant, valueTestComponentVariant)
			})
		}
	}
}

func runTileThresholdRuleParsingTestAndExpectSuccess(t *testing.T, colorTestComponentVariant colorTestComponent, valueTestComponentVariant valueTestComponent) {
	const testDataFolder = "./testdata/dashboards/data_explorer/tile_thresholds_parsing/success"

	handler := createHandlerWithTemplatedDashboard(t,
		filepath.Join(testDataFolder, "dashboard.template.json"),
		struct {
			ThresholdValues []*float64
			ThresholdColors []string
		}{
			ThresholdValues: valueTestComponentVariant.values,
			ThresholdColors: colorTestComponentVariant.colors,
		})

	requestBuilder := newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names")
	metricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testDataFolder, requestBuilder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("srt", 54896.50455400265, metricsRequest))
}

func runTileThresholdRuleParsingTestAndExpectError(t *testing.T, colorTestComponentVariant colorTestComponent, valueTestComponentVariant valueTestComponent) {
	const testDataFolder = "./testdata/dashboards/data_explorer/tile_thresholds_parsing/error"

	handler := createHandlerWithTemplatedDashboard(t,
		filepath.Join(testDataFolder, "dashboard.template.json"),
		struct {
			ThresholdValues []*float64
			ThresholdColors []string
		}{
			ThresholdValues: valueTestComponentVariant.values,
			ThresholdColors: colorTestComponentVariant.colors,
		})

	expectedErrorSubstrings := append(append([]string(nil), valueTestComponentVariant.expectedErrorSubstrings...), colorTestComponentVariant.expectedErrorSubstrings...)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("srt", expectedErrorSubstrings...))
}

func getThresholdParsingTestName(colorTestComponent colorTestComponent, valueTestComponent valueTestComponent) string {
	b := strings.Builder{}
	for i, color := range colorTestComponent.colors {
		if i != 0 {
			b.WriteString(" ")
		}

		var v *float64 = nil
		if i < 3 {
			v = valueTestComponent.values[i]
		}

		b.WriteString(dashboard.GetThresholdTypeNameForColor(color))
		b.WriteString(":")
		b.WriteString(formatPointerToFloat64AsString(v))
	}
	return b.String()
}

func formatPointerToFloat64AsString(v *float64) string {
	if v == nil {
		return "nil"
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func addExtraInvalidColorToColorTestComponent(v colorTestComponent) colorTestComponent {
	return colorTestComponent{
		colors:                  append(append([]string(nil), v.colors...), invalidThresholdColor),
		expectError:             true,
		expectedErrorSubstrings: append(append([]string(nil), v.expectedErrorSubstrings...), incorrectThresholdRuleCountErrorSubstring),
	}
}

func addExtraNilValueToValueTestComponent(v valueTestComponent) valueTestComponent {
	return valueTestComponent{
		values:                  append(append([]*float64(nil), v.values...), nil),
		expectError:             true,
		expectedErrorSubstrings: append(append([]string(nil), v.expectedErrorSubstrings...), incorrectThresholdRuleCountErrorSubstring),
	}
}

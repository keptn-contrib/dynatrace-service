package sli

import (
	"fmt"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

type dataExplorerThresholdErrorsTest struct {
	name                    string
	thresholdValues         []*float64
	thresholdColors         []string
	sliResultAssertionsFunc func(t *testing.T, actual sliResult)
}

type tileThresholdsTemplateData struct {
	ThresholdValues []*float64
	ThresholdColors []string
}

const (
	missingValueErrorSubstring                    = "missing value"
	invalidColorErrorSubstring                    = "invalid color"
	expected3RulesErrorSubstring                  = "expected 3 rules"
	invalidColorSequenceErrorSubstring            = "invalid color sequence"
	expectedMonotonicallyIncreasingErrorSubstring = "must increase strictly monotonically"
	atPosition1ErrorSubstring                     = "at position 1"
	atPosition2ErrorSubstring                     = "at position 2"
	atPosition3ErrorSubstring                     = "at position 3"
)

const (
	invalidThresholdColor = "#14a8f5"
	passThresholdColor    = "#7dc540"
	warnThresholdColor    = "#f5d30f"
	failThresholdColor    = "#dc172a"
)

var passWarnFailThresholdColors = []string{passThresholdColor, warnThresholdColor, failThresholdColor}

var thresholdValue0 float64 = 0
var thresholdValue68000 float64 = 68000
var thresholdValue69000 float64 = 69000

var validThresholdValues = []*float64{&thresholdValue0, &thresholdValue68000, &thresholdValue69000}

// TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsingErrors tests that errors while parsing Data Explorer tile thresholds are generated as expected.
// Includes tests with multiple errors are these should all be included in the overall error message.
func TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdRuleParsingErrors(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/tile_thresholds_errors/"

	tests := []struct {
		name                    string
		thresholdValues         []*float64
		thresholdColors         []string
		sliResultAssertionsFunc func(t *testing.T, actual sliResult)
	}{
		// Rule count
		{
			name:                    "Too few rules in thresholds",
			thresholdValues:         []*float64{&thresholdValue0, &thresholdValue69000},
			thresholdColors:         []string{passThresholdColor, warnThresholdColor},
			sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", expected3RulesErrorSubstring),
		},
		{
			name:                    "Too many rules in thresholds",
			thresholdValues:         []*float64{&thresholdValue0, &thresholdValue68000, &thresholdValue69000, &thresholdValue69000},
			thresholdColors:         []string{passThresholdColor, warnThresholdColor, failThresholdColor, failThresholdColor},
			sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", expected3RulesErrorSubstring),
		},

		// Missing values
		createDataExplorerThresholdsErrorTestWithMissingValues("Missing value at position 1", nil, &thresholdValue68000, &thresholdValue69000, atPosition1ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithMissingValues("Missing value at position 2", &thresholdValue0, nil, &thresholdValue69000, atPosition2ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithMissingValues("Missing value at position 3", &thresholdValue0, &thresholdValue68000, nil, atPosition3ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithMissingValues("Missing values at position 1 and 2", nil, nil, &thresholdValue69000, atPosition1ErrorSubstring, atPosition2ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithMissingValues("Missing values at position 2 and 3", &thresholdValue0, nil, nil, atPosition2ErrorSubstring, atPosition3ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithMissingValues("Missing values at position 1 and 3", nil, &thresholdValue68000, nil, atPosition1ErrorSubstring, atPosition3ErrorSubstring),

		// Combined rule count and missing value
		{
			name:                    "Too many rules in thresholds and missing value",
			thresholdValues:         []*float64{&thresholdValue0, &thresholdValue68000, nil, &thresholdValue69000},
			thresholdColors:         []string{passThresholdColor, warnThresholdColor, failThresholdColor, failThresholdColor},
			sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", expected3RulesErrorSubstring, missingValueErrorSubstring, atPosition3ErrorSubstring),
		},

		// Invalid color
		createDataExplorerThresholdsErrorTestWithInvalidColors("Invalid color at position 1", invalidThresholdColor, warnThresholdColor, failThresholdColor, atPosition1ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithInvalidColors("Invalid color at position 2", passThresholdColor, invalidThresholdColor, failThresholdColor, atPosition2ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithInvalidColors("Invalid color at position 3", passThresholdColor, warnThresholdColor, invalidThresholdColor, atPosition3ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithInvalidColors("Invalid color at position 1 and 2", invalidThresholdColor, invalidThresholdColor, failThresholdColor, atPosition1ErrorSubstring, atPosition2ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithInvalidColors("Invalid color at position 2 and 3", passThresholdColor, invalidThresholdColor, invalidThresholdColor, atPosition2ErrorSubstring, atPosition3ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithInvalidColors("Invalid color at position 1 and 3", invalidThresholdColor, warnThresholdColor, invalidThresholdColor, atPosition1ErrorSubstring, atPosition3ErrorSubstring),
		createDataExplorerThresholdsErrorTestWithInvalidColors("Invalid color at position 1, 2 and 3", invalidThresholdColor, invalidThresholdColor, invalidThresholdColor, atPosition1ErrorSubstring, atPosition2ErrorSubstring, atPosition3ErrorSubstring),

		// Combined invalid color and missing value
		{
			name:                    "Invalid color at position 1 and missing value at position 2",
			thresholdValues:         []*float64{&thresholdValue0, nil, &thresholdValue69000},
			thresholdColors:         []string{invalidThresholdColor, warnThresholdColor, failThresholdColor},
			sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", invalidColorErrorSubstring, atPosition1ErrorSubstring, missingValueErrorSubstring, atPosition2ErrorSubstring),
		},

		// Combined invalid color, missing value and too many rules
		{
			name:                    "Invalid color at position 1 and missing value at position 2 and too many rules",
			thresholdValues:         []*float64{&thresholdValue0, nil, &thresholdValue69000, &thresholdValue69000},
			thresholdColors:         []string{invalidThresholdColor, warnThresholdColor, failThresholdColor, failThresholdColor},
			sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", invalidColorErrorSubstring, atPosition1ErrorSubstring, missingValueErrorSubstring, atPosition2ErrorSubstring, expected3RulesErrorSubstring),
		},

		// Invalid color sequences
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-pass-pass sequence", passThresholdColor, passThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-pass-pass sequence", warnThresholdColor, passThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-pass-pass sequence", failThresholdColor, passThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-warn-pass sequence", passThresholdColor, warnThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-warn-pass sequence", warnThresholdColor, warnThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-fail-pass sequence", passThresholdColor, failThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-fail-pass sequence", warnThresholdColor, failThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-fail-pass sequence", failThresholdColor, failThresholdColor, passThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-pass-warn sequence", passThresholdColor, passThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-pass-warn sequence", warnThresholdColor, passThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-pass-warn sequence", failThresholdColor, passThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-warn-warn sequence", passThresholdColor, warnThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-warn-warn sequence", warnThresholdColor, warnThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-warn-warn sequence", failThresholdColor, warnThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-fail-warn sequence", passThresholdColor, failThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-fail-warn sequence", warnThresholdColor, failThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-fail-warn sequence", failThresholdColor, failThresholdColor, warnThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-pass-fail sequence", passThresholdColor, passThresholdColor, failThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-pass-fail sequence", warnThresholdColor, passThresholdColor, failThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-pass-fail sequence", failThresholdColor, passThresholdColor, failThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-warn-fail sequence", warnThresholdColor, warnThresholdColor, failThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-warn-fail sequence", failThresholdColor, warnThresholdColor, failThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid pass-fail-fail sequence", passThresholdColor, failThresholdColor, failThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid warn-fail-fail sequence", warnThresholdColor, failThresholdColor, failThresholdColor),
		createDataExplorerThresholdsErrorTestWithColorSequence("Invalid fail-fail-fail sequence", failThresholdColor, failThresholdColor, failThresholdColor),

		// Not strictly monotonically increasing values
		// Redundant cases have been removed
		createDataExplorerThresholdsErrorTestWithWrongValues(0, 0, 0),         // *-same-same
		createDataExplorerThresholdsErrorTestWithWrongValues(0, 0, 68000),     // *-same-higher
		createDataExplorerThresholdsErrorTestWithWrongValues(0, 68000, 68000), // *-higher-higher
		createDataExplorerThresholdsErrorTestWithWrongValues(0, 68000, 0),     // *-higher-same
		createDataExplorerThresholdsErrorTestWithWrongValues(0, 69000, 68000), // *-even_higher-higher
		createDataExplorerThresholdsErrorTestWithWrongValues(68000, 0, 0),     // *-lower-lower
		createDataExplorerThresholdsErrorTestWithWrongValues(68000, 0, 68000), // *-lower-same
		createDataExplorerThresholdsErrorTestWithWrongValues(68000, 0, 69000), // *-lower-higher
		createDataExplorerThresholdsErrorTestWithWrongValues(68000, 68000, 0), // *-same-lower
		createDataExplorerThresholdsErrorTestWithWrongValues(68000, 69000, 0), // *-higher-lower
		createDataExplorerThresholdsErrorTestWithWrongValues(69000, 0, 68000), // *-even_lower-lower
		createDataExplorerThresholdsErrorTestWithWrongValues(69000, 68000, 0), // *-lower-even_lower

		// Combined invalid color sequence and not strictly monotonically increasing values
		{
			name:                    "Invalid color sequence and not strictly monotonically increasing values",
			thresholdValues:         []*float64{&thresholdValue68000, &thresholdValue69000, &thresholdValue69000},
			thresholdColors:         []string{warnThresholdColor, warnThresholdColor, failThresholdColor},
			sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", invalidColorSequenceErrorSubstring, expectedMonotonicallyIncreasingErrorSubstring),
		},
	}

	for _, thresholdTest := range tests {
		t.Run(thresholdTest.name, func(t *testing.T) {
			handler := test.NewTemplatingPayloadBasedURLHandler(t)
			handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID,
				testDataFolder+"dashboard_thresholds_template.json",
				tileThresholdsTemplateData{ThresholdValues: thresholdTest.thresholdValues, ThresholdColors: thresholdTest.thresholdColors})

			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, thresholdTest.sliResultAssertionsFunc)
		})
	}
}

func createDataExplorerThresholdsErrorTestWithMissingValues(name string, v1 *float64, v2 *float64, v3 *float64, positionErrorSubstrings ...string) dataExplorerThresholdErrorsTest {
	return createDataExplorerThresholdsErrorTestWithValues(name, v1, v2, v3, missingValueErrorSubstring, positionErrorSubstrings...)
}

func createDataExplorerThresholdsErrorTestWithWrongValues(v1 float64, v2 float64, v3 float64) dataExplorerThresholdErrorsTest {
	name := fmt.Sprintf("Invalid not strictly monotonically increasing values test %f %f %f", v1, v2, v3)
	return createDataExplorerThresholdsErrorTestWithValues(name, &v1, &v2, &v3, expectedMonotonicallyIncreasingErrorSubstring)
}

func createDataExplorerThresholdsErrorTestWithValues(name string, v1 *float64, v2 *float64, v3 *float64, mainErrorSubstring string, positionErrorSubstrings ...string) dataExplorerThresholdErrorsTest {
	expectedErrorSubstrings := append([]string{mainErrorSubstring}, positionErrorSubstrings...)
	return dataExplorerThresholdErrorsTest{
		name:                    name,
		thresholdValues:         []*float64{v1, v2, v3},
		thresholdColors:         passWarnFailThresholdColors,
		sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", expectedErrorSubstrings...),
	}
}

func createDataExplorerThresholdsErrorTestWithInvalidColors(name string, c1 string, c2 string, c3 string, positionErrorSubstrings ...string) dataExplorerThresholdErrorsTest {
	return createDataExplorerThresholdsErrorTestWithColors(name, c1, c2, c3, invalidColorErrorSubstring, positionErrorSubstrings...)
}

func createDataExplorerThresholdsErrorTestWithColorSequence(name string, c1 string, c2 string, c3 string) dataExplorerThresholdErrorsTest {
	return createDataExplorerThresholdsErrorTestWithColors(name, c1, c2, c3, invalidColorSequenceErrorSubstring)
}

func createDataExplorerThresholdsErrorTestWithColors(name string, c1 string, c2 string, c3 string, mainErrorSubstring string, positionErrorSubstrings ...string) dataExplorerThresholdErrorsTest {
	expectedErrorSubstrings := append([]string{mainErrorSubstring}, positionErrorSubstrings...)
	return dataExplorerThresholdErrorsTest{
		name:                    name,
		thresholdValues:         validThresholdValues,
		thresholdColors:         []string{c1, c2, c3},
		sliResultAssertionsFunc: createFailedSLIResultAssertionsFunc("srt", expectedErrorSubstrings...),
	}
}

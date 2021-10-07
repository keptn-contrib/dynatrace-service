package unit

import (
	"regexp"
	"strings"
)

var microSecondPattern = regexp.MustCompile(`^[Mm]icro[Ss]econd$`)
var bytePattern = regexp.MustCompile(`^[Bb]yte$`)

// ScaleData
// scales data based on the timeseries identifier (e.g., service.responsetime needs to be scaled from microseconds to milliseocnds)
// Right now this method scales microseconds to milliseconds and bytes to Kilobytes
// At a later stage we should extend this with more conversions and even think of allowing custom scale targets, e.g: Byte to MegaByte
func ScaleData(metricID string, unit string, value float64) float64 {
	if microSecondPattern.MatchString(unit) || strings.Contains(metricID, "builtin:service.response.time") {
		// scale from microseconds to milliseconds
		return value / 1000.0
	}

	if bytePattern.MatchString(unit) {
		// convert Bytes to Kilobyte
		return value / 1024
	}

	return value
}

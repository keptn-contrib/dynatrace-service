package unit

import (
	"regexp"
)

var microSecondPattern = regexp.MustCompile(`^[Mm]icro[Ss]econd$`)
var bytePattern = regexp.MustCompile(`^[Bb]yte$`)

// ScaleData scales data based on the unit, i.e converts microseconds to milliseconds and bytes to Kilobytes.
func ScaleData(unit string, value float64) float64 {
	if isMicroSecondUnit(unit) {
		// scale from microseconds to milliseconds
		return value / 1000.0
	}

	if isByteUnit(unit) {
		// convert Bytes to Kilobyte
		return value / 1024
	}

	return value
}

func canBeConverted(unit string) bool {
	return isByteUnit(unit) || isMicroSecondUnit(unit)
}

func isByteUnit(unit string) bool {
	return bytePattern.MatchString(unit)
}

func isMicroSecondUnit(unit string) bool {
	return microSecondPattern.MatchString(unit)
}

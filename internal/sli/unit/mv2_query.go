package unit

import (
	"fmt"
	"regexp"
)

var mv2Pattern = regexp.MustCompile(`^MV2;(([bB]yte)|([Mm]icro[sS]econd));(.+)$`)

// ParseMV2Query tries to parse a MV2 query into a metrics query and a unit if it allows, otherwise it returns an error
func ParseMV2Query(mv2Query string) (metricsQuery string, metricUnit string, err error) {
	chunks := mv2Pattern.FindStringSubmatch(mv2Query)
	if len(chunks) != 5 {
		return "", "", createMV2FormatError(mv2Query)
	}

	metricUnit = chunks[1]
	if metricUnit == "" {
		return "", "", createMV2FormatError(mv2Query)
	}

	metricsQuery = chunks[4]
	if metricsQuery == "" {
		return "", "", createMV2FormatError(mv2Query)
	}

	return metricsQuery, metricUnit, nil
}

func createMV2FormatError(query string) error {
	return fmt.Errorf("could not parse SLI definition format - should either be 'MV2;Byte;<query>' or 'MV2;MicroSecond;<query>': %s", query)
}

// ConvertToMV2Query expresses the specified metrics query as a MV2 query if the unit allows, otherwise returns an error
func ConvertToMV2Query(metricsQuery string, metricUnit string) (string, error) {
	if canBeConverted(metricUnit) {
		return fmt.Sprintf("MV2;%s;%s", metricUnit, metricsQuery), nil
	}

	return "", fmt.Errorf("could not convert to MV2 query format - unexpected unit: %s", metricUnit)
}
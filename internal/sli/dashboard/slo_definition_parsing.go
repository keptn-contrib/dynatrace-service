package dashboard

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
)

const (
	sloDefSli     = "sli"
	sloDefPass    = "pass"
	sloDefWarning = "warning"
	sloDefKey     = "key"
	sloDefWeight  = "weight"
)

// parseSLODefinition takes a value such as
//   Example 1: Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true
//   Example 2: Response time (P95);sli=svc_rt_p95;pass=<+10%,<600
//   Example 3: Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false
// can also take a value like
// 	 "KQG;project=myproject;pass=90%;warning=75%;"
// This will return a SLO object or an error if parsing was not possible
func parseSLODefinition(sloDefinition string) (*keptncommon.SLO, error) {
	result := &keptncommon.SLO{
		Weight: 1,
		KeySLI: false,
	}
	var errs []error

	nameValueSplits := strings.Split(sloDefinition, ";")

	// let's iterate through all name-value pairs which are separated through ";" to extract keys such as warning, pass, weight, key, sli
	keyFound := make(map[string]bool)
	for i := 0; i < len(nameValueSplits); i++ {

		nameValueDividerIndex := strings.Index(nameValueSplits[i], "=")
		if nameValueDividerIndex < 0 {
			continue
		}

		// for each name=value pair we get the name as first part of the string until the first =
		// the value is the after that =
		nameString := strings.ToLower(nameValueSplits[i][:nameValueDividerIndex])
		valueString := strings.TrimSpace(nameValueSplits[i][nameValueDividerIndex+1:])
		var err error
		switch nameString {
		case sloDefSli:
			if keyFound[sloDefSli] {
				errs = append(errs, &duplicateKeyError{key: sloDefSli})
				break
			}
			result.SLI = valueString
			if valueString == "" {
				errs = append(errs, fmt.Errorf("sli name is empty"))
			}
			keyFound[sloDefSli] = true
		case sloDefPass:
			passCriteria, err := parseSLOCriteriaString(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", sloDefPass, err))
				break
			}
			result.Pass = append(result.Pass, passCriteria)
		case sloDefWarning:
			warningCriteria, err := parseSLOCriteriaString(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", sloDefWarning, err))
				break
			}
			result.Warning = append(result.Warning, warningCriteria)
		case sloDefKey:
			if keyFound[sloDefKey] {
				errs = append(errs, &duplicateKeyError{key: sloDefKey})
				break
			}
			result.KeySLI, err = strconv.ParseBool(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not a boolean value: %v", sloDefKey, valueString))
			}
			keyFound[sloDefKey] = true
		case sloDefWeight:
			if keyFound[sloDefWeight] {
				errs = append(errs, &duplicateKeyError{key: sloDefWeight})
				break
			}
			result.Weight, err = strconv.Atoi(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not an integer value: %v", sloDefWeight, valueString))
			}
			keyFound[sloDefWeight] = true
		}
	}

	if len(errs) > 0 {
		return nil, &sloDefinitionError{
			sliName:   result.SLI,
			tileTitle: sloDefinition,
			errors:    errs,
		}
	}

	return result, nil
}

func parseSLOCriteriaString(criteria string) (*keptncommon.SLOCriteria, error) {
	criteriaChunks := strings.Split(criteria, ",")
	var invalidCriteria []string
	for _, criterion := range criteriaChunks {
		if criterionIsNotValid(criterion) {
			invalidCriteria = append(invalidCriteria, criterion)
		}
	}

	if len(invalidCriteria) > 0 {
		return nil, fmt.Errorf("invalid criteria value(s): %s", strings.Join(invalidCriteria, ","))
	}

	return &keptncommon.SLOCriteria{Criteria: criteriaChunks}, nil
}

func criterionIsNotValid(criterion string) bool {
	pattern := regexp.MustCompile("^(<|<=|=|>|>=)([+-]?\\d+|[+-]?\\d+\\.\\d+)([%]?)$")

	return !pattern.MatchString(criterion)
}

// sloDefinitionError represents an error that occurred while parsing an SLO definition
type sloDefinitionError struct {
	tileTitle string
	sliName   string
	errors    []error
}

func (err *sloDefinitionError) Error() string {
	var errStrings = make([]string, len(err.errors))
	for i, e := range err.errors {
		errStrings[i] = e.Error()
	}
	return strings.Join(errStrings, ";")
}

// sliNameOrTileTitle returns the SLI name or the tile title, if the SLI name is empty
func (err *sloDefinitionError) sliNameOrTileTitle() string {
	if err.sliName != "" {
		return err.sliName
	}

	return err.tileTitle
}

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
	sloDefExclude = "exclude"
)

type sloDefinitionParsingResult struct {
	sloDefinition keptncommon.SLO
	exclude       bool
}

// parseSLODefinition takes a value such as
//   Example 1: Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true
//   Example 2: Response time (P95);sli=svc_rt_p95;pass=<+10%,<600
//   Example 3: Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false
// can also take a value like
// 	 "KQG;project=myproject;pass=90%;warning=75%;"
// This will return a SLO object or an error if parsing was not possible
func parseSLODefinition(sloDefinition string) (*sloDefinitionParsingResult, error) {
	result := &sloDefinitionParsingResult{
		sloDefinition: keptncommon.SLO{
			Weight: 1,
			KeySLI: false,
		},
		exclude: false,
	}
	var errs []error

	keyFound := make(map[string]bool)
	for i, kv := range newKeyValueParsing(sloDefinition).parse() {

		if !kv.split {
			if i == 0 {
				result.sloDefinition.DisplayName = kv.key
			}
			continue
		}

		var err error
		switch strings.ToLower(kv.key) {
		case sloDefSli:
			if keyFound[sloDefSli] {
				errs = append(errs, &duplicateKeyError{key: sloDefSli})
				break
			}
			keyFound[sloDefSli] = true

			if kv.value == "" {
				errs = append(errs, fmt.Errorf("sli name is empty"))
				break
			}

			if result.sloDefinition.DisplayName == "" {
				result.sloDefinition.DisplayName = kv.value
			}
			result.sloDefinition.SLI = cleanIndicatorName(kv.value)

		case sloDefPass:
			passCriteria, err := parseSLOCriteriaString(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", sloDefPass, err))
				break
			}
			result.sloDefinition.Pass = append(result.sloDefinition.Pass, passCriteria)

		case sloDefWarning:
			warningCriteria, err := parseSLOCriteriaString(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", sloDefWarning, err))
				break
			}
			result.sloDefinition.Warning = append(result.sloDefinition.Warning, warningCriteria)

		case sloDefKey:
			if keyFound[sloDefKey] {
				errs = append(errs, &duplicateKeyError{key: sloDefKey})
				break
			}
			keyFound[sloDefKey] = true

			result.sloDefinition.KeySLI, err = strconv.ParseBool(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not a boolean value: %v", sloDefKey, kv.value))
			}

		case sloDefWeight:
			if keyFound[sloDefWeight] {
				errs = append(errs, &duplicateKeyError{key: sloDefWeight})
				break
			}
			keyFound[sloDefWeight] = true

			result.sloDefinition.Weight, err = strconv.Atoi(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not an integer value: %v", sloDefWeight, kv.value))
			}

		case sloDefExclude:
			if keyFound[sloDefExclude] {
				errs = append(errs, &duplicateKeyError{key: sloDefExclude})
				break
			}
			keyFound[sloDefExclude] = true

			result.exclude, err = strconv.ParseBool(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not a boolean value: %v", sloDefExclude, kv.value))
			}
		}
	}

	if result.sloDefinition.SLI == "" && result.sloDefinition.DisplayName != "" {
		result.sloDefinition.SLI = cleanIndicatorName(result.sloDefinition.DisplayName)
	}

	if len(errs) > 0 {
		return nil, &sloDefinitionError{
			sliName:   result.sloDefinition.SLI,
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

// cleanIndicatorName makes sure we have a valid indicator name by getting rid of special characters.
// All spaces, periods, forward-slashs, and percent and dollar signs are replaced with an underscore.
func cleanIndicatorName(indicatorName string) string {
	indicatorName = strings.ReplaceAll(indicatorName, " ", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "/", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "%", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "$", "_")
	indicatorName = strings.ReplaceAll(indicatorName, ".", "_")
	return indicatorName
}

package dashboard

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/ff"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"regexp"
	"strconv"
	"strings"
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
	sloDefinition result.SLO
	exclude       bool
}

// parseSLODefinition takes a value such as
//
//	Example 1: Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true
//	Example 2: Response time (P95);sli=svc_rt_p95;pass=<+10%,<600
//	Example 3: Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false
//
// can also take a value like
//
//	"KQG;project=myproject;pass=90%;warning=75%;"
//
// This will return a SLO object or an error if parsing was not possible
func parseSLODefinition(flags ff.GetSLIFeatureFlags, sloDefinition string) (sloDefinitionParsingResult, error) {
	res := sloDefinitionParsingResult{
		sloDefinition: result.SLO{
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
				res.sloDefinition.DisplayName = kv.key
			}
			continue
		}

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

			if res.sloDefinition.DisplayName == "" {
				res.sloDefinition.DisplayName = kv.value
			}
			res.sloDefinition.SLI = cleanIndicatorName(flags.SkipLowercaseSLINames(), kv.value)

		case sloDefPass:
			passCriteria, err := parseSLOCriteriaString(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", sloDefPass, err))
				break
			}
			res.sloDefinition.Pass = append(res.sloDefinition.Pass, passCriteria)

		case sloDefWarning:
			warningCriteria, err := parseSLOCriteriaString(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", sloDefWarning, err))
				break
			}
			res.sloDefinition.Warning = append(res.sloDefinition.Warning, warningCriteria)

		case sloDefKey:
			if keyFound[sloDefKey] {
				errs = append(errs, &duplicateKeyError{key: sloDefKey})
				break
			}
			keyFound[sloDefKey] = true

			val, err := strconv.ParseBool(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not a boolean value: %v", sloDefKey, kv.value))
				break
			}
			res.sloDefinition.KeySLI = val

		case sloDefWeight:
			if keyFound[sloDefWeight] {
				errs = append(errs, &duplicateKeyError{key: sloDefWeight})
				break
			}
			keyFound[sloDefWeight] = true

			val, err := strconv.Atoi(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not an integer value: %v", sloDefWeight, kv.value))
				break
			}
			res.sloDefinition.Weight = val

		case sloDefExclude:
			if keyFound[sloDefExclude] {
				errs = append(errs, &duplicateKeyError{key: sloDefExclude})
				break
			}
			keyFound[sloDefExclude] = true

			val, err := strconv.ParseBool(kv.value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not a boolean value: %v", sloDefExclude, kv.value))
				break
			}
			res.exclude = val
		}
	}

	if res.sloDefinition.SLI == "" && res.sloDefinition.DisplayName != "" {
		// do not skip lowercase operation here, as SLI was not set - so it cannot be legacy behavior
		res.sloDefinition.SLI = cleanIndicatorName(false, res.sloDefinition.DisplayName)
	}

	if len(errs) > 0 {

		return res, &sloDefinitionError{
			errors: errs,
		}
	}

	return res, nil
}

func parseSLOCriteriaString(criteria string) (*result.SLOCriteria, error) {
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

	return &result.SLOCriteria{Criteria: criteriaChunks}, nil
}

func criterionIsNotValid(criterion string) bool {
	pattern := regexp.MustCompile("^(<|<=|=|>|>=)([+-]?\\d+|[+-]?\\d+\\.\\d+)([%]?)$")

	return !pattern.MatchString(criterion)
}

// sloDefinitionError represents an error that occurred while parsing an SLO definition
type sloDefinitionError struct {
	errors []error
}

func (err *sloDefinitionError) Error() string {
	var errStrings = make([]string, len(err.errors))
	for i, e := range err.errors {
		errStrings[i] = e.Error()
	}
	return fmt.Sprintf("error parsing SLO definition: %s", strings.Join(errStrings, "; "))
}

// cleanIndicatorName makes sure we have a valid indicator name by forcing lower case and getting rid of special characters.
// All spaces, periods, forward-slashes, and percent and dollar signs are replaced with an underscore.
func cleanIndicatorName(skipLowercaseSLINames bool, indicatorName string) string {
	if !skipLowercaseSLINames {
		indicatorName = strings.ToLower(indicatorName)
	}

	indicatorName = strings.ReplaceAll(indicatorName, " ", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "/", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "%", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "$", "_")
	indicatorName = strings.ReplaceAll(indicatorName, ".", "_")
	return indicatorName
}

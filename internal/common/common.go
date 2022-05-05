package common

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
)

const ProblemURLLabel = "Problem URL"

// DynatraceConfigDashboardQUERY defines the Dynatrace Configuration File structure and supporting Constants
const DynatraceConfigDashboardQUERY = "query"

// ReplaceQueryParameters replaces query parameters based on sli filters and keptn event data
func ReplaceQueryParameters(query string, customFilters []*keptnv2.SLIFilter, keptnEvent adapter.EventContentAdapter) string {
	// apply custom filters
	for _, filter := range customFilters {
		filter.Value = strings.Replace(filter.Value, "'", "", -1)
		filter.Value = strings.Replace(filter.Value, "\"", "", -1)

		// replace the key in both variants, "normal" and uppercased
		query = strings.Replace(query, "$"+filter.Key, filter.Value, -1)
		query = strings.Replace(query, "$"+strings.ToUpper(filter.Key), filter.Value, -1)
	}

	query = ReplaceKeptnPlaceholders(query, keptnEvent)

	return query
}

// ReplaceKeptnPlaceholders will replaces $ placeholders with actual values
// $CONTEXT, $EVENT, $SOURCE
// $PROJECT, $STAGE, $SERVICE, $DEPLOYMENT
// $TESTSTRATEGY
// $LABEL.XXXX  -> will replace that with a label called XXXX
// $ENV.XXXX    -> will replace that with an env variable called XXXX
func ReplaceKeptnPlaceholders(input string, keptnEvent adapter.EventContentAdapter) string {
	result := input
	// first we do the regular keptn values
	result = strings.Replace(result, "$CONTEXT", keptnEvent.GetShKeptnContext(), -1)
	result = strings.Replace(result, "$EVENT", keptnEvent.GetEvent(), -1)
	result = strings.Replace(result, "$SOURCE", keptnEvent.GetSource(), -1)
	result = strings.Replace(result, "$PROJECT", keptnEvent.GetProject(), -1)
	result = strings.Replace(result, "$STAGE", keptnEvent.GetStage(), -1)
	result = strings.Replace(result, "$SERVICE", keptnEvent.GetService(), -1)
	result = strings.Replace(result, "$DEPLOYMENT", keptnEvent.GetDeployment(), -1)
	result = strings.Replace(result, "$TESTSTRATEGY", keptnEvent.GetTestStrategy(), -1)

	// now we do the labels
	for key, value := range keptnEvent.GetLabels() {
		result = strings.Replace(result, "$LABEL."+key, value, -1)
	}

	// now we do all environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		result = strings.Replace(result, "$ENV."+pair[0], pair[1], -1)
	}

	// TODO: 2021-12-21: Consider adding support for $SECRET.YYYY would be replaced with the k8s secret called YYYY

	return result
}

// TimestampToUnixMillisecondsString converts timestamp into a Unix milliseconds string.
func TimestampToUnixMillisecondsString(time time.Time) string {
	return strconv.FormatInt(time.Unix()*1000, 10)
}

type sloDefinitionError struct {
	errors []error
}

func (err *sloDefinitionError) Error() string {
	var errStrings = make([]string, len(err.errors))
	for i, e := range err.errors {
		errStrings[i] = e.Error()
	}
	return strings.Join(errStrings, ";")
}

type duplicateKeyError struct {
	key string
}

func (err *duplicateKeyError) Error() string {
	return fmt.Sprintf("duplicate key '%s' in SLO definition", err.key)
}

const (
	SloDefSli     = "sli"
	SloDefPass    = "pass"
	SloDefWarning = "warning"
	SloDefKey     = "key"
	SloDefWeight  = "weight"
)

// ParseSLOFromString takes a value such as
//   Example 1: Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true
//   Example 2: Response time (P95);sli=svc_rt_p95;pass=<+10%,<600
//   Example 3: Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false
// can also take a value like
// 	 "KQG;project=myproject;pass=90%;warning=75%;"
// This will return a SLO object or an error if parsing was not possible
func ParseSLOFromString(customName string) (*keptncommon.SLO, error) {
	result := &keptncommon.SLO{
		Weight:  1,
		KeySLI:  false,
		Pass:    []*keptncommon.SLOCriteria{},
		Warning: []*keptncommon.SLOCriteria{},
	}
	var errs []error

	nameValueSplits := strings.Split(customName, ";")

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
		case SloDefSli:
			if keyFound[SloDefSli] {
				errs = append(errs, &duplicateKeyError{key: SloDefSli})
				break
			}
			result.SLI = valueString
			if valueString == "" {
				errs = append(errs, fmt.Errorf("sli name is empty"))
			}
			keyFound[SloDefSli] = true
		case SloDefPass:
			passCriteria, err := parseSLOCriteriaString(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", SloDefPass, err))
				break
			}
			result.Pass = append(result.Pass, passCriteria)
		case SloDefWarning:
			warningCriteria, err := parseSLOCriteriaString(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': %w", SloDefWarning, err))
				break
			}
			result.Warning = append(result.Warning, warningCriteria)
		case SloDefKey:
			if keyFound[SloDefKey] {
				errs = append(errs, &duplicateKeyError{key: SloDefKey})
				break
			}
			result.KeySLI, err = strconv.ParseBool(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not a boolean value: %v", SloDefKey, valueString))
			}
			keyFound[SloDefKey] = true
		case SloDefWeight:
			if keyFound[SloDefWeight] {
				errs = append(errs, &duplicateKeyError{key: SloDefWeight})
				break
			}
			result.Weight, err = strconv.Atoi(valueString)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid definition for '%s': not an integer value: %v", SloDefWeight, valueString))
			}
			keyFound[SloDefWeight] = true
		}
	}

	// if we have no criteria for warn or pass we just return nil
	// not having a value for 'pass' means: this SLI is for informational purposes only and will not be evaluated.
	if len(result.Pass) == 0 {
		result.Pass = nil
	}
	if len(result.Warning) == 0 {
		result.Warning = nil
	}

	if len(errs) > 0 {
		return nil, &sloDefinitionError{errors: errs}
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

// CleanIndicatorName makes sure we have a valid indicator name by getting rid of special characters
func CleanIndicatorName(indicatorName string) string {
	indicatorName = strings.ReplaceAll(indicatorName, " ", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "/", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "%", "_")

	return indicatorName
}

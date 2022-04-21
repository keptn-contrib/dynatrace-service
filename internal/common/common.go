package common

import (
	"os"
	"strconv"
	"strings"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"

	log "github.com/sirupsen/logrus"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
)

const BridgeLabel = "Keptns Bridge"
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

func ParsePassAndWarningWithoutDefaultsFrom(customName string) *keptncommon.SLO {
	return ParsePassAndWarningFromString(customName, []string{}, []string{})
}

// ParsePassAndWarningFromString takes a value such as
//   Example 1: Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true
//   Example 2: Response time (P95);sli=svc_rt_p95;pass=<+10%,<600
//   Example 3: Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false
// can also take a value like
// 	 "KQG;project=myproject;pass=90%;warning=75%;"
// This will return a SLO object
func ParsePassAndWarningFromString(customName string, defaultPass []string, defaultWarning []string) *keptncommon.SLO {
	result := &keptncommon.SLO{
		Weight:  1,
		KeySLI:  false,
		Pass:    []*keptncommon.SLOCriteria{},
		Warning: []*keptncommon.SLOCriteria{},
	}

	nameValueSplits := strings.Split(customName, ";")

	// lets iterate through all name-value pairs which are separated through ";" to extract keys such as warning, pass, weight, key, sli
	for i := 0; i < len(nameValueSplits); i++ {

		nameValueDividerIndex := strings.Index(nameValueSplits[i], "=")
		if nameValueDividerIndex < 0 {
			continue
		}

		// for each name=value pair we get the name as first part of the string until the first =
		// the value is the after that =
		nameString := strings.ToLower(nameValueSplits[i][:nameValueDividerIndex])
		valueString := nameValueSplits[i][nameValueDividerIndex+1:]
		var err error
		switch nameString /*nameValueSplit[0]*/ {
		case "sli":
			result.SLI = valueString
		case "pass":
			result.Pass = append(
				result.Pass,
				&keptncommon.SLOCriteria{Criteria: strings.Split(valueString, ",")})
		case "warning":
			result.Warning = append(
				result.Warning,
				&keptncommon.SLOCriteria{Criteria: strings.Split(valueString, ",")})
		case "key":
			result.KeySLI, err = strconv.ParseBool(valueString)
			if err != nil {
				log.WithError(err).Warn("Error parsing bool")
			}
		case "weight":
			result.Weight, err = strconv.Atoi(valueString)
			if err != nil {
				log.WithError(err).Warn("Error parsing weight")
			}
		}
	}

	// use the defaults if nothing was specified
	if (len(result.Pass) == 0) && (len(defaultPass) > 0) {
		result.Pass = append(result.Pass, &keptncommon.SLOCriteria{Criteria: defaultPass})
	}

	if (len(result.Warning) == 0) && (len(defaultWarning) > 0) {
		result.Warning = append(result.Warning, &keptncommon.SLOCriteria{Criteria: defaultWarning})
	}

	// if we have no criteria for warn or pass we just return nil
	if len(result.Pass) == 0 {
		result.Pass = nil
	}
	if len(result.Warning) == 0 {
		result.Warning = nil
	}

	return result
}

// CleanIndicatorName makes sure we have a valid indicator name by getting rid of special characters
func CleanIndicatorName(indicatorName string) string {
	indicatorName = strings.ReplaceAll(indicatorName, " ", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "/", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "%", "_")

	return indicatorName
}

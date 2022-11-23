package common

import (
	"os"
	"strconv"
	"strings"
	"time"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
)

const ProblemURLLabel = "Problem URL"

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

// CreateDefaultSLOScore creates a keptncommon.SLOScore with default values.
func CreateDefaultSLOScore() keptncommon.SLOScore {
	return keptncommon.SLOScore{
		Pass:    "90%",
		Warning: "75%",
	}
}

// CreateDefaultSLOComparison creates a keptncommon.SLOComparison with default values.
func CreateDefaultSLOComparison() keptncommon.SLOComparison {
	return keptncommon.SLOComparison{
		CompareWith:               "single_result",
		IncludeResultWithScore:    "pass",
		NumberOfComparisonResults: 1,
		AggregateFunction:         "avg",
	}
}

package metrics

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
	"time"
)

// store url to the metrics api format migration document
const metricsAPIOldFormatNewFormatDoc = "https://github.com/keptn-contrib/dynatrace-sli-service/blob/master/docs/CustomQueryFormatMigration.md"

type QueryBuilder struct {
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
}

func NewQueryBuilder(eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter) *QueryBuilder {
	return &QueryBuilder{
		eventData:     eventData,
		customFilters: customFilters,
	}
}

// Build builds the complete query string based on start, end and filters
// metricQuery should contain metricSelector and entitySelector
// Returns:
//  #1: Dynatrace API metric query string
//  #2: Metric selector that this query will return, e.g: builtin:host.cpu
//  #3: error
func (b *QueryBuilder) Build(metricQuery string, startUnix time.Time, endUnix time.Time) (string, string, error) {
	// replace query params (e.g., $PROJECT, $STAGE, $SERVICE ...)
	metricQuery = common.ReplaceQueryParameters(metricQuery, b.customFilters, b.eventData)

	if strings.HasPrefix(metricQuery, "?metricSelector=") {
		log.WithFields(
			log.Fields{
				"query":        metricQuery,
				"helpDocument": metricsAPIOldFormatNewFormatDoc,
			}).Debug("COMPATIBILITY WARNING: query string is not compatible. Auto-removing the ? in front.")
		metricQuery = strings.Replace(metricQuery, "?metricSelector=", "metricSelector=", 1)
	}

	// split query string by first occurrence of "?"
	querySplit := strings.Split(metricQuery, "?")
	metricSelector := ""
	metricQueryParams := ""

	// support the old format with "metricSelector:someFilters()?scope=..." as well as the new format with
	// "?metricSelector=metricSelector&entitySelector=...&scope=..."
	if len(querySplit) == 1 {
		// new format without "?" -> everything within the query string are query parameters
		metricQueryParams = querySplit[0]
	} else {
		log.WithFields(
			log.Fields{
				"query":        metricQueryParams,
				"helpDocument": metricsAPIOldFormatNewFormatDoc,
			}).Debug("COMPATIBILITY WARNING: query uses the old format")
		// old format with "?" - everything left of the ? is the identifier, everything right are query params
		metricSelector = querySplit[0]

		// build the new query
		metricQueryParams = fmt.Sprintf("metricSelector=%s&%s", querySplit[0], querySplit[1])
	}

	q, err := url.ParseQuery(metricQueryParams)
	if err != nil {
		return "", "", fmt.Errorf("could not parse metrics URL: %s", err.Error())
	}

	q.Add("resolution", "Inf") // resolution=Inf means that we only get 1 datapoint (per service)
	q.Add("from", common.TimestampToString(startUnix))
	q.Add("to", common.TimestampToString(endUnix))

	// check if q contains "scope"
	scopeData := q.Get("scope")

	// compatibility with old scope=... custom queries
	if scopeData != "" {
		log.WithField("helpDocument", metricsAPIOldFormatNewFormatDoc).Debug("COMPATIBILITY WARNING: querying the new metrics API requires use of entitySelector rather than scope")
		// scope is no longer supported in the new API, it needs to be called "entitySelector" and contain type(SERVICE)
		if !strings.Contains(scopeData, "type(SERVICE)") {
			log.WithField("helpDocument", metricsAPIOldFormatNewFormatDoc).Debug("COMPATIBILITY WARNING: Automatically adding type(SERVICE) to entitySelector for compatibility with the new Metrics API")
			scopeData = fmt.Sprintf("%s,type(SERVICE)", scopeData)
		}
		// add scope as entitySelector
		q.Add("entitySelector", scopeData)
	}

	// check metricSelector
	if metricSelector == "" {
		metricSelector = q.Get("metricSelector")
	}

	return q.Encode(), metricSelector, nil
}

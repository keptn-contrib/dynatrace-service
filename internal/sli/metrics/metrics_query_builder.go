package metrics

import (
	"fmt"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

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

	// try to do the legacy query transformation
	transformedQuery, err := NewLegacyQueryTransformation(metricQuery).Transform()
	if err != nil {
		return "", "", fmt.Errorf("could not parse old format metrics query: %v, %w", metricQuery, err)
	}

	q, err := NewQueryParsing(transformedQuery).Parse()
	if err != nil {
		return "", "", fmt.Errorf("could not parse metrics query: %v, %w", transformedQuery, err)
	}

	// resolution=Inf means that we only get 1 datapoint (per service)
	err = q.Add(resolutionKey, "Inf")
	if err != nil {
		return "", "", err
	}

	err = q.Add(fromKey, common.TimestampToString(startUnix))
	if err != nil {
		return "", "", err
	}

	err = q.Add(toKey, common.TimestampToString(endUnix))
	if err != nil {
		return "", "", err
	}

	return q.Encode(), q.GetMetricSelector(), nil
}

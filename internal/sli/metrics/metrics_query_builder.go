package metrics

import (
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

type QueryBuilder struct {
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

// Build builds the complete query string based on start, end and filters
// metricQuery should contain metricSelector and entitySelector
// Returns:
//  #1: Dynatrace API metric query string
//  #2: Metric selector that this query will return, e.g: builtin:host.cpu
//  #3: error
func (b *QueryBuilder) Build(metricQuery string, startUnix time.Time, endUnix time.Time) (string, string, error) {
	q, err := NewQueryParsing(metricQuery).Parse()
	if err != nil {
		return "", "", fmt.Errorf("could not parse metrics query: %v, %w", metricQuery, err)
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

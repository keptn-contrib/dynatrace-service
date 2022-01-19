package dashboard

import (
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
)

type queryComponents struct {
	metricsQuery                *metrics.Query
	startTime                   time.Time
	endTime                     time.Time
	metricUnit                  string
	entitySelectorTargetSnippet string
	metricSelectorTargetSnippet string
}

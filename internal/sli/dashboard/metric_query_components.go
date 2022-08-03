package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
)

type queryComponents struct {
	metricsQuery metrics.Query
	timeframe    common.Timeframe
	metricUnit   string
}

package usql

import (
	"net/url"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
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

// Build builds a USQL query based on the incoming values
func (b *QueryBuilder) Build(query string, startUnix time.Time, endUnix time.Time) string {
	log.WithField("query", query).Debug("Finalize USQL query")

	// replace query params (e.g., $PROJECT, $STAGE, $SERVICE ...)
	// default query params that are required: resolution, from and to
	q := make(url.Values)
	q.Add("query", query)
	q.Add("explain", "false")
	q.Add("addDeepLinkFields", "false")
	q.Add("startTimestamp", common.TimestampToString(startUnix))
	q.Add("endTimestamp", common.TimestampToString(endUnix))

	return q.Encode()
}

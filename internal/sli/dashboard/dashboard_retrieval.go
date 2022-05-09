package dashboard

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type Retrieval struct {
	client    dynatrace.ClientInterface
	eventData adapter.EventContentAdapter
}

func NewRetrieval(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter) *Retrieval {
	return &Retrieval{
		client:    client,
		eventData: eventData,
	}
}

// Retrieve Depending on the dashboard parameter which is pulled from dynatrace.conf.yaml:dashboard this method either
//   - query:        queries all dashboards on the Dynatrace Tenant and returns the one that matches project/service/stage, or
//   - dashboard-ID: if this is a valid dashboard ID it will query the dashboard with this ID, e.g: ddb6a571-4bda-4e8b-a9c0-4a3e02c2e14a, or
// It returns a parsed Dynatrace Dashboard and the actual dashboard ID in case we queried a dashboard.
func (r *Retrieval) Retrieve(ctx context.Context, dashboard string) (*dynatrace.Dashboard, string, error) {
	// dashboard property is invalid
	if dashboard == "" {
		return nil, "", fmt.Errorf("invalid 'dashboard' property - either specify a dashboard ID or use 'query'")
	}

	// Option 1: Query dashboards
	if dashboard == common.DynatraceConfigDashboardQUERY {
		var err error
		dashboard, err = r.findDynatraceDashboard(ctx)
		if err != nil {
			log.WithError(err).WithFields(
				log.Fields{
					"project": r.eventData.GetProject(),
					"stage":   r.eventData.GetStage(),
					"service": r.eventData.GetService(),
				}).Debug("Dashboard option query but could not find KQG dashboard")
			return nil, "", err
		}

		log.WithFields(
			log.Fields{
				"project":   r.eventData.GetProject(),
				"stage":     r.eventData.GetStage(),
				"service":   r.eventData.GetService(),
				"dashboard": dashboard,
			}).Debug("Dashboard option query found for dashboard")
	}

	// Option 2: We (now) have a Dashboard UUID - so let's query it!
	log.WithField("dashboard", dashboard).Debug("Query dashboard")
	dynatraceDashboard, err := dynatrace.NewDashboardsClient(r.client).GetByID(ctx, dashboard)
	if err != nil {
		return nil, dashboard, err
	}

	return dynatraceDashboard, dashboard, nil
}

func (r *Retrieval) findDynatraceDashboard(ctx context.Context) (string, error) {
	dashboardList, err := dynatrace.NewDashboardsClient(r.client).GetAll(ctx)
	if err != nil {
		return "", err
	}

	return dashboardList.SearchForDashboardMatching(r.eventData.GetProject(), r.eventData.GetStage(), r.eventData.GetService())
}

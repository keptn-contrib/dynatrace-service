package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	log "github.com/sirupsen/logrus"
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
//   - <empty>:      it will not query any dashboard.
// It returns a parsed Dynatrace Dashboard and the actual dashboard ID in case we queried a dashboard.
func (r *Retrieval) Retrieve(dashboard string) (*dynatrace.Dashboard, string, error) {
	// Option 1: there is no dashboard we should query
	if dashboard == "" {
		return nil, dashboard, nil
	}

	// Option 2: Query dashboards
	if dashboard == common.DynatraceConfigDashboardQUERY {
		var err error
		dashboard, err = r.findDynatraceDashboard()
		if dashboard == "" || err != nil {
			log.WithError(err).WithFields(
				log.Fields{
					"project": r.eventData.GetProject(),
					"stage":   r.eventData.GetStage(),
					"service": r.eventData.GetService(),
				}).Debug("Dashboard option query but could not find KQG dashboard")

			// TODO 2021-08-03: should this really return no error, if querying dashboards found no match?
			// this would be the same result as option 1 then
			return nil, dashboard, nil
		}

		log.WithFields(
			log.Fields{
				"project":   r.eventData.GetProject(),
				"stage":     r.eventData.GetStage(),
				"service":   r.eventData.GetService(),
				"dashboard": dashboard,
			}).Debug("Dashboard option query found for dashboard")
	}

	// We have a Dashboard UUID - now lets query it!
	log.WithField("dashboard", dashboard).Debug("Query dashboard")
	dynatraceDashboard, err := dynatrace.NewDashboardsClient(r.client).GetByID(dashboard)
	if err != nil {
		return nil, dashboard, err
	}

	return dynatraceDashboard, dashboard, nil
}

func (r *Retrieval) findDynatraceDashboard() (string, error) {
	// Lets query the list of all Dashboards and find the one that matches project, stage, service based on the title (in the future - we can do it via tags)
	// create dashboard query URL and set additional headers
	dashboards, err := dynatrace.NewDashboardsClient(r.client).GetAll()
	if err != nil {
		return "", err
	}

	return dashboards.SearchForDashboardMatching(r.eventData.GetProject(), r.eventData.GetStage(), r.eventData.GetService()), nil
}

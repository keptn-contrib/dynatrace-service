package sli

import (
	"context"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/ff"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/dashboard"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/query"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const NoMetricIndicator = "no_metric"

type GetSLIEventHandler struct {
	event             GetSLITriggeredAdapterInterface
	dtClient          dynatrace.ClientInterface
	eventSenderClient keptn.EventSenderClientInterface
	configClient      configClientInterface
	secretName        string
	dashboardProperty string
	featureFlags      ff.GetSLIFeatureFlags
}

// configClientInterface is a subset of a keptn.ConfigClientInterface for processing sh.keptn.event.get-sli.triggered events.
// It can read SLIs and read and write SLOs.
type configClientInterface interface {

	// GetSLIs gets the SLIs stored for the specified project, stage and service.
	GetSLIs(ctx context.Context, project string, stage string, service string) (map[string]string, error)

	// GetSLOs gets the SLOs stored for exactly the specified project, stage and service.
	GetSLOs(ctx context.Context, project string, stage string, service string) (*keptncommon.ServiceLevelObjectives, error)

	// UploadSLOs uploads the SLOs for the specified project, stage and service.
	UploadSLOs(ctx context.Context, project string, stage string, service string, slos *keptncommon.ServiceLevelObjectives) error
}

func NewGetSLITriggeredHandler(event GetSLITriggeredAdapterInterface, dtClient dynatrace.ClientInterface, eventSenderClient keptn.EventSenderClientInterface, configClient configClientInterface, secretName string, dashboardProperty string, flags ff.GetSLIFeatureFlags) GetSLIEventHandler {
	return GetSLIEventHandler{
		event:             event,
		dtClient:          dtClient,
		eventSenderClient: eventSenderClient,
		configClient:      configClient,
		secretName:        secretName,
		dashboardProperty: dashboardProperty,
		featureFlags:      flags,
	}
}

// HandleEvent handles a get-SLI triggered event.
func (eh GetSLIEventHandler) HandleEvent(workCtx context.Context, _ context.Context) error {
	if err := eh.sendGetSLIStartedEvent(); err != nil {
		return err
	}

	getSLIFinishedEventFactory := eh.processEvent(workCtx)

	log.Info("Finished retrieving SLI results, sending sh.keptn.event.get-sli.finished event now...")
	return eh.sendEvent(getSLIFinishedEventFactory)
}

func (eh *GetSLIEventHandler) processEvent(ctx context.Context) *GetSLIFinishedEventFactory {
	log.WithFields(
		log.Fields{
			"project": eh.event.GetProject(),
			"stage":   eh.event.GetStage(),
			"service": eh.event.GetService(),
		}).Info("Processing sh.keptn.event.get-sli.triggered")

	results, err := eh.retrieveResults(ctx)
	if err != nil {
		log.WithError(err).Error("error retrieving SLIs")
		return NewSuccessfulGetSLIFinishedEventFactoryFromError(eh.event, err)
	}

	// log SLI results
	for _, r := range results {
		if r.SLIResult().IndicatorResult == result.IndicatorResultSuccessful {
			log.WithField("sliWithSLO", r).Debug("Retrieved SLI result")
			continue
		}

		log.WithField("sliWithSLO", r).Warn("Failed to retrieve SLI result")
	}

	return NewSuccessfulGetSLIFinishedEventFactoryFromResults(eh.event, results)
}

// retrieveResults will retrieve metrics either from a dashboard or from an SLI file.
func (eh *GetSLIEventHandler) retrieveResults(ctx context.Context) ([]result.SLIWithSLO, error) {
	// Adding DtCreds as a label so users know which DtCreds was used
	eh.event.AddLabel("DtCreds", eh.secretName)

	timeframe, err := common.NewTimeframeParser(eh.event.GetSLIStart(), eh.event.GetSLIEnd()).Parse()
	if err != nil {
		return nil, err
	}

	return eh.getResults(ctx, *timeframe)
}

func (eh *GetSLIEventHandler) getResults(ctx context.Context, timeframe common.Timeframe) ([]result.SLIWithSLO, error) {
	// If no dashboard specified, query the SLIs based on the SLI.yaml definition
	if eh.dashboardProperty == "" {
		return eh.getResultsFromCustomQueries(ctx, timeframe)
	}

	return eh.getResultsFromDynatraceDashboard(ctx, timeframe)
}

// getResultsFromDynatraceDashboard will process dynatrace dashboard (if found) and return SLIWithSLO
func (eh *GetSLIEventHandler) getResultsFromDynatraceDashboard(ctx context.Context, timeframe common.Timeframe) ([]result.SLIWithSLO, error) {
	d, err := dashboard.NewRetrieval(eh.dtClient, eh.event).Retrieve(ctx, eh.dashboardProperty)
	if err != nil {
		return nil, dashboard.NewDashboardError(err)
	}

	eh.event.AddLabel("Dashboard Link", dashboard.NewLink(eh.dtClient.Credentials().GetTenant(), timeframe, d.ID, d.GetFilter()).String())

	results, err := dashboard.NewProcessing(eh.dtClient, eh.event, eh.event.GetCustomSLIFilters(), timeframe, eh.configClient, eh.featureFlags).Process(ctx, d)
	if err != nil {
		return nil, dashboard.NewDashboardError(err)
	}

	return results, nil
}

func (eh *GetSLIEventHandler) getResultsFromCustomQueries(ctx context.Context, timeframe common.Timeframe) ([]result.SLIWithSLO, error) {
	indicators := eh.event.GetIndicators()
	if len(indicators) == 0 {
		return []result.SLIWithSLO{}, nil
	}

	slis, err := eh.configClient.GetSLIs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
	if err != nil {
		log.WithError(err).Error("could not retrieve custom SLI definitions")
		return nil, fmt.Errorf("could not retrieve custom SLI definitions: %w", err)
	}

	return query.NewProcessing(eh.dtClient, eh.event, eh.event.GetCustomSLIFilters(), query.NewCustomQueries(slis), timeframe, eh.configClient).Process(ctx, indicators)
}

func (eh *GetSLIEventHandler) sendGetSLIStartedEvent() error {
	return eh.sendEvent(NewGetSLIStartedEventFactory(eh.event))
}

func (eh *GetSLIEventHandler) sendEvent(factory adapter.CloudEventFactoryInterface) error {
	err := eh.eventSenderClient.SendCloudEvent(factory)
	if err != nil {
		log.WithError(err).Error("Could not send get-sli cloud event")
		return err
	}

	return nil
}

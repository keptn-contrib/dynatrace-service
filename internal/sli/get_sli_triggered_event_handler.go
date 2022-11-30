package sli

import (
	"context"
	"errors"
	"fmt"

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

func NewGetSLITriggeredHandler(event GetSLITriggeredAdapterInterface, dtClient dynatrace.ClientInterface, eventSenderClient keptn.EventSenderClientInterface, configClient configClientInterface, secretName string, dashboardProperty string) GetSLIEventHandler {
	return GetSLIEventHandler{
		event:             event,
		dtClient:          dtClient,
		eventSenderClient: eventSenderClient,
		configClient:      configClient,
		secretName:        secretName,
		dashboardProperty: dashboardProperty,
	}
}

// HandleEvent handles a get-SLI triggered event.
func (eh GetSLIEventHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
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

	processingResult, err := eh.retrieveSLIResults(ctx)
	if err != nil {
		log.WithError(err).Error("error retrieving SLIs")
		return NewSucceededGetSLIFinishedEventFactory(eh.event, makeSLIResultsForError(err, eh.event), err)
	}

	// log SLI results
	for _, sliResult := range processingResult.SLIResults() {
		if sliResult.IndicatorResult == result.IndicatorResultSuccessful {
			log.WithField("sliResult", sliResult).Debug("Retrieved SLI result")
			continue
		}

		log.WithField("sliResult", sliResult).Warn("Failed to retrieve SLI result")
	}

	return NewSucceededGetSLIFinishedEventFactory(eh.event, processingResult.SLIResults(), err)
}

// retrieveSLIResults will retrieve metrics either from a dashboard or from an SLI file.
func (eh *GetSLIEventHandler) retrieveSLIResults(ctx context.Context) (*result.ProcessingResult, error) {
	// Adding DtCreds as a label so users know which DtCreds was used
	eh.event.AddLabel("DtCreds", eh.secretName)

	timeframe, err := common.NewTimeframeParser(eh.event.GetSLIStart(), eh.event.GetSLIEnd()).Parse()
	if err != nil {
		return nil, err
	}

	processingResult, err := eh.getSLIResults(ctx, *timeframe)
	if err != nil {
		return nil, err
	}

	// if no result values have been captured, return an error
	if len(processingResult.SLIResults()) == 0 {
		return nil, errors.New("could not retrieve any SLI results")
	}

	return processingResult, nil
}

func (eh *GetSLIEventHandler) getSLIResults(ctx context.Context, timeframe common.Timeframe) (*result.ProcessingResult, error) {
	// If no dashboard specified, query the SLIs based on the SLI.yaml definition
	if eh.dashboardProperty == "" {
		return eh.getSLIResultsFromCustomQueries(ctx, timeframe)
	}

	return eh.getSLIResultsFromDynatraceDashboard(ctx, timeframe)
}

// getSLIResultsFromDynatraceDashboard will process dynatrace dashboard (if found) and return SLIResults
func (eh *GetSLIEventHandler) getSLIResultsFromDynatraceDashboard(ctx context.Context, timeframe common.Timeframe) (*result.ProcessingResult, error) {
	d, err := dashboard.NewRetrieval(eh.dtClient, eh.event).Retrieve(ctx, eh.dashboardProperty)
	if err != nil {
		return nil, dashboard.NewQueryError(fmt.Errorf("error while retrieving dashboard: %w", err))
	}

	eh.event.AddLabel("Dashboard Link", dashboard.NewLink(eh.dtClient.Credentials().GetTenant(), timeframe, d.ID, d.GetFilter()).String())

	processingResult, err := dashboard.NewProcessing(eh.dtClient, eh.event, eh.event.GetCustomSLIFilters(), timeframe).Process(ctx, d)
	if err != nil {
		return nil, dashboard.NewQueryError(err)
	}

	// let's write the SLO to the config repo
	if processingResult.HasSLOs() {
		err = eh.configClient.UploadSLOs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService(), processingResult.SLOs())
		if err != nil {
			return nil, dashboard.NewUploadFileError("SLO", err)
		}
	}

	return processingResult, nil
}

func (eh *GetSLIEventHandler) getSLIResultsFromCustomQueries(ctx context.Context, timeframe common.Timeframe) (*result.ProcessingResult, error) {
	indicators := eh.event.GetIndicators()
	if len(indicators) == 0 {
		return nil, errors.New("no SLIs were requested")
	}

	slis, err := eh.configClient.GetSLIs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
	if err != nil {
		log.WithError(err).Error("could not retrieve custom SLI definitions")
		return nil, fmt.Errorf("could not retrieve custom SLI definitions: %w", err)
	}

	sliResults := query.NewProcessing(eh.dtClient, eh.event, eh.event.GetCustomSLIFilters(), query.NewCustomQueries(slis), timeframe).Process(ctx, indicators)

	slos, err := eh.configClient.GetSLOs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
	if err != nil {
		log.WithError(err).Error("could not retrieve SLO definitions")
		return nil, fmt.Errorf("could not retrieve SLO definitions: %w", err)
	}

	return result.NewProcessingResult(slos, sliResults), nil
}

func (eh *GetSLIEventHandler) sendGetSLIStartedEvent() error {
	return eh.sendEvent(NewGetSLIStartedEventFactory(eh.event))
}

func makeSLIResultsForError(err error, eventData GetSLITriggeredAdapterInterface) []result.SLIResult {
	indicators := eventData.GetIndicators()

	var errType *dashboard.ProcessingError
	if len(indicators) == 0 || errors.As(err, &errType) {
		return []result.SLIResult{result.NewFailedSLIResult(NoMetricIndicator, err.Error())}
	}

	sliResults := make([]result.SLIResult, len(indicators))
	for i, indicatorName := range indicators {
		sliResults[i] = result.NewFailedSLIResult(indicatorName, err.Error())
	}

	return sliResults
}

func (eh *GetSLIEventHandler) sendEvent(factory adapter.CloudEventFactoryInterface) error {
	err := eh.eventSenderClient.SendCloudEvent(factory)
	if err != nil {
		log.WithError(err).Error("Could not send get-sli cloud event")
		return err
	}

	return nil
}

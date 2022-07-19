package sli

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

const ProblemOpenSLI = "problem_open"
const NoMetricIndicator = "no metric"

type GetSLIEventHandler struct {
	event          GetSLITriggeredAdapterInterface
	dtClient       dynatrace.ClientInterface
	kClient        keptn.ClientInterface
	resourceClient keptn.SLOAndSLIClientInterface

	secretName string
	dashboard  string
}

func NewGetSLITriggeredHandler(event GetSLITriggeredAdapterInterface, dtClient dynatrace.ClientInterface, kClient keptn.ClientInterface, resourceClient keptn.SLOAndSLIClientInterface, secretName string, dashboard string) GetSLIEventHandler {
	return GetSLIEventHandler{
		event:          event,
		dtClient:       dtClient,
		kClient:        kClient,
		resourceClient: resourceClient,
		secretName:     secretName,
		dashboard:      dashboard,
	}
}

// HandleEvent handles a get-SLI triggered event.
func (eh GetSLIEventHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	if err := eh.sendGetSLIStartedEvent(); err != nil {
		return err
	}

	log.WithFields(
		log.Fields{
			"project": eh.event.GetProject(),
			"stage":   eh.event.GetStage(),
			"service": eh.event.GetService(),
		}).Info("Processing sh.keptn.event.get-sli.triggered")

	sliResults, err := eh.retrieveSLIResults(workCtx)
	if err != nil {
		log.WithError(err).Error("error retrieving SLIs")
		return eh.sendGetSLIFinishedEvent(nil, err)
	}

	log.Info("Finished retrieving SLI results, sending sh.keptn.event.get-sli.finished event now...")
	return eh.sendGetSLIFinishedEvent(sliResults, err)
}

// retrieveSLIResults will retrieve metrics either from a dashboard or from an SLI file.
func (eh *GetSLIEventHandler) retrieveSLIResults(ctx context.Context) ([]result.SLIResult, error) {
	// Adding DtCreds as a label so users know which DtCreds was used
	eh.event.AddLabel("DtCreds", eh.secretName)

	timeframe, err := common.NewTimeframeParser(eh.event.GetSLIStart(), eh.event.GetSLIEnd()).Parse()
	if err != nil {
		return nil, err
	}

	sliResults, err := eh.getSLIResults(ctx, *timeframe)
	if err != nil {
		return nil, err
	}

	// ARE WE CALLED IN CONTEXT OF A PROBLEM REMEDIATION??
	// If so - we should try to query the status of the Dynatrace Problem that triggered this evaluation
	problemID := keptn.TryGetProblemIDFromLabels(eh.event)
	if problemID != "" {
		sliResults = append(sliResults, eh.getSLIResultsFromProblemContext(ctx, problemID))
	}

	// if no result values have been captured, return an error
	if len(sliResults) == 0 {
		return nil, errors.New("could not retrieve any SLI results")
	}

	return sliResults, nil
}

func (eh *GetSLIEventHandler) getSLIResults(ctx context.Context, timeframe common.Timeframe) ([]result.SLIResult, error) {
	// If no dashboard specified, query the SLIs based on the SLI.yaml definition
	if eh.dashboard == "" {
		return eh.getSLIResultsFromCustomQueries(ctx, timeframe)
	}

	// See if we can get the data from a Dynatrace Dashboard
	var dashboardLinkAsLabel *dashboard.DashboardLink
	dashboardLinkAsLabel, sliResults, err := eh.getSLIResultsFromDynatraceDashboard(ctx, timeframe)
	if err != nil {
		return nil, err
	}

	// add link to dynatrace dashboard to labels
	if dashboardLinkAsLabel != nil {
		eh.event.AddLabel("Dashboard Link", dashboardLinkAsLabel.String())
	}
	return sliResults, nil
}

// getSLIResultsFromDynatraceDashboard will process dynatrace dashboard (if found) and return SLIResults
func (eh *GetSLIEventHandler) getSLIResultsFromDynatraceDashboard(ctx context.Context, timeframe common.Timeframe) (*dashboard.DashboardLink, []result.SLIResult, error) {

	sliQuerying := dashboard.NewQuerying(eh.event, eh.event.GetCustomSLIFilters(), eh.dtClient)
	queryResult, err := sliQuerying.GetSLIValues(ctx, eh.dashboard, timeframe)
	if err != nil {
		return nil, nil, dashboard.NewQueryError(err)
	}

	// let's write the SLI to the config repo
	if queryResult.HasSLIs() {
		err = eh.resourceClient.UploadSLIs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService(), queryResult.SLIs())
		if err != nil {
			return nil, nil, dashboard.NewUploadFileError("SLI", err)
		}
	}

	// let's write the SLO to the config repo
	if queryResult.HasSLOs() {
		err = eh.resourceClient.UploadSLOs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService(), queryResult.SLOs())
		if err != nil {
			return nil, nil, dashboard.NewUploadFileError("SLO", err)
		}
	}

	return queryResult.DashboardLink(), queryResult.SLIResults(), nil
}

func (eh *GetSLIEventHandler) getSLIResultsFromCustomQueries(ctx context.Context, timeframe common.Timeframe) ([]result.SLIResult, error) {
	slis, err := eh.resourceClient.GetSLIs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
	if err != nil {
		log.WithError(err).Error("could not retrieve custom SLI definitions")
		return nil, fmt.Errorf("could not retrieve custom SLI definitions: %w", err)
	}

	queryProcessing := query.NewProcessing(eh.dtClient, eh.event, eh.event.GetCustomSLIFilters(), query.NewCustomQueries(slis), timeframe)

	var sliResults []result.SLIResult

	// query all indicators
	for _, indicator := range eh.event.GetIndicators() {
		if strings.Compare(indicator, ProblemOpenSLI) == 0 {
			log.WithField("indicator", indicator).Info("Skipping indicator as it is handled later")
			continue
		}

		sliResults = append(sliResults, queryProcessing.GetSLIResultFromIndicator(ctx, indicator))
	}

	return sliResults, nil
}

func createDefaultProblemSLO() *keptncommon.SLO {
	return &keptncommon.SLO{
		SLI: ProblemOpenSLI,
		Pass: []*keptncommon.SLOCriteria{
			{
				Criteria: []string{"pass=<=0"},
			},
		},
		Weight: 1,
		KeySLI: true,
	}
}

func (eh *GetSLIEventHandler) getSLIResultsFromProblemContext(ctx context.Context, problemID string) result.SLIResult {
	// let's add this to the SLO in case this indicator is not yet in SLO.yaml.
	// Because if it does not get added the lighthouse will not evaluate the SLI values
	// we default it to open_problems<=0
	errAddSLO := eh.addSLO(ctx, createDefaultProblemSLO())
	if errAddSLO != nil {
		// TODO 2021-08-10: should this be added to the error object for sendGetSLIFinishedEvent below?
		log.WithError(errAddSLO).Error("problem while adding SLOs")
	}

	status, err := dynatrace.NewProblemsV2Client(eh.dtClient).GetStatusByID(ctx, problemID)
	if err != nil {
		return result.NewFailedSLIResult(ProblemOpenSLI, err.Error())
	}

	switch status {
	case dynatrace.ProblemStatusOpen:
		return result.NewSuccessfulSLIResult(ProblemOpenSLI, 1.0)
	case "":
		return result.NewFailedSLIResult(ProblemOpenSLI, "Unexpected empty status")
	default:
		return result.NewSuccessfulSLIResult(ProblemOpenSLI, 0)
	}
}

// addSLO adds an SLO Entry to the SLO.yaml
func (eh GetSLIEventHandler) addSLO(ctx context.Context, newSLO *keptncommon.SLO) error {

	// first - lets load the SLO.yaml from the config repo
	dashboardSLO, err := eh.resourceClient.GetSLOs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
	if err != nil {
		var rnfErr *keptn.ResourceNotFoundError
		if !errors.As(err, &rnfErr) {
			return err
		}

		// this is the default SLO in case none has yet been uploaded
		dashboardSLO = &keptncommon.ServiceLevelObjectives{
			Objectives: []*keptncommon.SLO{},
			TotalScore: &keptncommon.SLOScore{
				Pass:    "90%",
				Warning: "75%"},
			Comparison: &keptncommon.SLOComparison{
				CompareWith:               "single_result",
				IncludeResultWithScore:    "pass",
				NumberOfComparisonResults: 1,
				AggregateFunction:         "avg"},
		}
	}

	// now we add the SLO Definition to the objectives - but first validate if it is not already there
	for _, objective := range dashboardSLO.Objectives {
		if objective.SLI == newSLO.SLI {
			return nil
		}
	}

	// now - lets add our newSLO to the list
	dashboardSLO.Objectives = append(dashboardSLO.Objectives, newSLO)
	err = eh.resourceClient.UploadSLOs(ctx, eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService(), dashboardSLO)
	if err != nil {
		return err
	}

	return nil
}

func (eh *GetSLIEventHandler) sendGetSLIStartedEvent() error {
	return eh.sendEvent(NewGetSLIStartedEventFactory(eh.event))
}

// sendGetSLIFinishedEvent sends the SLI finished event. If err != nil it will send an error message
func (eh *GetSLIEventHandler) sendGetSLIFinishedEvent(sliResults []result.SLIResult, err error) error {
	// if an error was set - the SLI results will be set to failed and an error message is set to each
	sliResults = resetSLIResultsInCaseOfError(err, eh.event, sliResults)

	return eh.sendEvent(NewSucceededGetSLIFinishedEventFactory(eh.event, sliResults, err))
}

func resetSLIResultsInCaseOfError(err error, eventData GetSLITriggeredAdapterInterface, sliResults []result.SLIResult) []result.SLIResult {
	if err == nil {
		return sliResults
	}

	indicators := eventData.GetIndicators()
	if len(sliResults) == 0 {
		var errType *dashboard.QueryError
		if len(indicators) == 0 || errors.As(err, &errType) {
			indicators = []string{NoMetricIndicator}
		}

		for _, indicatorName := range indicators {
			sliResults = []result.SLIResult{
				result.NewFailedSLIResult(indicatorName, ""),
			}
		}
	}

	erroredSLIResults := make([]result.SLIResult, 0, len(sliResults))
	for _, sliResult := range sliResults {
		erroredSLIResults = append(erroredSLIResults, result.NewFailedSLIResult(sliResult.Metric(), err.Error()))
	}

	return erroredSLIResults
}

func (eh *GetSLIEventHandler) sendEvent(factory adapter.CloudEventFactoryInterface) error {
	err := eh.kClient.SendCloudEvent(factory)
	if err != nil {
		log.WithError(err).Error("Could not send get sli cloud event")
		return err
	}

	return nil
}

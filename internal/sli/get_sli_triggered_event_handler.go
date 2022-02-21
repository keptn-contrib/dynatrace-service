package sli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/dashboard"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/query"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	// configutils "github.com/keptn/go-utils/pkg/configuration-service/utils"
	// keptnevents "github.com/keptn/go-utils/pkg/events"
	// keptnutils "github.com/keptn/go-utils/pkg/utils"
	// v1 "k8s.io/client-go/kubernetes/typed/core/// "
)

const ProblemOpenSLI = "problem_open"
const NoMetricIndicator = "no metric"

type GetSLIEventHandler struct {
	event          GetSLITriggeredAdapterInterface
	dtClient       dynatrace.ClientInterface
	kClient        keptn.ClientInterface
	resourceClient keptn.ResourceClientInterface

	secretName string
	dashboard  string
}

func NewGetSLITriggeredHandler(event GetSLITriggeredAdapterInterface, dtClient dynatrace.ClientInterface, kClient keptn.ClientInterface, resourceClient keptn.ResourceClientInterface, secretName string, dashboard string) GetSLIEventHandler {
	return GetSLIEventHandler{
		event:          event,
		dtClient:       dtClient,
		kClient:        kClient,
		resourceClient: resourceClient,
		secretName:     secretName,
		dashboard:      dashboard,
	}
}

func (eh GetSLIEventHandler) HandleEvent() error {
	// do not continue if SLIProvider is not dynatrace
	if eh.event.IsNotForDynatrace() {
		return nil
	}

	if err := eh.sendGetSLIStartedEvent(); err != nil {
		return err
	}

	log.WithFields(
		log.Fields{
			"project": eh.event.GetProject(),
			"stage":   eh.event.GetStage(),
			"service": eh.event.GetService(),
		}).Info("Processing sh.keptn.event.get-sli.triggered")

	sliResults, err := eh.retrieveSLIResults()
	if err != nil {
		log.WithError(err).Error("error retrieving SLIs")
		return eh.sendGetSLIFinishedEvent(nil, err)
	}

	log.Info("Finished retrieving SLI results, sending sh.keptn.event.get-sli.finished event now...")
	return eh.sendGetSLIFinishedEvent(sliResults, err)
}

// retrieveSLIResults will retrieve metrics either from a dashboard or from an SLI file
func (eh *GetSLIEventHandler) retrieveSLIResults() ([]*keptnv2.SLIResult, error) {
	// Adding DtCreds as a label so users know which DtCreds was used
	eh.event.AddLabel("DtCreds", eh.secretName)

	// parse start and end (which are datetime strings) and convert them into unix timestamps
	startUnix, endUnix, err := ensureRightTimestamps(eh.event.GetSLIStart(), eh.event.GetSLIEnd())
	if err != nil {
		return nil, err
	}

	sliResults, err := eh.getSLIResults(startUnix, endUnix)
	if err != nil {
		return nil, err
	}

	// ARE WE CALLED IN CONTEXT OF A PROBLEM REMEDIATION??
	// If so - we should try to query the status of the Dynatrace Problem that triggered this evaluation
	problemID := getDynatraceProblemContext(eh.event)
	if problemID != "" {
		sliResults = append(sliResults, eh.getSLIResultsFromProblemContext(problemID))
	}

	// if no result values have been captured, return an error
	if len(sliResults) == 0 {
		return nil, errors.New("could not retrieve any SLI results")
	}

	return sliResults, nil
}

/**
 * AG-27052020: When using keptn send event start-evaluation and clocks are not 100% in sync, e.g: workstation is 1-2 seconds off
 *              we might run into the issue that we detect the endtime to be in the future. I ran into this problem after my laptop ran out of sync for about 1.5s
 *              to circumvent this issue I am changing the check to also allow a time difference of up to 2 minutes (120 seconds). This shouldnt be a problem as our SLI Service retries the DYnatrace API anyway
 * Here is the issue: https://github.com/keptn-contrib/dynatrace-sli-service/issues/55
 */
func ensureRightTimestamps(start string, end string) (time.Time, time.Time, error) {

	startUnix, err := common.ParseUnixTimestamp(start)
	if err != nil {
		return time.Now(), time.Now(), errors.New("Error parsing start date: " + err.Error())
	}
	endUnix, err := common.ParseUnixTimestamp(end)
	if err != nil {
		return startUnix, time.Now(), errors.New("Error parsing end date: " + err.Error())
	}

	// ensure end time is not in the future
	now := time.Now()
	timeDiffInSeconds := now.Sub(endUnix).Seconds()
	if timeDiffInSeconds < -120 { // used to be 0
		return startUnix, endUnix, fmt.Errorf("error validating time range: Supplied end-time %v is too far (>120seconds) in the future (now: %v - diff in sec: %v)\n", endUnix, now, timeDiffInSeconds)
	}

	// ensure start time is before end time
	timeframeInSeconds := endUnix.Sub(startUnix).Seconds()
	if timeframeInSeconds < 0 {
		return startUnix, endUnix, errors.New("error validating time range: start time needs to be before end time")
	}

	// AG-2020-07-16: Wait so Dynatrace has enough data but dont wait every time to shorten processing time
	// if we have a very short evaluation window and the end timestampe is now then we need to give Dynatrace some time to make sure we have relevant data
	// if the evalutaion timeframe is > 2 minutes we dont wait and just live with the fact that we may miss one minute or two at the end

	waitForSeconds := 120.0        // by default lets make sure we are at least 120 seconds away from "now()"
	if timeframeInSeconds >= 300 { // if our evaluated timeframe however is larger than 5 minutes its ok to continue right away. 5 minutes is the default timeframe for most evaluations
		waitForSeconds = 0.0
	} else if timeframeInSeconds >= 120 { // if the evaluation span is between 2 and 5 minutes make sure we at least have the last minute of data
		waitForSeconds = 60.0
	}

	// log output while we are waiting
	if time.Now().Sub(endUnix).Seconds() < waitForSeconds {
		log.Debug("As the end date is too close to Now() we are going to wait to make sure we have all the data for the requested timeframe(start-end)")
	}

	// make sure the end timestamp is at least waitForSeconds seconds in the past such that dynatrace metrics API has processed data
	for time.Now().Sub(endUnix).Seconds() < waitForSeconds {
		log.WithField("sleepSeconds", int(waitForSeconds-time.Now().Sub(endUnix).Seconds())).Debug("Sleeping while waiting for Dynatrace Metrics API")
		time.Sleep(10 * time.Second)
	}

	return startUnix, endUnix, nil
}

func (eh *GetSLIEventHandler) getSLIResults(startUnix time.Time, endUnix time.Time) ([]*keptnv2.SLIResult, error) {
	// If no dashboard specified, query the SLIs based on the SLI.yaml definition
	if eh.dashboard == "" {
		return eh.getSLIResultsFromCustomQueries(startUnix, endUnix)
	}

	// See if we can get the data from a Dynatrace Dashboard
	var dashboardLinkAsLabel *dashboard.DashboardLink
	dashboardLinkAsLabel, sliResults, err := eh.getSLIResultsFromDynatraceDashboard(startUnix, endUnix)
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
func (eh *GetSLIEventHandler) getSLIResultsFromDynatraceDashboard(startUnix time.Time, endUnix time.Time) (*dashboard.DashboardLink, []*keptnv2.SLIResult, error) {

	sliQuerying := dashboard.NewQuerying(eh.event, eh.event.GetCustomSLIFilters(), eh.dtClient)
	result, err := sliQuerying.GetSLIValues(eh.dashboard, startUnix, endUnix)
	if err != nil {
		return nil, nil, dashboard.NewQueryError(err)
	}

	// let's write the SLI to the config repo
	if result.HasSLIs() {
		err = eh.resourceClient.UploadSLI(eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService(), result.SLI())
		if err != nil {
			return nil, nil, dashboard.NewUploadFileError("SLI", err)
		}
	}

	// let's write the SLO to the config repo
	if result.HasSLOs() {
		err = eh.resourceClient.UploadSLOs(eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService(), result.SLO())
		if err != nil {
			return nil, nil, dashboard.NewUploadFileError("SLO", err)
		}
	}

	return result.DashboardLink(), result.SLIResults(), nil
}

func (eh *GetSLIEventHandler) getSLIResultsFromCustomQueries(startUnix time.Time, endUnix time.Time) ([]*keptnv2.SLIResult, error) {
	// get custom metrics for project if they exist
	projectCustomQueries, err := eh.kClient.GetCustomQueries(eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
	if err != nil {
		log.WithError(err).Errorf("could not retrieve custom queries: %v", err)
		return nil, fmt.Errorf("could not retrieve custom SLI definitions: %w", err)
	}

	queryProcessing := query.NewProcessing(eh.dtClient, eh.event, eh.event.GetCustomSLIFilters(), projectCustomQueries, startUnix, endUnix)

	var sliResults []*keptnv2.SLIResult

	// query all indicators
	for _, indicator := range eh.event.GetIndicators() {
		if strings.Compare(indicator, ProblemOpenSLI) == 0 {
			log.WithField("indicator", indicator).Info("Skipping indicator as it is handled later")
			continue
		}

		sliResults = append(sliResults, getSLIResultFromIndicator(indicator, queryProcessing))
	}

	return sliResults, nil
}

func getSLIResultFromIndicator(indicator string, queryProcessing *query.Processing) *keptnv2.SLIResult {
	log.WithField("indicator", indicator).Info("Fetching indicator")

	sliValue, err := queryProcessing.GetSLIValue(indicator)
	if err != nil {
		// failed to fetch metric
		log.WithError(err).Error("GetSLIValue failed")
		return &keptnv2.SLIResult{
			Metric:  indicator,
			Value:   0,
			Success: false, // mark as failure
			Message: err.Error(),
		}
	}

	// successfully fetched metric
	return &keptnv2.SLIResult{
		Metric:  indicator,
		Value:   sliValue,
		Success: true, // mark as success
	}
}

//getDynatraceProblemContext will evaluate the event and - returns dynatrace problem ID if found, 0 otherwise
func getDynatraceProblemContext(eventData GetSLITriggeredAdapterInterface) string {

	// iterate through the labels and find Problem URL
	if eventData.GetLabels() == nil || len(eventData.GetLabels()) == 0 {
		return ""
	}

	for labelName, labelValue := range eventData.GetLabels() {
		if strings.ToLower(labelName) == "problem url" {
			// the value should be of form https://dynatracetenant/#problems/problemdetails;pid=8485558334848276629_1604413609638V2
			// so - lets get the last part after pid=

			ix := strings.LastIndex(labelValue, ";pid=")
			if ix > 0 {
				return labelValue[ix+5:]
			}
		}
	}

	return ""
}

func (eh *GetSLIEventHandler) getSLIResultsFromProblemContext(problemID string) *keptnv2.SLIResult {
	problemIndicator := ProblemOpenSLI
	openProblemValue := 0.0
	success := false
	message := ""

	// let's query the status of this problem and add it to the SLI Result
	status, err := dynatrace.NewProblemsV2Client(eh.dtClient).GetStatusByID(problemID)
	if err != nil {
		message = err.Error()
	}

	if status != "" {
		success = true
		if status == dynatrace.ProblemStatusOpen {
			openProblemValue = 1.0
		}
	}

	// let's add this to the sliResults
	sliResult := &keptnv2.SLIResult{
		Metric:  problemIndicator,
		Value:   openProblemValue,
		Success: success,
		Message: message,
	}

	// let's add this to the SLO in case this indicator is not yet in SLO.yaml.
	// Because if it does not get added the lighthouse will not evaluate the SLI values
	// we default it to open_problems<=0
	sloString := fmt.Sprintf("sli=%s;pass=<=0;key=true", problemIndicator)
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(sloString)

	errAddSlo := eh.addSLO(sloDefinition)
	if errAddSlo != nil {
		// TODO 2021-08-10: should this be added to the error object for sendGetSLIFinishedEvent below?
		log.WithError(errAddSlo).Error("problem while adding SLOs")
	}

	return sliResult
}

// addSLO adds an SLO Entry to the SLO.yaml
func (eh GetSLIEventHandler) addSLO(newSLO *keptncommon.SLO) error {

	// first - lets load the SLO.yaml from the config repo
	dashboardSLO, err := eh.resourceClient.GetSLOs(eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService())
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
	err = eh.resourceClient.UploadSLOs(eh.event.GetProject(), eh.event.GetStage(), eh.event.GetService(), dashboardSLO)
	if err != nil {
		return err
	}

	return nil
}

func (eh *GetSLIEventHandler) sendGetSLIStartedEvent() error {
	return eh.sendEvent(NewGetSliStartedEventFactory(eh.event))
}

// sendGetSLIFinishedEvent sends the SLI finished event. If err != nil it will send an error message
func (eh *GetSLIEventHandler) sendGetSLIFinishedEvent(sliResults []*keptnv2.SLIResult, err error) error {
	// if an error was set - the SLI results will be set to failed and an error message is set to each
	sliResults = resetSLIResultsInCaseOfError(err, eh.event, sliResults)

	return eh.sendEvent(NewSucceededGetSLIFinishedEventFactory(eh.event, sliResults, err))
}

func resetSLIResultsInCaseOfError(err error, eventData GetSLITriggeredAdapterInterface, sliResults []*keptnv2.SLIResult) []*keptnv2.SLIResult {
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
			sliResults = []*keptnv2.SLIResult{
				{
					Metric: indicatorName,
					Value:  0.0,
				},
			}
		}
	}

	errMessage := err.Error()
	for _, sliResult := range sliResults {
		sliResult.Success = false
		sliResult.Message = errMessage
	}

	return sliResults
}

func (eh *GetSLIEventHandler) sendEvent(factory adapter.CloudEventFactoryInterface) error {
	err := eh.kClient.SendCloudEvent(factory)
	if err != nil {
		log.WithError(err).Error("Could not send get sli cloud event")
		return err
	}

	return nil
}

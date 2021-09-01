package sli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"strings"
	"time"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/event"

	"gopkg.in/yaml.v2"

	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	// configutils "github.com/keptn/go-utils/pkg/configuration-service/utils"
	// keptnevents "github.com/keptn/go-utils/pkg/events"
	// keptnutils "github.com/keptn/go-utils/pkg/utils"
	// v1 "k8s.io/client-go/kubernetes/typed/core/// "
)

const ProblemOpenSLI = "problem_open"

type GetSLIEventHandler struct {
	event      *GetSLITriggeredAdapter
	dtClient   *dynatrace.Client
	secretName string
	dashboard  string
}

func NewGetSLITriggeredHandler(event *GetSLITriggeredAdapter, dtClient *dynatrace.Client, secretName string, dashboard string) GetSLIEventHandler {
	return GetSLIEventHandler{
		event:      event,
		dtClient:   dtClient,
		secretName: secretName,
		dashboard:  dashboard,
	}
}

func (eh GetSLIEventHandler) HandleEvent() error {
	// prepare event

	// do not continue if SLIProvider is not dynatrace
	if eh.event.IsNotForDynatrace() {
		return nil
	}

	go eh.retrieveMetrics()

	return nil
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

/**
 * Adds an SLO Entry to the SLO.yaml
 */
func addSLO(keptnEvent adapter.EventContentAdapter, newSLO *keptncommon.SLO) error {

	// this is the default SLO in case none has yet been uploaded
	dashboardSLO := &keptncommon.ServiceLevelObjectives{
		Objectives: []*keptncommon.SLO{},
		TotalScore: &keptncommon.SLOScore{Pass: "90%", Warning: "75%"},
		Comparison: &keptncommon.SLOComparison{CompareWith: "single_result", IncludeResultWithScore: "pass", NumberOfComparisonResults: 1, AggregateFunction: "avg"},
	}

	// first - lets load the SLO.yaml from the config repo
	sloContent, err := common.GetKeptnResource(keptnEvent, common.KeptnSLOFilename)
	if err == nil && sloContent != "" {
		err := json.Unmarshal([]byte(sloContent), dashboardSLO)
		if err != nil {
			return fmt.Errorf("Couldnt parse existing SLO.yaml: %v", err)
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

	// and now we save it back to Keptn
	yamlAsByteArray, err := yaml.Marshal(dashboardSLO)
	if err != nil {
		return err
	}

	err = common.UploadKeptnResource(yamlAsByteArray, common.KeptnSLOFilename, keptnEvent)
	if err != nil {
		return fmt.Errorf("could not store %s : %v", common.KeptnSLOFilename, err)
	}

	return nil
}

/**
 * Tries to find a dynatrace dashboard that matches our project. If so - returns the SLI, SLO and SLIResults
 */
func getDataFromDynatraceDashboard(dynatraceHandler *Handler, keptnEvent adapter.EventContentAdapter, startUnix time.Time, endUnix time.Time, dashboardConfig string) (*DashboardLink, []*keptnv2.SLIResult, error) {

	//
	// Option 1: We query the data from a dashboard instead of the uploaded SLI.yaml
	// ==============================================================================
	// Lets see if we have a Dashboard in Dynatrace that we should parse
	result, err := dynatraceHandler.QueryDynatraceDashboardForSLIs(keptnEvent, dashboardConfig, startUnix, endUnix)
	if result == nil && err == nil {
		return nil, nil, nil
	}

	if err != nil {
		return nil, nil, fmt.Errorf("could not query Dynatrace dashboard for SLIs: %v", err)
	}

	// lets store the dashboard as well
	if result.Dashboard() != nil {
		jsonAsByteArray, err := json.MarshalIndent(result.Dashboard(), "", "  ")
		if err != nil {
			return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not convert dashboard to JSON: %s", err)
		}
		err = common.UploadKeptnResource(jsonAsByteArray, common.DynatraceDashboardFilename, keptnEvent)
		if err != nil {
			return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not store %s : %v", common.DynatraceDashboardFilename, err)
		}
	}

	// lets write the SLI to the config repo
	if result.SLI() != nil {
		yamlAsByteArray, err := yaml.Marshal(result.SLI())
		if err != nil {
			return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not convert dashboardSLI to JSON: %s", err)
		}

		err = common.UploadKeptnResource(yamlAsByteArray, common.DynatraceSLIFilename, keptnEvent)
		if err != nil {
			return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not store %s : %v", common.DynatraceSLIFilename, err)
		}
	}

	// lets write the SLO to the config repo
	if result.SLO() != nil {
		yamlAsByteArray, err := yaml.Marshal(result.SLO())
		if err != nil {
			return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not convert dashboardSLO to JSON: %s", err)
		}
		err = common.UploadKeptnResource(yamlAsByteArray, common.KeptnSLOFilename, keptnEvent)
		if err != nil {
			return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not store %s : %v", common.KeptnSLOFilename, err)
		}
	}

	// lets also write the result to a local file in local test mode
	if result.SLIResults() != nil {
		if common.RunLocal || common.RunLocalTest {
			log.Info("(RunLocal Output) Write SLIResult to sliresult.json")
			jsonAsByteArray, err := json.MarshalIndent(result.SLIResults(), "", "  ")
			if err != nil {
				return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not convert sliResults to JSON: %s", err)
			}

			err = common.UploadKeptnResource(jsonAsByteArray, common.KeptnSLIResultFilename, keptnEvent)
			if err != nil {
				return result.DashboardLink(), result.SLIResults(), fmt.Errorf("could not store %s : %v", common.KeptnSLIResultFilename, err)
			}
		}
	}

	return result.DashboardLink(), result.SLIResults(), nil
}

/**
 * getDynatraceProblemContext
 *
 * Will evaluate the event and - if it finds a dynatrace problem ID - will return this - otherwise it will return 0
 */
func getDynatraceProblemContext(eventData *GetSLITriggeredAdapter) string {

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

// retrieveMetrics Handles keptn.InternalGetSLIEventType
//
// First tries to find a Dynatrace dashboard and then parses it for SLIs and SLOs
// Second will go to parse the SLI.yaml and returns the SLI as passed in by the event
func (eh *GetSLIEventHandler) retrieveMetrics() error {
	// send get-sli.started event
	if err := sendGetSLIStartedEvent(eh.event); err != nil {
		return sendGetSLIFinishedEvent(eh.event, nil, err)
	}

	log.WithFields(
		log.Fields{
			"project": eh.event.GetProject(),
			"stage":   eh.event.GetStage(),
			"service": eh.event.GetService(),
		}).Info("Processing sh.keptn.internal.event.get-sli")

	// Adding DtCreds as a label so users know which DtCreds was used
	eh.event.addLabel("DtCreds", eh.secretName)

	//
	// creating Dynatrace Handler which allows us to call the Dynatrace API
	dynatraceHandler := NewDynatraceHandler(eh.event, eh.dtClient)

	//
	// parse start and end (which are datetime strings) and convert them into unix timestamps
	startUnix, endUnix, err := ensureRightTimestamps(eh.event.GetSLIStart(), eh.event.GetSLIEnd())
	if err != nil {
		log.WithError(err).Error("ensureRightTimestamps failed")
		return sendGetSLIFinishedEvent(eh.event, nil, err)
	}

	//
	// THIS IS OUR RETURN OBJECT: sliResult
	// Whether option 1 or option 2 - this will hold our SLIResults
	var sliResults []*keptnv2.SLIResult

	//
	// Option 1 - see if we can get the data from a Dynatrace Dashboard
	dashboardLinkAsLabel, sliResults, err := getDataFromDynatraceDashboard(dynatraceHandler, eh.event, startUnix, endUnix, eh.dashboard)
	if err != nil {
		// log the error, but continue with loading sli.yaml
		log.WithError(err).Error("getDataFromDynatraceDashboard failed")
	}

	// add link to dynatrace dashboard to labels
	if dashboardLinkAsLabel != nil {
		eh.event.addLabel("Dashboard Link", dashboardLinkAsLabel.String())
	}

	//
	// Option 2: If we have not received any data via a Dynatrace Dashboard lets query the SLIs based on the SLI.yaml definition
	if sliResults == nil {
		// get custom metrics for project if they exist
		projectCustomQueries := common.GetCustomQueries(eh.event)

		// query all indicators
		for _, indicator := range eh.event.GetIndicators() {
			if strings.Compare(indicator, ProblemOpenSLI) == 0 {
				log.WithField("indicator", indicator).Info("Skipping indicator as it is handled later")
			} else {
				log.WithField("indicator", indicator).Info("Fetching indicator")
				sliValue, err := dynatraceHandler.GetSLIValue(indicator, startUnix, endUnix, projectCustomQueries)
				if err != nil {
					log.WithError(err).Error("GetSLIValue failed")
					// failed to fetch metric
					sliResults = append(sliResults, &keptnv2.SLIResult{
						Metric:  indicator,
						Value:   0,
						Success: false, // Mark as failure
						Message: err.Error(),
					})
				} else {
					// successfully fetched metric
					sliResults = append(sliResults, &keptnv2.SLIResult{
						Metric:  indicator,
						Value:   sliValue,
						Success: true, // mark as success
					})
				}
			}
		}

		if common.RunLocal || common.RunLocalTest {
			log.WithField("sliResults", sliResults).Print("(RunLocal Output) sliResults")
			return nil
		}
	}

	//
	// ARE WE CALLED IN CONTEXT OF A PROBLEM REMEDIATION??
	// If so - we should try to query the status of the Dynatrace Problem that triggered this evaluation
	problemID := getDynatraceProblemContext(eh.event)
	if problemID != "" {
		problemIndicator := ProblemOpenSLI
		openProblemValue := 0.0
		success := false
		message := ""

		// lets query the status of this problem and add it to the SLI Result
		dynatraceProblem, err := dynatraceHandler.ExecuteGetDynatraceProblemById(problemID)
		if err != nil {
			message = err.Error()
		}

		if dynatraceProblem != nil {
			success = true
			if dynatraceProblem.Status == "OPEN" {
				openProblemValue = 1.0
			}
		}

		// lets add this to the sliResults
		sliResults = append(sliResults, &keptnv2.SLIResult{
			Metric:  problemIndicator,
			Value:   openProblemValue,
			Success: success,
			Message: message,
		})

		// lets add this to the SLO in case this indicator is not yet in SLO.yaml. Becuase if it doesnt get added the lighthouse wont evaluate the SLI values
		// we default it to open_problems<=0
		sloString := fmt.Sprintf("sli=%s;pass=<=0;key=true", problemIndicator)
		sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(sloString)

		errAddSlo := addSLO(eh.event, sloDefinition)
		if errAddSlo != nil {
			// TODO 2021-08-10: should this be added to the error object for sendGetSLIFinishedEvent below?
			log.WithError(errAddSlo).Error("problem while adding SLOs")
		}
	}

	// now - lets see if we have captured any result values - if not - return send an error
	err = nil
	if sliResults == nil {
		err = errors.New("Couldn't retrieve any SLI Results")
	}

	log.Info("Finished fetching metrics; Sending SLIDone event now ...")

	return sendGetSLIFinishedEvent(eh.event, sliResults, err)
}

/**
 * Sends the SLI Done Event. If err != nil it will send an error message
 */
func sendGetSLIFinishedEvent(eventData *GetSLITriggeredAdapter, indicatorValues []*keptnv2.SLIResult, err error) error {

	// if an error was set - the indicators will be set to failed and error message is set to each
	indicatorValues = resetIndicatorsInCaseOfError(err, eventData, indicatorValues)

	getSLIEvent := keptnv2.GetSLIFinishedEventData{
		EventData: keptnv2.EventData{
			Project: eventData.GetProject(),
			Stage:   eventData.GetStage(),
			Service: eventData.GetService(),
			Labels:  eventData.GetLabels(),
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
		},

		GetSLI: keptnv2.GetSLIFinished{
			IndicatorValues: indicatorValues,
			Start:           eventData.GetSLIStart(),
			End:             eventData.GetSLIEnd(),
		},
	}

	ev := cloudevents.NewEvent()
	ev.SetType(keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName))
	ev.SetSource(event.GetEventSource())
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetExtension("shkeptncontext", eventData.GetShKeptnContext())
	ev.SetExtension("triggeredid", eventData.GetEventID())
	ev.SetData(cloudevents.ApplicationJSON, getSLIEvent)

	return sendEvent(ev)
}

func resetIndicatorsInCaseOfError(err error, eventData *GetSLITriggeredAdapter, indicatorValues []*keptnv2.SLIResult) []*keptnv2.SLIResult {
	if err != nil {
		indicators := eventData.GetIndicators()
		if (indicatorValues == nil) || (len(indicatorValues) == 0) {
			if indicators == nil || len(indicators) == 0 {
				indicators = []string{"no metric"}
			}

			for _, indicatorName := range indicators {
				indicatorValues = []*keptnv2.SLIResult{
					{
						Metric: indicatorName,
						Value:  0.0,
					},
				}
			}
		}

		errMessage := err.Error()
		for _, indicator := range indicatorValues {
			indicator.Success = false
			indicator.Message = errMessage
		}
	}

	return indicatorValues
}

func sendGetSLIStartedEvent(eventData *GetSLITriggeredAdapter) error {

	getSLIStartedEvent := keptnv2.GetSLIStartedEventData{
		EventData: keptnv2.EventData{
			Project: eventData.GetProject(),
			Stage:   eventData.GetStage(),
			Service: eventData.GetService(),
			Labels:  eventData.GetLabels(),
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
		},
	}

	ev := cloudevents.NewEvent()
	ev.SetType(keptnv2.GetStartedEventType(keptnv2.GetSLITaskName))
	ev.SetSource(event.GetEventSource())
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetExtension("shkeptncontext", eventData.GetShKeptnContext())
	ev.SetExtension("triggeredid", eventData.GetEventID())
	ev.SetData(cloudevents.ApplicationJSON, getSLIStartedEvent)

	return sendEvent(ev)
}

/**
 * sends cloud event back to keptn
 */
func sendEvent(event cloudevents.Event) error {

	keptnHandler, err := keptnv2.NewKeptn(&event, keptn.KeptnOpts{})
	if err != nil {
		return err
	}

	return keptnHandler.SendCloudEvent(event)
}

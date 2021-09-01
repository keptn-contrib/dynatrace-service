package sli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
)

const Throughput = "throughput"
const ErrorRate = "error_rate"
const ResponseTimeP50 = "response_time_p50"
const ResponseTimeP90 = "response_time_p90"
const ResponseTimeP95 = "response_time_p95"

// store url to the metrics api format migration document
const MetricsAPIOldFormatNewFormatDoc = "https://github.com/keptn-contrib/dynatrace-sli-service/blob/master/docs/CustomQueryFormatMigration.md"

// Handler interacts with a dynatrace API endpoint
type Handler struct {
	KeptnEvent GetSLITriggeredAdapterInterface
	client     *dynatrace.Client
}

// NewDynatraceHandler returns a new dynatrace handler that interacts with the Dynatrace REST API
func NewDynatraceHandler(keptnEvent GetSLITriggeredAdapterInterface, client *dynatrace.Client) *Handler {
	return &Handler{
		KeptnEvent: keptnEvent,
		client:     client,
	}
}

// isValidUUID Helper function to validate whether string is a valid UUID in version 4, variant 1
func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89aAbB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

// findDynatraceDashboard Queries all Dynatrace Dashboards and returns the dashboard ID that matches the following name pattern:
//   KQG;project=%project%;service=%service%;stage=%stage%;xxx
// It returns the UUID of the dashboard that was found. If no dashboard was found it returns ""
func (ph *Handler) findDynatraceDashboard(keptnEvent adapter.EventContentAdapter) (string, error) {
	// Lets query the list of all Dashboards and find the one that matches project, stage, service based on the title (in the future - we can do it via tags)
	// create dashboard query URL and set additional headers
	body, err := ph.client.Get("/api/config/v1/dashboards")
	if err != nil {
		return "", err
	}

	// parse json
	dashboards := &DynatraceDashboards{}
	err = json.Unmarshal(body, &dashboards)
	if err != nil {
		return "", err
	}

	return dashboards.SearchForDashboardMatching(keptnEvent), nil
}

// loadDynatraceDashboard Depending on the dashboard parameter which is pulled from dynatrace.conf.yaml:dashboard this method either
//   - query:        queries all dashboards on the Dynatrace Tenant and returns the one that matches project/service/stage, or
//   - dashboard-ID: if this is a valid dashboard ID it will query the dashboard with this ID, e.g: ddb6a571-4bda-4e8b-a9c0-4a3e02c2e14a, or
//   - <empty>:      it will not query any dashboard.
// It returns a parsed Dynatrace Dashboard and the actual dashboard ID in case we queried a dashboard.
func (ph *Handler) loadDynatraceDashboard(keptnEvent adapter.EventContentAdapter, dashboard string) (*DynatraceDashboard, string, error) {

	// Option 1: there is no dashboard we should query
	if dashboard == "" {
		return nil, dashboard, nil
	}

	// Option 2: Query dashboards
	if dashboard == common.DynatraceConfigDashboardQUERY {
		var err error
		dashboard, err = ph.findDynatraceDashboard(keptnEvent)
		if dashboard == "" || err != nil {
			log.WithError(err).WithFields(
				log.Fields{
					"project": keptnEvent.GetProject(),
					"stage":   keptnEvent.GetStage(),
					"service": keptnEvent.GetService(),
				}).Debug("Dashboard option query but couldnt find KQG dashboard")

			// TODO 2021-08-03: should this really return no error, if querying dashboards found no match?
			// this would be the same result as option 1 then
			return nil, dashboard, nil
		}

		log.WithFields(
			log.Fields{
				"project":   keptnEvent.GetProject(),
				"stage":     keptnEvent.GetStage(),
				"service":   keptnEvent.GetService(),
				"dashboard": dashboard,
			}).Debug("Dashboard option query found for dashboard")
	}

	// Lets validate if we have a valid UUID - either because it was passed or because queried
	// If not - we are going down the dashboard route!
	if !isValidUUID(dashboard) {
		return nil, dashboard, fmt.Errorf("Dashboard ID %s not a valid UUID", dashboard)
	}

	// We have a valid Dashboard UUID - now lets query it!
	log.WithField("dashboard", dashboard).Debug("Query dashboard")
	body, err := ph.client.Get("/api/config/v1/dashboards/" + dashboard)
	if err != nil {
		return nil, dashboard, err
	}

	// parse json
	dynatraceDashboard := &DynatraceDashboard{}
	err = json.Unmarshal(body, &dynatraceDashboard)
	if err != nil {
		return nil, dashboard, fmt.Errorf("could not decode response payload: %v", err)
	}

	return dynatraceDashboard, dashboard, nil
}

// executeGetDynatraceSLO Calls the /slo/{sloId} API call to retrieve the values of the Dynatrace SLO for that timeframe
// It returns a DynatraceSLOResult object on success, an error otherwise
func (ph *Handler) executeGetDynatraceSLO(sloID string, startUnix time.Time, endUnix time.Time) (*DynatraceSLOResult, error) {
	body, err := ph.client.Get(
		fmt.Sprintf("/api/v2/slo/%s?from=%s&to=%s",
			sloID,
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix)))
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DynatraceSLOResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// for SLO - its also possible that there is an HTTP 200 but there is an error text in the error property!
	// Since Sprint 206 the error property is always there - but - will have the value "NONE" in case there is no actual error retrieving the value
	if result.Error != "NONE" {
		return nil, fmt.Errorf("Dynatrace API returned an error: %s", result.Error)
	}

	return &result, nil
}

// executeGetDynatraceProblems Calls the /problems/ API call to retrieve the the list of problems for that timeframe
// It returns a DynatraceProblemQueryResult object on success, an error otherwise
func (ph *Handler) executeGetDynatraceProblems(problemQuery string, startUnix time.Time, endUnix time.Time) (*DynatraceProblemQueryResult, error) {
	body, err := ph.client.Get(
		fmt.Sprintf("/api/v2/problems?from=%s&to=%s&%s",
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix),
			problemQuery))
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DynatraceProblemQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// executeGetDynatraceSecurityProblems Calls the /securityProblems/ API call to retrieve the list of security problems for that timeframe.
// It returns a DynatraceSecurityProblemQueryResult object on success, an error otherwise.
func (ph *Handler) executeGetDynatraceSecurityProblems(problemQuery string, startUnix time.Time, endUnix time.Time) (*DynatraceSecurityProblemQueryResult, error) {
	body, err := ph.client.Get(
		fmt.Sprintf("/api/v2/securityProblems?from=%s&to=%s&%s",
			common.TimestampToString(startUnix),
			common.TimestampToString(endUnix),
			problemQuery))
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DynatraceSecurityProblemQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// executeMetricAPIDescribe Calls the /metrics/<metricID> API call to retrieve Metric Definition Details.
func (ph *Handler) executeMetricAPIDescribe(metricId string) (*MetricDefinition, error) {
	body, err := ph.client.Get("/api/v2/metrics/" + metricId)
	if err != nil {
		return nil, err
	}

	// parse response json if we have a 200
	var result MetricDefinition
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// executeMetricsAPIQuery executes the passed Metrics API Call, validates that the call returns data and returns the data set
func (ph *Handler) executeMetricsAPIQuery(metricsQuery string) (*DynatraceMetricsQueryResult, error) {
	path := "/api/v2/metrics/query?" + metricsQuery
	log.WithField("query", ph.client.DynatraceCreds.Tenant+path).Debug("Final Query")

	body, err := ph.client.Get(path)
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DynatraceMetricsQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Result) == 0 {
		// datapoints is empty - try again?
		return nil, errors.New("Dynatrace Metrics API returned no DataPoints")
	}

	return &result, nil
}

// ExecuteGetDynatraceProblemById Calls the /problems/<problemId> API call to retrieve Problem Details
func (ph *Handler) ExecuteGetDynatraceProblemById(problemId string) (*DynatraceProblem, error) {
	body, err := ph.client.Get("/api/v2/problems/" + problemId)
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DynatraceProblem
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// executeGetDynatraceUSQLQuery executes the passed Metrics API Call, validates that the call returns data and returns the data set
func (ph *Handler) executeGetDynatraceUSQLQuery(usql string) (*DTUSQLResult, error) {
	path := "/api/v1/userSessionQueryLanguage/table?" + usql
	log.WithField("query", ph.client.DynatraceCreds.Tenant+path).Debug("Final USQL Query")

	body, err := ph.client.Get(path)
	if err != nil {
		return nil, err
	}

	// parse response json
	var result DTUSQLResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// if no data comes back
	if len(result.Values) == 0 {
		// datapoints is empty - try again?
		return nil, errors.New("Dynatrace USQL Query didnt return any DataPoints")
	}

	return &result, nil
}

// buildDynatraceUSQLQuery builds a USQL query based on the incoming values
func (ph *Handler) buildDynatraceUSQLQuery(query string, startUnix time.Time, endUnix time.Time) string {
	log.WithField("query", query).Debug("Finalize USQL query")

	// replace query params (e.g., $PROJECT, $STAGE, $SERVICE ...)
	// default query params that are required: resolution, from and to
	q := make(url.Values)
	q.Add("query", ph.replaceQueryParameters(query))
	q.Add("explain", "false")
	q.Add("addDeepLinkFields", "false")
	q.Add("startTimestamp", common.TimestampToString(startUnix))
	q.Add("endTimestamp", common.TimestampToString(endUnix))

	return q.Encode()
}

// buildDynatraceMetricsQuery builds the complete query string based on start, end and filters
// metricQuery should contain metricSelector and entitySelector
// Returns:
//  #1: Dynatrace API metric query string
//  #2: Metric selector that this query will return, e.g: builtin:host.cpu
//  #3: error
func (ph *Handler) buildDynatraceMetricsQuery(metricQuery string, startUnix time.Time, endUnix time.Time) (string, string, error) {
	// replace query params (e.g., $PROJECT, $STAGE, $SERVICE ...)
	metricQuery = ph.replaceQueryParameters(metricQuery)

	if strings.HasPrefix(metricQuery, "?metricSelector=") {
		log.WithFields(
			log.Fields{
				"query":        metricQuery,
				"helpDocument": MetricsAPIOldFormatNewFormatDoc,
			}).Debug("COMPATIBILITY WARNING: query string is not compatible. Auto-removing the ? in front.")
		metricQuery = strings.Replace(metricQuery, "?metricSelector=", "metricSelector=", 1)
	}

	// split query string by first occurrence of "?"
	querySplit := strings.Split(metricQuery, "?")
	metricSelector := ""
	metricQueryParams := ""

	// support the old format with "metricSelector:someFilters()?scope=..." as well as the new format with
	// "?metricSelector=metricSelector&entitySelector=...&scope=..."
	if len(querySplit) == 1 {
		// new format without "?" -> everything within the query string are query parameters
		metricQueryParams = querySplit[0]
	} else {
		log.WithFields(
			log.Fields{
				"query":        metricQueryParams,
				"helpDocument": MetricsAPIOldFormatNewFormatDoc,
			}).Debug("COMPATIBILITY WARNING: query uses the old format")
		// old format with "?" - everything left of the ? is the identifier, everything right are query params
		metricSelector = querySplit[0]

		// build the new query
		metricQueryParams = fmt.Sprintf("metricSelector=%s&%s", querySplit[0], querySplit[1])
	}

	q, err := url.ParseQuery(metricQueryParams)
	if err != nil {
		return "", "", fmt.Errorf("could not parse metrics URL: %s", err.Error())
	}

	q.Add("resolution", "Inf") // resolution=Inf means that we only get 1 datapoint (per service)
	q.Add("from", common.TimestampToString(startUnix))
	q.Add("to", common.TimestampToString(endUnix))

	// check if q contains "scope"
	scopeData := q.Get("scope")

	// compatibility with old scope=... custom queries
	if scopeData != "" {
		log.WithField("helpDocument", MetricsAPIOldFormatNewFormatDoc).Debug("COMPATIBILITY WARNING: querying the new metrics API requires use of entitySelector rather than scope")
		// scope is no longer supported in the new API, it needs to be called "entitySelector" and contain type(SERVICE)
		if !strings.Contains(scopeData, "type(SERVICE)") {
			log.WithField("helpDocument", MetricsAPIOldFormatNewFormatDoc).Debug("COMPATIBILITY WARNING: Automatically adding type(SERVICE) to entitySelector for compatibility with the new Metrics API")
			scopeData = fmt.Sprintf("%s,type(SERVICE)", scopeData)
		}
		// add scope as entitySelector
		q.Add("entitySelector", scopeData)
	}

	// check metricSelector
	if metricSelector == "" {
		metricSelector = q.Get("metricSelector")
	}

	return q.Encode(), metricSelector, nil
}

/**
 * When passing a query to dynatrace using filter expressions - the dimension names in a filter will be escaped with specifal characters, e.g: filter(dt.entity.browser,IE) becomes filter(dt~entity~browser,ie)
 * This function here tries to come up with a better matching algorithm
 * WHILE NOT PERFECT - HERE IS THE FIRST IMPLEMENTATION
 */
func isMatchingMetricID(singleResultMetricID string, queryMetricID string) bool {
	if strings.Compare(singleResultMetricID, queryMetricID) == 0 {
		return true
	}

	// lets do some basic fuzzy matching
	if strings.Contains(singleResultMetricID, "~") {
		log.WithFields(
			log.Fields{
				"singleResultMetricID": singleResultMetricID,
				"queryMetricID":        queryMetricID,
			}).Debug("Need fuzzy matching")

		//
		// lets just see whether everything until the first : matches
		if strings.Contains(singleResultMetricID, ":") {
			log.Debug("Just compare before first")

			fuzzyResultMetricID := strings.Split(singleResultMetricID, ":")[0]
			fuzzyQueryMetricID := strings.Split(queryMetricID, ":")[0]
			if strings.Compare(fuzzyResultMetricID, fuzzyQueryMetricID) == 0 {
				log.Debug("FUZZY MATCH")
				return true
			}
		}

		// TODO - more fuzzy checks
	}

	return false
}

// getEntitySelectorFromEntityFilter Parses the filtersPerEntityType dashboard definition and returns the entitySelector query filter -
// the return value always starts with a , (comma)
//   return example: ,entityId("ABAD-222121321321")
func getEntitySelectorFromEntityFilter(filtersPerEntityType map[string]map[string][]string, entityType string) string {
	entityTileFilter := ""
	if filtersPerEntityType, containsEntityType := filtersPerEntityType[entityType]; containsEntityType {
		// Check for SPECIFIC_ENTITIES - if we have an array then we filter for each entity
		if entityArray, containsSpecificEntities := filtersPerEntityType["SPECIFIC_ENTITIES"]; containsSpecificEntities {
			for _, entityId := range entityArray {
				entityTileFilter = entityTileFilter + ","
				entityTileFilter = entityTileFilter + fmt.Sprintf("entityId(\"%s\")", entityId)
			}
		}
		// Check for SPECIFIC_ENTITIES - if we have an array then we filter for each entity
		if tagArray, containsAutoTags := filtersPerEntityType["AUTO_TAGS"]; containsAutoTags {
			for _, tag := range tagArray {
				entityTileFilter = entityTileFilter + ","
				entityTileFilter = entityTileFilter + fmt.Sprintf("tag(\"%s\")", tag)
			}
		}
	}
	return entityTileFilter
}

// processSLOTile Processes an SLO Tile and queries the data from the Dynatrace API.
// If successful returns sliResult, sliIndicatorName, sliQuery & sloDefinition
func (ph *Handler) processSLOTile(sloID string, startUnix time.Time, endUnix time.Time) (*keptnv2.SLIResult, string, string, *keptncommon.SLO, error) {

	// Step 1: Query the Dynatrace API to get the actual value for this sloID
	sloResult, err := ph.executeGetDynatraceSLO(sloID, startUnix, endUnix)
	if err != nil {
		return nil, "", "", nil, err
	}

	// Step 2: As we have the SLO Result including SLO Definition we add it to the SLI & SLO objects
	// IndicatorName is based on the slo Name
	// the value defaults to the E
	indicatorName := common.CleanIndicatorName(sloResult.Name)
	value := sloResult.EvaluatedPercentage
	sliResult := &keptnv2.SLIResult{
		Metric:  indicatorName,
		Value:   value,
		Success: true,
	}

	log.WithFields(
		log.Fields{
			"indicatorName": indicatorName,
			"value":         value,
		}).Debug("Adding SLO to sloResult")

	// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
	// we prepend this with SLO;<SLO-ID>
	sliQuery := fmt.Sprintf("SLO;%s", sloID)

	// lets add the SLO definition in case we need to generate an SLO.yaml
	// we normally parse these values from the tile name. In this case we just build that tile name -> maybe in the future we will allow users to add additional SLO defs via the Tile Name, e.g: weight or KeySli

	// Please see https://github.com/keptn-contrib/dynatrace-sli-service/issues/97 - for more information on that change of Dynatrace SLO API
	// if we still run against an old API we fall back to the old fields
	warning := sloResult.Warning
	if warning <= 0.0 {
		warning = sloResult.TargetWarningOLD
	}
	target := sloResult.Target
	if target <= 0.0 {
		target = sloResult.TargetSuccessOLD
	}
	sloString := fmt.Sprintf("sli=%s;pass=>=%f;warning=>=%f", indicatorName, warning, target)
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(sloString)

	return sliResult, indicatorName, sliQuery, sloDefinition, nil
}

// processOpenProblemTile Processes an Open Problem Tile and queries the number of open problems. The current default is that there is a pass criteria of <= 0 as we dont allow problems
// If successful returns sliResult, sliIndicatorName, sliQuery & sloDefinition
func (ph *Handler) processOpenProblemTile(problemSelector string, startUnix time.Time, endUnix time.Time) (*keptnv2.SLIResult, string, string, *keptncommon.SLO, error) {

	problemQuery := ""
	if problemSelector != "" {
		problemQuery = fmt.Sprintf("problemSelector=%s", problemSelector)
	}

	// Step 1: Query the Dynatrace API to get the number of actual problems matching that query and timeframe
	problemQueryResult, err := ph.executeGetDynatraceProblems(problemQuery, startUnix, endUnix)
	if err != nil {
		return nil, "", "", nil, err
	}

	// Step 2: As we have the SLO Result including SLO Definition we add it to the SLI & SLO objects
	// IndicatorName is based on the slo Name
	// the value defaults to the E
	indicatorName := "problems"
	value := float64(problemQueryResult.TotalCount)
	sliResult := &keptnv2.SLIResult{
		Metric:  indicatorName,
		Value:   value,
		Success: true,
	}

	log.WithFields(
		log.Fields{
			"indicatorName": indicatorName,
			"value":         value,
		}).Debug("Adding SLO to sloResult")

	// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
	// we prepend this with PV2;entitySelector=asdaf&problemSelector=asdf
	sliQuery := fmt.Sprintf("PV2;%s", problemQuery)

	// lets add the SLO definitin in case we need to generate an SLO.yaml
	// we normally parse these values from the tile name. In this case we just build that tile name -> maybe in the future we will allow users to add additional SLO defs via the Tile Name, e.g: weight or KeySli
	sloString := fmt.Sprintf("sli=%s;pass=<=0;key=true", indicatorName)
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(sloString)

	return sliResult, indicatorName, sliQuery, sloDefinition, nil
}

// processOpenSecurityProblemTile Processes an Open Problem Tile and queries the number of open problems. The current default is that there is a pass criteria of <= 0 as we dont allow problems
// If successful returns sliResult, sliIndicatorName, sliQuery & sloDefinition
func (ph *Handler) processOpenSecurityProblemTile(securityProblemSelector string, startUnix time.Time, endUnix time.Time) (*keptnv2.SLIResult, string, string, *keptncommon.SLO, error) {

	problemQuery := ""
	if securityProblemSelector != "" {
		problemQuery = fmt.Sprintf("securityProblemSelector=%s", securityProblemSelector)
	}

	// Step 1: Query the Dynatrace API to get the number of actual problems matching that query and timeframe
	problemQueryResult, err := ph.executeGetDynatraceSecurityProblems(problemQuery, startUnix, endUnix)
	if err != nil {
		return nil, "", "", nil, err
	}

	// Step 2: As we have the SLO Result including SLO Definition we add it to the SLI & SLO objects
	// IndicatorName is based on the slo Name
	// the value defaults to the E
	indicatorName := "security_problems"
	value := float64(problemQueryResult.TotalCount)
	sliResult := &keptnv2.SLIResult{
		Metric:  indicatorName,
		Value:   value,
		Success: true,
	}

	log.WithFields(
		log.Fields{
			"indicatorName": indicatorName,
			"value":         value,
		}).Debug("Adding SLO to sloResult")

	// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
	// we prepend this with SECPV2;entitySelector=asdaf&problemSelector=asdf
	sliQuery := fmt.Sprintf("SECPV2;%s", problemQuery)

	// lets add the SLO definitin in case we need to generate an SLO.yaml
	// we normally parse these values from the tile name. In this case we just build that tile name -> maybe in the future we will allow users to add additional SLO defs via the Tile Name, e.g: weight or KeySli
	sloString := fmt.Sprintf("sli=%s;pass=<=0;key=true", indicatorName)
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(sloString)

	return sliResult, indicatorName, sliQuery, sloDefinition, nil
}

// Looks at the DataExplorerQuery configuration of a data explorer chart and generates the Metrics Query.
//
// Returns a MetricQueryComponents object
//   - metricId, e.g: built-in:mymetric
//   - metricUnit, e.g: MilliSeconds
//   - metricQuery, e.g: metricSelector=metric&filter...
//   - fullMetricQuery, e.g: metricQuery&from=123213&to=2323
//   - entitySelectorSLIDefinition, e.g: ,entityid(FILTERDIMENSIONVALUE)
//   - filterSLIDefinitionAggregator, e.g: , filter(eq(Test Step,FILTERDIMENSIONVALUE))
func (ph *Handler) generateMetricQueryFromDataExplorer(dataQuery DataExplorerQuery, tileManagementZoneFilter *ManagementZoneFilter, startUnix time.Time, endUnix time.Time) (*MetricQueryComponents, error) {

	// TODO 2021-08-04: there are too many return values and they are have the same type

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := ph.executeMetricAPIDescribe(dataQuery.Metric)
	if err != nil {
		log.WithError(err).WithField("metric", dataQuery.Metric).Debug("Error retrieving metric description")
		return nil, err
	}

	// building the merge aggregator string, e.g: merge(1):merge(0) - or merge(0)
	metricDimensionCount := len(metricDefinition.DimensionDefinitions)
	metricAggregation := metricDefinition.DefaultAggregation.Type
	mergeAggregator := ""
	filterAggregator := ""
	filterSLIDefinitionAggregator := ""
	entitySelectorSLIDefinition := ""
	entityFilter := ""

	// we need to merge all those dimensions based on the metric definition that are not included in the "splitBy"
	// so - we iterate through the dimensions based on the metric definition from the back to front - and then merge those not included in splitBy
	for metricDimIx := metricDimensionCount - 1; metricDimIx >= 0; metricDimIx-- {
		log.WithField("metricDimIx", metricDimIx).Debug("Processing Dimension Ix")

		doMergeDimension := true
		for _, splitDimension := range dataQuery.SplitBy {
			log.WithFields(
				log.Fields{
					"dimension1": splitDimension,
					"dimension2": metricDefinition.DimensionDefinitions[metricDimIx].Key,
				}).Debug("Comparing Dimensions %")

			if strings.Compare(splitDimension, metricDefinition.DimensionDefinitions[metricDimIx].Key) == 0 {
				doMergeDimension = false
			}
		}

		if doMergeDimension {
			// this is a dimension we want to merge as it is not split by in the chart
			log.WithField("dimension", metricDefinition.DimensionDefinitions[metricDimIx].Key).Debug("merging dimension")
			mergeAggregator = mergeAggregator + fmt.Sprintf(":merge(%d)", metricDimIx)
		}
	}

	// Create the right entity Selectors for the queries execute
	// TODO: we currently only support a single filter - if we want to support more we need to build this in
	if dataQuery.FilterBy != nil && len(dataQuery.FilterBy.NestedFilters) > 0 {

		if len(dataQuery.FilterBy.NestedFilters[0].Criteria) == 1 {
			if strings.HasPrefix(dataQuery.FilterBy.NestedFilters[0].Filter, "dt.entity.") {
				entitySelectorSLIDefinition = ",entityId(FILTERDIMENSIONVALUE)"
				entityFilter = fmt.Sprintf("&entitySelector=entityId(%s)", dataQuery.FilterBy.NestedFilters[0].Criteria[0].Value)
			} else {
				filterSLIDefinitionAggregator = fmt.Sprintf(":filter(eq(%s,FILTERDIMENSIONVALUE))", dataQuery.FilterBy.NestedFilters[0].Filter)
				filterAggregator = fmt.Sprintf(":filter(%s(%s,%s))", dataQuery.FilterBy.NestedFilters[0].Criteria[0].Evaluator, dataQuery.FilterBy.NestedFilters[0].Filter, dataQuery.FilterBy.NestedFilters[0].Criteria[0].Value)
			}
		} else {
			log.Debug("Code only supports a single filter for data explorer")
		}
	}

	// TODO: we currently only support one split dimension
	// but - if we split by a dimension we need to include that dimension in our individual SLI query definitions - thats why we hand this back in the filter clause
	if dataQuery.SplitBy != nil {
		if len(dataQuery.SplitBy) == 1 {
			filterSLIDefinitionAggregator = fmt.Sprintf("%s:filter(eq(%s,FILTERDIMENSIONVALUE))", filterSLIDefinitionAggregator, dataQuery.SplitBy[0])
		} else {
			log.Debug("Code only supports a single splitby dimension for data explorer")
		}
	}

	// lets create the metricSelector and entitySelector
	// ATTENTION: adding :names so we also get the names of the dimensions and not just the entities. This means we get two values for each dimension
	metricQuery := fmt.Sprintf("metricSelector=%s%s%s:%s:names%s%s",
		dataQuery.Metric, mergeAggregator, filterAggregator, strings.ToLower(metricAggregation),
		entityFilter, tileManagementZoneFilter.ForEntitySelector())

	// lets build the Dynatrace API Metric query for the proposed timeframe and additonal filters!
	fullMetricQuery, metricID, err := ph.buildDynatraceMetricsQuery(metricQuery, startUnix, endUnix)
	if err != nil {
		return nil, err
	}

	return &MetricQueryComponents{
		metricID:                      metricID,
		metricUnit:                    metricDefinition.Unit,
		metricQuery:                   metricQuery,
		fullMetricQueryString:         fullMetricQuery,
		entitySelectorSLIDefinition:   entitySelectorSLIDefinition,
		filterSLIDefinitionAggregator: filterSLIDefinitionAggregator,
	}, nil

}

// Looks at the ChartSeries configuration of a regular chart and generates the Metrics Query
//
// Returns a MetricQueryComponents object
//   - metricId, e.g: built-in:mymetric
//   - metricUnit, e.g: MilliSeconds
//   - metricQuery, e.g: metricSelector=metric&filter...
//   - fullMetricQuery, e.g: metricQuery&from=123213&to=2323
//   - entitySelectorSLIDefinition, e.g: ,entityid(FILTERDIMENSIONVALUE)
//   - filterSLIDefinitionAggregator, e.g: , filter(eq(Test Step,FILTERDIMENSIONVALUE))
func (ph *Handler) generateMetricQueryFromChart(series ChartSeries, tileManagementZoneFilter *ManagementZoneFilter, filtersPerEntityType map[string]map[string][]string, startUnix time.Time, endUnix time.Time) (*MetricQueryComponents, error) {

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := ph.executeMetricAPIDescribe(series.Metric)
	if err != nil {
		log.WithError(err).WithField("metric", series.Metric).Debug("Error retrieving metric description")
		return nil, err
	}

	// building the merge aggregator string, e.g: merge(1):merge(0) - or merge(0)
	metricDimensionCount := len(metricDefinition.DimensionDefinitions)
	metricAggregation := metricDefinition.DefaultAggregation.Type
	mergeAggregator := ""
	filterAggregator := ""
	filterSLIDefinitionAggregator := ""
	entitySelectorSLIDefinition := ""

	// now we need to merge all the dimensions that are not part of the series.dimensions, e.g: if the metric has two dimensions but only one dimension is used in the chart we need to merge the others
	// as multiple-merges are possible but as they are executed in sequence we have to use the right index
	for metricDimIx := metricDimensionCount - 1; metricDimIx >= 0; metricDimIx-- {
		doMergeDimension := true
		metricDimIxAsString := strconv.Itoa(metricDimIx)
		// lets check if this dimension is in the chart
		for _, seriesDim := range series.Dimensions {
			log.WithFields(
				log.Fields{
					"seriesDim.id": seriesDim.ID,
					"metricDimIx":  metricDimIxAsString,
				}).Debug("check")
			if strings.Compare(seriesDim.ID, metricDimIxAsString) == 0 {
				// this is a dimension we want to keep and not merge
				log.WithField("dimension", metricDefinition.DimensionDefinitions[metricDimIx].Name).Debug("not merging dimension")
				doMergeDimension = false

				// lets check if we need to apply a dimension filter
				// TODO: support multiple filters - right now we only support 1
				if len(seriesDim.Values) > 0 {
					filterAggregator = fmt.Sprintf(":filter(eq(%s,%s))", seriesDim.Name, seriesDim.Values[0])
				} else {
					// we need this for the generation of the SLI for each individual dimension value
					// if the dimension is a dt.entity we have to add an addiotnal entityId to the entitySelector - otherwise we add a filter for the dimension
					if strings.HasPrefix(seriesDim.Name, "dt.entity.") {
						entitySelectorSLIDefinition = fmt.Sprintf(",entityId(FILTERDIMENSIONVALUE)")
					} else {
						filterSLIDefinitionAggregator = fmt.Sprintf(":filter(eq(%s,FILTERDIMENSIONVALUE))", seriesDim.Name)
					}
				}
			}
		}

		if doMergeDimension {
			// this is a dimension we want to merge as it is not split by in the chart
			log.WithField("dimension", metricDefinition.DimensionDefinitions[metricDimIx].Name).Debug("merging dimension")
			mergeAggregator = mergeAggregator + fmt.Sprintf(":merge(%d)", metricDimIx)
		}
	}

	// handle aggregation. If "NONE" is specified we go to the defaultAggregration
	if series.Aggregation != "NONE" {
		metricAggregation = series.Aggregation
	}
	// for percentile we need to specify the percentile itself
	if metricAggregation == "PERCENTILE" {
		metricAggregation = fmt.Sprintf("%s(%f)", metricAggregation, series.Percentile)
	}
	// for rate measures such as failure rate we take average if it is "OF_INTEREST_RATIO"
	if metricAggregation == "OF_INTEREST_RATIO" {
		metricAggregation = "avg"
	}
	// for rate measures charting also provides the "OTHER_RATIO" option which is the inverse
	// TODO: not supported via API - so we default to avg
	if metricAggregation == "OTHER_RATIO" {
		metricAggregation = "avg"
	}

	// TODO - handle aggregation rates -> probably doesnt make sense as we always evalute a short timeframe
	// if series.AggregationRate

	// lets get the true entity type as the one in the dashboard might not be accurate, e.g: IOT might be used instead of CUSTOM_DEVICE
	// so - if the metric definition has EntityTypes defined we take the first one
	entityType := series.EntityType
	if len(metricDefinition.EntityType) > 0 {
		entityType = metricDefinition.EntityType[0]
	}

	// Need to implement chart filters per entity type, e.g: its possible that a chart has a filter on entites or tags
	// lets see if we have a FiltersPerEntityType for the tiles EntityType
	entityTileFilter := getEntitySelectorFromEntityFilter(filtersPerEntityType, entityType)

	// lets create the metricSelector and entitySelector
	// ATTENTION: adding :names so we also get the names of the dimensions and not just the entities. This means we get two values for each dimension
	metricQuery := fmt.Sprintf("metricSelector=%s%s%s:%s:names&entitySelector=type(%s)%s%s",
		series.Metric, mergeAggregator, filterAggregator, strings.ToLower(metricAggregation),
		entityType, entityTileFilter, tileManagementZoneFilter.ForEntitySelector())

	// lets build the Dynatrace API Metric query for the proposed timeframe and additional filters!
	fullMetricQuery, metricID, err := ph.buildDynatraceMetricsQuery(metricQuery, startUnix, endUnix)
	if err != nil {
		return nil, err
	}

	return &MetricQueryComponents{
		metricID:                      metricID,
		metricUnit:                    metricDefinition.Unit,
		metricQuery:                   metricQuery,
		fullMetricQueryString:         fullMetricQuery,
		entitySelectorSLIDefinition:   entitySelectorSLIDefinition,
		filterSLIDefinitionAggregator: filterSLIDefinitionAggregator,
	}, nil
}

// Generates the relevant SLIs & SLO definitions based on the metric query
// noOfDimensionsInChart: how many dimensions did we have in the chart definition
func (ph *Handler) generateSLISLOFromMetricsAPIQuery(noOfDimensionsInChart int, sloDefinition *keptncommon.SLO, metricQueryComponents *MetricQueryComponents, dashboardSLI *SLI, dashboardSLO *keptncommon.ServiceLevelObjectives) []*keptnv2.SLIResult {

	// TODO 2021-08-04: there are too many parameters and many of them have the same type

	var sliResults []*keptnv2.SLIResult

	// Lets run the Query and iterate through all data per dimension. Each Dimension will become its own indicator
	queryResult, err := ph.executeMetricsAPIQuery(metricQueryComponents.fullMetricQueryString)
	if err != nil {
		log.WithError(err).Debug("No result for query")

		// ERROR-CASE: Metric API return no values or an error
		// we could not query data - so - we return the error back as part of our SLIResults
		sliResults = append(sliResults, &keptnv2.SLIResult{
			Metric:  sloDefinition.SLI,
			Value:   0,
			Success: false, // Mark as failure
			Message: err.Error(),
		})

		// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
		dashboardSLI.Indicators[sloDefinition.SLI] = metricQueryComponents.metricQuery

		return sliResults
	}

	// SUCCESS-CASE: we retrieved values - now we iterate through the results and create an indicator result for every dimension
	for _, singleResult := range queryResult.Result {
		log.WithFields(
			log.Fields{
				"metricId":                      singleResult.MetricID,
				"filterSLIDefinitionAggregator": metricQueryComponents.filterSLIDefinitionAggregator,
				"entitySelectorSLIDefinition":   metricQueryComponents.entitySelectorSLIDefinition,
			}).Debug("Processing result")
		if !isMatchingMetricID(singleResult.MetricID, metricQueryComponents.metricID) {
			log.WithFields(
				log.Fields{
					"wantedMetricId": metricQueryComponents.metricID,
					"gotMetricId":    singleResult.MetricID,
				}).Debug("Retrieving unintended metric")

			continue
		}

		dataResultCount := len(singleResult.Data)
		if dataResultCount == 0 {
			log.Debug("No data for metric")
		}
		for _, singleDataEntry := range singleResult.Data {
			//
			// we need to generate the indicator name based on the base name + all dimensions, e.g: teststep_MYTESTSTEP, teststep_MYOTHERTESTSTEP
			// EXCEPTION: If there is only ONE data value then we skip this and just use the base SLI name
			indicatorName := sloDefinition.SLI

			metricQueryForSLI := metricQueryComponents.metricQuery

			// we need this one to "fake" the MetricQuery for the SLi.yaml to include the dynamic dimension name for each value
			// we initialize it with ":names" as this is the part of the metric query string we will replace
			filterSLIDefinitionAggregatorValue := ":names"

			if dataResultCount > 1 {
				// because we use the ":names" transformation we always get two dimension entries for entity dimensions, e.g: Host, Service .... First is the Name of the entity, then the ID of the Entity
				// lets first validate that we really received Dimension Names
				dimensionCount := len(singleDataEntry.Dimensions)
				dimensionIncrement := 2
				if dimensionCount != (noOfDimensionsInChart * 2) {
					// ph.Logger.Debug(fmt.Sprintf("DIDNT RECEIVE ID and Names. Lets assume we just received the dimension IDs"))
					dimensionIncrement = 1
				}

				// lets iterate through the list and get all names
				for dimIx := 0; dimIx < len(singleDataEntry.Dimensions); dimIx = dimIx + dimensionIncrement {
					dimensionValue := singleDataEntry.Dimensions[dimIx]
					indicatorName = indicatorName + "_" + dimensionValue

					filterSLIDefinitionAggregatorValue = ":names" + strings.Replace(metricQueryComponents.filterSLIDefinitionAggregator, "FILTERDIMENSIONVALUE", dimensionValue, 1)

					if metricQueryComponents.entitySelectorSLIDefinition != "" && dimensionIncrement == 2 {
						dimensionEntityID := singleDataEntry.Dimensions[dimIx+1]
						metricQueryForSLI = metricQueryForSLI + strings.Replace(metricQueryComponents.entitySelectorSLIDefinition, "FILTERDIMENSIONVALUE", dimensionEntityID, 1)
					}
				}
			}

			// make sure we have a valid indicator name by getting rid of special characters
			indicatorName = common.CleanIndicatorName(indicatorName)

			// calculating the value
			value := 0.0
			for _, singleValue := range singleDataEntry.Values {
				value = value + singleValue
			}
			value = value / float64(len(singleDataEntry.Values))

			// lets scale the metric
			value = scaleData(metricQueryComponents.metricID, metricQueryComponents.metricUnit, value)

			// we got our metric, slos and the value

			log.WithFields(
				log.Fields{
					"name":  indicatorName,
					"value": value,
				}).Debug("Got indicator value")

			// lets add the value to our SLIResult array
			sliResults = append(sliResults, &keptnv2.SLIResult{
				Metric:  indicatorName,
				Value:   value,
				Success: true,
			})

			// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
			// we use ":names" to find the right spot to add our custom dimension filter
			// we also "pre-pend" the metricDefinition.Unit - which allows us later on to do the scaling right
			dashboardSLI.Indicators[indicatorName] = fmt.Sprintf("MV2;%s;%s", metricQueryComponents.metricUnit, strings.Replace(metricQueryForSLI, ":names", filterSLIDefinitionAggregatorValue, 1))

			// lets add the SLO definition in case we need to generate an SLO.yaml
			dashboardSLO.Objectives = append(
				dashboardSLO.Objectives,
				&keptncommon.SLO{
					SLI:     indicatorName,
					Weight:  sloDefinition.Weight,
					KeySLI:  sloDefinition.KeySLI,
					Pass:    sloDefinition.Pass,
					Warning: sloDefinition.Warning,
				})
		}
	}

	return sliResults
}

// QueryDynatraceDashboardForSLIs implements - https://github.com/keptn-contrib/dynatrace-sli-service/issues/60
// Queries Dynatrace for the existance of a dashboard tagged with keptn_project:project, keptn_stage:stage, keptn_service:service, SLI
// if this dashboard exists it will be parsed and a custom SLI_dashboard.yaml and an SLO_dashboard.yaml will be created
// Returns:
//  #1: Link to Dashboard
//  #2: SLI
//  #3: ServiceLevelObjectives
//  #4: SLIResult
//  #5: Error
func (ph *Handler) QueryDynatraceDashboardForSLIs(keptnEvent adapter.EventContentAdapter, dashboard string, startUnix time.Time, endUnix time.Time) (*DashboardQueryResult, error) {

	// Lets see if there is a dashboard.json already in the configuration repo - if so its an indicator that we should query the dashboard
	// This check is especially important for backward compatibility as the new dynatrace.conf.yaml:dashboard property is changing the default behavior
	// If a dashboard.json exists and dashboard property is empty we default to QUERY - which is the old default behavior
	existingDashboardContent, err := common.GetKeptnResource(keptnEvent, common.DynatraceDashboardFilename)
	if err == nil && existingDashboardContent != "" && dashboard == "" {
		log.Debug("Set dashboard=query for backward compatibility as dashboard.json was present!")
		dashboard = common.DynatraceConfigDashboardQUERY
	}

	// lets load the dashboard if needed
	dashboardJSON, dashboard, err := ph.loadDynatraceDashboard(keptnEvent, dashboard)
	if err != nil {
		return nil, fmt.Errorf("Error while processing dashboard config '%s' - %v", dashboard, err)
	}

	if dashboardJSON == nil {
		return nil, nil
	}

	// lets also generate the dashboard link for that timeframe (gtf=c_START_END) as well as management zone (gf=MZID) to pass back as label to Keptn
	dashboardLinkAsLabel := NewDashboardLink(ph.client.DynatraceCreds.Tenant, startUnix, endUnix, dashboardJSON.ID, dashboardJSON.DashboardMetadata.DashboardFilter)

	// Lets validate if we really need to process this dashboard as it might be the same (without change) from the previous runs
	// see https://github.com/keptn-contrib/dynatrace-sli-service/issues/92 for more details
	if dashboardJSON.isTheSameAs(existingDashboardContent) {
		log.Debug("Dashboard hasn't changed: skipping parsing of dashboard")
		return NewDashboardQueryResultFrom(dashboardLinkAsLabel), nil
	}

	// generate our own SLIResult array based on the dashboard configuration
	result := &DashboardQueryResult{
		dashboardLink: dashboardLinkAsLabel,
		dashboard:     dashboardJSON,
		sli: &SLI{
			SpecVersion: "0.1.4",
			Indicators:  make(map[string]string),
		},
		slo: &keptncommon.ServiceLevelObjectives{
			Objectives: []*keptncommon.SLO{},
			TotalScore: &keptncommon.SLOScore{
				Pass:    "90%",
				Warning: "75%",
			},
			Comparison: &keptncommon.SLOComparison{
				CompareWith:               "single_result",
				IncludeResultWithScore:    "pass",
				NumberOfComparisonResults: 1,
				AggregateFunction:         "avg",
			},
		},
		sliResults: []*keptnv2.SLIResult{},
	}

	log.Debug("Dashboard has changed: reparsing it!")

	// now lets iterate through the dashboard to find our SLIs
	for _, tile := range dashboardJSON.Tiles {
		switch tile.TileType {
		case "MARKDOWN":
			ph.addSLIAndSLOToResultFromMarkdownTile(&tile, result)
		case "SLO":
			ph.addSLIAndSLOToResultFromSLOTile(&tile, startUnix, endUnix, result)
		case "OPEN_PROBLEMS":
			ph.addSLIAndSLOToResultFromOpenProblemsTile(&tile, startUnix, endUnix, result)
			// current logic also does security tile processing for open problem tiles
			ph.addSLIAndSLOToResultFromOpenSecurityProblemsTile(&tile, startUnix, endUnix, result)
		case "DATA_EXPLORER":
			// here we handle the new Metric Data Explorer Tile
			ph.addSLIAndSLOToResultFromDataExplorerTile(&tile, startUnix, endUnix, result)
		case "CUSTOM_CHARTING":
			ph.addSLIAndSLOToResultFromCustomChartsTile(&tile, startUnix, endUnix, result)
		case "DTAQL":
			ph.addSLIAndSLOToResultFromUserSessionQueryTile(&tile, startUnix, endUnix, result)
		default:
			// we do not do markdowns (HEADER) or synthetic tests (SYNTHETIC_TESTS)
			continue
		}
	}

	return result, nil
}

func (ph *Handler) addSLIAndSLOToResultFromMarkdownTile(tile *Tile, result *DashboardQueryResult) {
	// we allow the user to use a markdown to specify SLI/SLO properties, e.g: KQG.Total.Pass
	// if we find KQG. we process the markdown
	if strings.Contains(tile.Markdown, "KQG.") {
		common.ParseMarkdownConfiguration(tile.Markdown, result.slo)
	}
}

func (ph *Handler) addSLIAndSLOToResultFromSLOTile(tile *Tile, startUnix time.Time, endUnix time.Time, result *DashboardQueryResult) {
	// we will take the SLO definition from Dynatrace
	for _, sloEntity := range tile.AssignedEntities {
		log.WithField("sloEntity", sloEntity).Debug("Processing SLO Definition")

		sliResult, sliIndicator, sliQuery, sloDefinition, err := ph.processSLOTile(sloEntity, startUnix, endUnix)
		if err != nil {
			log.WithError(err).Error("Error Processing SLO")
			continue
		}

		result.sliResults = append(result.sliResults, sliResult)
		result.sli.Indicators[sliIndicator] = sliQuery
		result.slo.Objectives = append(result.slo.Objectives, sloDefinition)
	}
}

func (ph *Handler) addSLIAndSLOToResultFromOpenProblemsTile(tile *Tile, startUnix time.Time, endUnix time.Time, result *DashboardQueryResult) {
	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(
		result.dashboard.DashboardMetadata.DashboardFilter,
		tile.TileFilter.ManagementZone)

	// we will query the number of open problems based on the specification of that tile
	problemSelector := "status(open)" + tileManagementZoneFilter.ForProblemSelector()

	sliResult, sliIndicator, sliQuery, sloDefinition, err := ph.processOpenProblemTile(problemSelector, startUnix, endUnix)
	if err != nil {
		log.WithError(err).Error("Error Processing OPEN_PROBLEMS")
		return
	}

	result.sliResults = append(result.sliResults, sliResult)
	result.sli.Indicators[sliIndicator] = sliQuery
	result.slo.Objectives = append(result.slo.Objectives, sloDefinition)
}

func (ph *Handler) addSLIAndSLOToResultFromOpenSecurityProblemsTile(tile *Tile, startUnix time.Time, endUnix time.Time, result *DashboardQueryResult) {
	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(
		result.dashboard.DashboardMetadata.DashboardFilter,
		tile.TileFilter.ManagementZone)

	// we will query the number of open security problems based on the specification of that tile
	problemSelector := "status(OPEN)" + tileManagementZoneFilter.ForProblemSelector()

	sliResult, sliIndicator, sliQuery, sloDefinition, err := ph.processOpenSecurityProblemTile(problemSelector, startUnix, endUnix)
	if err != nil {
		log.WithError(err).Error("Error Processing OPEN_SECURITY_PROBLEMS")
		return
	}

	result.sliResults = append(result.sliResults, sliResult)
	result.sli.Indicators[sliIndicator] = sliQuery
	result.slo.Objectives = append(result.slo.Objectives, sloDefinition)
}

func (ph *Handler) addSLIAndSLOToResultFromDataExplorerTile(tile *Tile, startUnix time.Time, endUnix time.Time, result *DashboardQueryResult) {
	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(
		result.dashboard.DashboardMetadata.DashboardFilter,
		tile.TileFilter.ManagementZone)

	// first - lets figure out if this tile should be included in SLI validation or not - we parse the title and look for "sli=sliname"
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(tile.Name)
	if sloDefinition.SLI == "" {
		log.WithField("tileName", tile.Name).Debug("Data explorer tile not included as name doesnt include sli=SLINAME")
		return
	}

	// now lets process that tile - lets run through each query
	for _, dataQuery := range tile.Queries {
		log.WithField("metric", dataQuery.Metric).Debug("Processing data explorer query")

		// First lets generate the query and extract all important metric information we need for generating SLIs & SLOs
		metricQuery, err := ph.generateMetricQueryFromDataExplorer(dataQuery, tileManagementZoneFilter, startUnix, endUnix)

		// if there was no error we generate the SLO & SLO definition
		if err != nil {
			log.WithError(err).Warn("generateMetricQueryFromDataExplorer returned an error, SLI will not be used")
			continue
		}

		newSliResults := ph.generateSLISLOFromMetricsAPIQuery(len(dataQuery.SplitBy), sloDefinition, metricQuery, result.sli, result.slo)
		result.sliResults = append(result.sliResults, newSliResults...)
	}
}

func (ph *Handler) addSLIAndSLOToResultFromCustomChartsTile(tile *Tile, startUnix time.Time, endUnix time.Time, result *DashboardQueryResult) {
	tileTitle := tile.Title()

	// first - lets figure out if this tile should be included in SLI validation or not - we parse the title and look for "sli=sliname"
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(tileTitle)
	if sloDefinition.SLI == "" {
		log.WithField("tileTitle", tileTitle).Debug("Tile not included as name doesnt include sli=SLINAME")
		return
	}

	log.WithFields(
		log.Fields{
			"tileTitle":         tileTitle,
			"baseIndicatorName": sloDefinition.SLI,
		}).Debug("Processing custom chart")

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(
		result.dashboard.DashboardMetadata.DashboardFilter,
		tile.TileFilter.ManagementZone)

	// we can potentially have multiple series on that chart
	for _, series := range tile.FilterConfig.ChartConfig.Series {

		// First lets generate the query and extract all important metric information we need for generating SLIs & SLOs
		metricQuery, err := ph.generateMetricQueryFromChart(series, tileManagementZoneFilter, tile.FilterConfig.FiltersPerEntityType, startUnix, endUnix)

		// if there was no error we generate the SLO & SLO definition
		if err != nil {
			log.WithError(err).Warn("generateMetricQueryFromChart returned an error, SLI will not be used")
			continue
		}

		newSliResults := ph.generateSLISLOFromMetricsAPIQuery(len(series.Dimensions), sloDefinition, metricQuery, result.sli, result.slo)
		result.sliResults = append(result.sliResults, newSliResults...)
	}
}

func (ph *Handler) addSLIAndSLOToResultFromUserSessionQueryTile(tile *Tile, startUnix time.Time, endUnix time.Time, result *DashboardQueryResult) {
	// for Dynatrace Query Language we currently support the following
	// SINGLE_VALUE: we just take the one value that comes back
	// PIE_CHART, COLUMN_CHART: we assume the first column is the dimension and the second column is the value column
	// TABLE: we assume the first column is the dimension and the last is the value

	usql := ph.buildDynatraceUSQLQuery(tile.Query, startUnix, endUnix)
	usqlResult, err := ph.executeGetDynatraceUSQLQuery(usql)
	if err != nil {
		log.WithError(err).Warn("executeGetDynatraceUSQLQuery returned an error")
		return
	}

	tileTitle := tile.Title()

	// first - lets figure out if this tile should be included in SLI validation or not - we parse the title and look for "sli=sliname"
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(tileTitle)
	if sloDefinition.SLI == "" {
		log.WithField("tileTitle", tileTitle).Debug("Tile not included as name doesnt include sli=SLINAME")
		return
	}

	for _, rowValue := range usqlResult.Values {
		dimensionName := ""
		dimensionValue := 0.0

		switch tile.Type {
		case "SINGLE_VALUE":
			dimensionValue = rowValue[0].(float64)
		case "PIE_CHART":
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		case "COLUMN_CHART":
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		case "TABLE":
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[len(rowValue)-1].(float64)
		default:
			log.WithField("tileType", tile.Type).Debug("Unsupport USQL tile type")
			continue
		}

		// lets scale the metric
		// value = scaleData(metricDefinition.MetricID, metricDefinition.Unit, value)

		// we got our metric, slos and the value
		indicatorName := sloDefinition.SLI
		if dimensionName != "" {
			indicatorName = indicatorName + "_" + dimensionName
		}

		log.WithFields(
			log.Fields{
				"name":           indicatorName,
				"dimensionValue": dimensionValue,
			}).Debug("Appending SLIResult")

		// lets add the value to our SLIResult array
		result.sliResults = append(result.sliResults, &keptnv2.SLIResult{
			Metric:  indicatorName,
			Value:   dimensionValue,
			Success: true,
		})

		// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
		// in that case we also need to mask it with USQL, TITLE_TYPE, DIMENSIONNAME
		result.sli.Indicators[indicatorName] = fmt.Sprintf("USQL;%s;%s;%s", tile.Type, dimensionName, tile.Query)

		// lets add the SLO definition in case we need to generate an SLO.yaml
		result.slo.Objectives = append(
			result.slo.Objectives,
			&keptncommon.SLO{
				SLI:     indicatorName,
				Weight:  sloDefinition.Weight,
				KeySLI:  sloDefinition.KeySLI,
				Pass:    sloDefinition.Pass,
				Warning: sloDefinition.Warning,
			})
	}
}

// GetSLIValue queries a single metric value from Dynatrace API.
// Can handle both Metric Queries as well as USQL
func (ph *Handler) GetSLIValue(name string, startUnix time.Time, endUnix time.Time, customQueries map[string]string) (float64, error) {

	// first we get the query from the SLI configuration based on its logical name
	query, err := ph.getSLIQuery(name, customQueries)
	if err != nil {
		return 0, fmt.Errorf("Error when fetching SLI config for %s %s.", name, err.Error())
	}

	log.WithFields(
		log.Fields{
			"name":  name,
			"query": query,
		}).Debug("Retrieved SLI query")

	switch {
	case strings.HasPrefix(query, "USQL;"):
		return ph.executeUSQLQuery(query, startUnix, endUnix)
	case strings.HasPrefix(query, "SLO;"):
		return ph.executeSLOQuery(query, startUnix, endUnix)
	case strings.HasPrefix(query, "PV2;"):
		return ph.executeProblemQuery(query, startUnix, endUnix)
	case strings.HasPrefix(query, "SECPV2;"):
		return ph.executeSecurityProblemQuery(query, startUnix, endUnix)
	case strings.HasPrefix(query, "MV2;"):
		return ph.executeMetricsV2Query(query, startUnix, endUnix)
	default:
		return ph.executeMetricsQuery(query, "", startUnix, endUnix)
	}
}

// USQL query
func (ph *Handler) executeUSQLQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	// In this case we need to parse USQL;TILE_TYPE;DIMENSION;QUERY
	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 4 {
		return 0, fmt.Errorf("USQL Query incorrect format: %s", metricsQuery)
	}

	tileName := querySplits[1]
	requestedDimensionName := querySplits[2]
	usqlRawQuery := querySplits[3]

	usql := ph.buildDynatraceUSQLQuery(usqlRawQuery, startUnix, endUnix)
	usqlResult, err := ph.executeGetDynatraceUSQLQuery(usql)

	if err != nil {
		return 0, fmt.Errorf("Error executing USQL Query %v", err)
	}

	for _, rowValue := range usqlResult.Values {
		dimensionName := ""
		dimensionValue := 0.0

		if tileName == "SINGLE_VALUE" {
			dimensionValue = rowValue[0].(float64)
		} else if tileName == "PIE_CHART" {
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		} else if tileName == "COLUMN_CHART" {
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		} else if tileName == "TABLE" {
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[len(rowValue)-1].(float64)
		} else {
			log.WithField("tileName", tileName).Debug("Unsupported USQL Tile Type")
			continue
		}

		// did we find the value we were looking for?
		if strings.Compare(dimensionName, requestedDimensionName) == 0 {
			return dimensionValue, nil
		}
	}

	return 0, fmt.Errorf("Error executing USQL Query")
}

// query a specific SLO
func (ph *Handler) executeSLOQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {

	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 2 {
		return 0, fmt.Errorf("SLO Indicator query has wrong format. Should be SLO;<SLID> but is: %s", metricsQuery)
	}

	sloID := querySplits[1]
	sloResult, err := ph.executeGetDynatraceSLO(sloID, startUnix, endUnix)
	if err != nil {
		return 0, fmt.Errorf("Error executing SLO Dynatrace Query %v", err)
	}

	return sloResult.EvaluatedPercentage, nil
}

func (ph *Handler) executeProblemQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	// we query number of problems
	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 2 {
		return 0, fmt.Errorf("Problemv2 Indicator query has wrong format. Should be PV2;entitySelectory=selector&problemSelector=selector but is: %s", metricsQuery)
	}

	problemQuery := querySplits[1]
	problemQueryResult, err := ph.executeGetDynatraceProblems(problemQuery, startUnix, endUnix)
	if err != nil {
		return 0, fmt.Errorf("Error executing Dynatrace Problem v2 Query %v", err)
	}

	return float64(problemQueryResult.TotalCount), nil
}

//  query number of problems
func (ph *Handler) executeSecurityProblemQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {

	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 2 {
		return 0, fmt.Errorf("Security Problemv2 Indicator query has wrong format. Should be SECPV2;securityProblemSelector=selector but is: %s", metricsQuery)
	}

	problemQuery := querySplits[1]
	problemQueryResult, err := ph.executeGetDynatraceSecurityProblems(problemQuery, startUnix, endUnix)
	if err != nil {
		return 0, fmt.Errorf("Error executing Dynatrace Security Problem v2 Query %v", err)
	}

	return float64(problemQueryResult.TotalCount), nil
}

func (ph *Handler) executeMetricsV2Query(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {

	// lets first start to query for the MV2 prefix, e.g: MV2;byte;actualQuery
	// if it starts with MV2 we extract metric unit and the actual query

	metricsQuery = metricsQuery[4:]
	queryStartIndex := strings.Index(metricsQuery, ";")
	unit := metricsQuery[:queryStartIndex]
	metricsQuery = metricsQuery[queryStartIndex+1:]

	return ph.executeMetricsQuery(metricsQuery, unit, startUnix, endUnix)

}

func (ph *Handler) executeMetricsQuery(metricsQuery string, unit string, startUnix time.Time, endUnix time.Time) (float64, error) {

	metricsQuery, metricSelector, err := ph.buildDynatraceMetricsQuery(metricsQuery, startUnix, endUnix)
	if err != nil {
		return 0, err
	}
	result, err := ph.executeMetricsAPIQuery(metricsQuery)

	if err != nil {
		return 0, fmt.Errorf("Dynatrace Metrics API returned an error: %s. This was the query executed: %s", err.Error(), metricsQuery)
	}

	if result == nil {
		return 0, fmt.Errorf("Dynatrace Metrics API failed to return a result")
	}
	for _, i := range result.Result {

		if isMatchingMetricID(i.MetricID, metricSelector) {

			if len(i.Data) != 1 {
				jsonString, _ := json.Marshal(i)
				return 0, fmt.Errorf("Dynatrace Metrics API returned %d result values, expected 1 for query: %s.\nPlease ensure the response contains exactly one value (e.g., by using :merge(0):avg for the metric). Here is the output for troubleshooting: %s", len(i.Data), metricsQuery, string(jsonString))
			}

			return scaleData(metricSelector, unit, i.Data[0].Values[0]), nil
		}
	}

	return 0, fmt.Errorf("No result matched the query's metric selector")
}

// scaleData
// scales data based on the timeseries identifier (e.g., service.responsetime needs to be scaled from microseconds to milliseocnds)
// Right now this method scales microseconds to milliseconds and bytes to Kilobytes
// At a later stage we should extend this with more conversions and even think of allowing custom scale targets, e.g: Byte to MegaByte
func scaleData(metricID string, unit string, value float64) float64 {
	if (strings.Compare(unit, "MicroSecond") == 0) || strings.Contains(metricID, "builtin:service.response.time") {
		// scale from microseconds to milliseconds
		return value / 1000.0
	}

	// convert Bytes to Kilobyte
	if strings.Compare(unit, "Byte") == 0 {
		return value / 1024
	}

	return value
}

func (ph *Handler) replaceQueryParameters(query string) string {
	// apply customfilters
	for _, filter := range ph.KeptnEvent.GetCustomSLIFilters() {
		filter.Value = strings.Replace(filter.Value, "'", "", -1)
		filter.Value = strings.Replace(filter.Value, "\"", "", -1)

		// replace the key in both variants, "normal" and uppercased
		query = strings.Replace(query, "$"+filter.Key, filter.Value, -1)
		query = strings.Replace(query, "$"+strings.ToUpper(filter.Key), filter.Value, -1)
	}

	// apply default values
	/* query = strings.Replace(query, "$PROJECT", ph.Project, -1)
	query = strings.Replace(query, "$STAGE", ph.Stage, -1)
	query = strings.Replace(query, "$SERVICE", ph.Service, -1)
	query = strings.Replace(query, "$DEPLOYMENT", ph.Deployment, -1)*/

	query = common.ReplaceKeptnPlaceholders(query, ph.KeptnEvent)

	return query
}

// getSLIQuery get query associated with an SLI name
func (ph *Handler) getSLIQuery(name string, customQueries map[string]string) (string, error) {
	if customQueries != nil {
		if val, ok := customQueries[name]; ok {
			return val, nil
		}
	}

	log.WithField("name", name).Debug("No custom SLI found - Looking in defaults")

	// default SLI configs
	// Switched to new metric v2 query language as discussed here: https://github.com/keptn-contrib/dynatrace-sli-service/issues/91
	switch name {
	case Throughput:
		return "metricSelector=builtin:service.requestCount.total:merge(0):sum&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ErrorRate:
		return "metricSelector=builtin:service.errors.total.rate:merge(0):avg&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ResponseTimeP50:
		return "metricSelector=builtin:service.response.time:merge(0):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ResponseTimeP90:
		return "metricSelector=builtin:service.response.time:merge(0):percentile(90)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ResponseTimeP95:
		return "metricSelector=builtin:service.response.time:merge(0):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	default:
		return "", fmt.Errorf("Unsupported SLI %s", name)
	}
}
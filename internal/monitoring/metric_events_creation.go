package monitoring

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/query"
	keptnlib "github.com/keptn/go-utils/pkg/lib"

	log "github.com/sirupsen/logrus"
)

const keptnService = "keptn_service"
const keptnDeployment = "keptn_deployment"

type criteriaObject struct {
	Operator        string
	Value           float64
	CheckPercentage bool
	IsComparison    bool
	CheckIncrease   bool
}

type metricEventCreation struct {
	dtClient          dynatrace.ClientInterface
	eventSenderClient keptn.EventSenderClientInterface
	sliAndSLOReader   keptn.SLIAndSLOReaderInterface
}

func newMetricEventCreation(dynatraceClient dynatrace.ClientInterface, eventSenderClient keptn.EventSenderClientInterface, sliAndSLOReader keptn.SLIAndSLOReaderInterface) metricEventCreation {
	return metricEventCreation{
		dtClient:          dynatraceClient,
		eventSenderClient: eventSenderClient,
		sliAndSLOReader:   sliAndSLOReader,
	}
}

// create creates new metric events if SLOs are specified.
func (mec metricEventCreation) create(ctx context.Context, project string, stage string, service string) []configResult {
	log.Info("Creating custom metric events for project SLIs")
	slos, err := mec.sliAndSLOReader.GetSLOs(ctx, project, stage, service)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"service": service,
				"stage":   stage}).Info("No SLOs defined for service. Skipping creation of custom metric events.")
		return nil
	}

	// get custom metrics for project
	slis, err := mec.sliAndSLOReader.GetSLIs(ctx, project, stage, service)
	if err != nil {
		log.WithError(err).WithField("project", project).Error("Failed to get SLIs for project")
		return nil
	}
	projectCustomQueries := query.NewCustomQueries(slis)

	managementZones, err := dynatrace.NewManagementZonesClient(mec.dtClient).GetAll(ctx)
	var mzId int64 = -1
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"project": project, "stage": stage}).Error("Could not retrieve management zones")
		return nil
	}

	// TODO 2021-08-20: check the logic below - if parsing management zone id does not work, we will continue anyway?
	if zone, wasFound := managementZones.GetByName(GetManagementZoneNameForProjectAndStage(project, stage)); wasFound {
		mzId, err = strconv.ParseInt(zone.ID, 10, 64)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"project": project, "stage": stage}).Warn("Could not parse management zone ID")
		}
	} else {
		log.WithError(err).WithFields(log.Fields{"project": project, "stage": stage}).Warn("Could not find management zone")
		return nil
	}

	metricEventsClient := dynatrace.NewMetricEventsClient(mec.dtClient)
	var metricsEventResults []configResult
	// try to create metric events using best effort.
	for _, objective := range slos.Objectives {
		query, err := projectCustomQueries.GetQueryByNameOrDefault(objective.SLI)
		if err != nil {
			// Error occurred but continue
			log.WithField("sli", objective.SLI).Error("Could not find query for SLI")

			// if .GetQueryByNameOrDefault(...) would fail, it would return an empty string. So we would do all the work
			// in the two for loops below until we come to the point in createKeptnMetricEventDTO where it would be checked
			// whether 'config' not equals to "" and return an error, log error and continue for each iteration.
			// therefore, continue
			continue
		}

		metricsEventResults = append(
			metricsEventResults,
			setupAllMetricEvents(ctx, metricEventsClient, project, stage, service, objective, query, mzId)...)

	}

	if len(metricsEventResults) > 0 {
		log.WithField("metricEventsURL", mec.dtClient.Credentials().GetTenant()+"/#settings/anomalydetection/metricevents").Info("Custom metric events have been generated")
	}

	return metricsEventResults
}

func setupAllMetricEvents(ctx context.Context, client *dynatrace.MetricEventsClient, project string, stage string, service string, slo *keptnlib.SLO, query string, managementZoneID int64) []configResult {
	var metricEventsResults []configResult
	for _, criteria := range slo.Pass {
		for _, crit := range criteria.Criteria {

			metricEventsResult, err := setupSingleMetricEvent(ctx, client, project, stage, service, slo.SLI, query, crit, managementZoneID)
			if err != nil {
				continue
			}

			metricEventsResults = append(metricEventsResults, *metricEventsResult)
		}
	}

	return metricEventsResults
}

func setupSingleMetricEvent(ctx context.Context, client *dynatrace.MetricEventsClient, project string, stage string, service string, metric string, query string, crit string, managementZoneID int64) (*configResult, error) {
	// criteria.Criteria
	criteriaObject, err := parseCriteriaString(crit)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).WithField("criteria", crit).Error("Could not parse criteria")
		return nil, fmt.Errorf("could not parse criteria: %s", crit)
	}

	if criteriaObject.IsComparison {
		// comparison-based criteria cannot be mapped to alerts
		return nil, fmt.Errorf("comparison-based criteria cannot be mapped to alerts")
	}

	newMetricEvent, err := createKeptnMetricEventDTO(project, stage, service, metric, query, crit, criteriaObject.Value, managementZoneID)
	if err != nil {
		// Error occurred but continue
		log.WithError(err).WithFields(
			log.Fields{
				"sli":      metric,
				"criteria": crit,
			}).Error("Could not create metric event definition for criteria")
		return nil, fmt.Errorf("could not create metric event definition for criteria, sli: %s, criteria: %s", metric, crit)
	}

	err = createOrUpdateMetricEvent(ctx, client, newMetricEvent)
	if err != nil {
		log.WithError(err).WithField("metricName", newMetricEvent.Name).Error("Could not create metric event")
		return nil, fmt.Errorf("could not create metric event: %s", newMetricEvent.Name)
	}

	log.WithFields(log.Fields{"name": newMetricEvent.Name, "criteria": crit}).Info("Created metric event")
	return &configResult{
		Name:    newMetricEvent.Name,
		Success: true,
	}, nil
}

func createOrUpdateMetricEvent(ctx context.Context, client *dynatrace.MetricEventsClient, newMetricEvent *dynatrace.MetricEvent) error {
	existingMetricEvent, err := client.GetMetricEventByName(ctx, newMetricEvent.Name)
	if err != nil {
		return err
	}

	if existingMetricEvent != nil {
		// adapt all properties that have initially been defaulted to some value from previous (potentially modified event)
		existingMetricEvent.Threshold = newMetricEvent.Threshold
		existingMetricEvent.TagFilters = nil

		err := client.Update(ctx, existingMetricEvent)
		if err != nil {
			return err
		}

		return nil
	}

	err = client.Create(ctx, newMetricEvent)
	if err != nil {
		return err
	}

	return nil
}

func parseCriteriaString(criteria string) (*criteriaObject, error) {
	// example values: <+15%, <500, >-8%, =0
	// possible operators: <, <=, =, >, >=
	// regex: ^([<|<=|=|>|>=]{1,2})([+|-]{0,1}\\d*\.?\d*)([%]{0,1})
	regex := `^([<|<=|=|>|>=]{1,2})([+|-]{0,1}\d*\.?\d*)([%]{0,1})`
	var re *regexp.Regexp
	re = regexp.MustCompile(regex)

	// remove whitespaces
	criteria = strings.Replace(criteria, " ", "", -1)

	if !re.MatchString(criteria) {
		return nil, errors.New("invalid criteria string")
	}

	c := &criteriaObject{}

	if strings.HasSuffix(criteria, "%") {
		c.CheckPercentage = true
		criteria = strings.TrimSuffix(criteria, "%")
	}

	operators := []string{"<=", "<", "=", ">=", ">"}

	for _, operator := range operators {
		if strings.HasPrefix(criteria, operator) {
			c.Operator = operator
			criteria = strings.TrimPrefix(criteria, operator)
			break
		}
	}

	if strings.HasPrefix(criteria, "-") {
		c.IsComparison = true
		c.CheckIncrease = false
		criteria = strings.TrimPrefix(criteria, "-")
	} else if strings.HasPrefix(criteria, "+") {
		c.IsComparison = true
		c.CheckIncrease = true
		criteria = strings.TrimPrefix(criteria, "+")
	} else {
		c.IsComparison = false
		c.CheckIncrease = false
	}

	floatValue, err := strconv.ParseFloat(criteria, 64)
	if err != nil {
		return nil, errors.New("could not parse criteria target value")
	}
	c.Value = floatValue

	return c, nil
}

var supportedAggregations = [...]string{"avg", "max", "min", "count", "sum", "value", "percentile"}

func createKeptnMetricEventDTO(project string, stage string, service string, metric string, query string, condition string, threshold float64, managementZoneID int64) (*dynatrace.MetricEvent, error) {

	// TODO: 2021-09-20: Check what parts are still needed
	/*
		need to map queries used by SLI-service to metric event definition.
		example: builtin:service.response.time:merge("dt.entity.service"):percentile(90)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)

		1. split by '?' and get first part => builtin:service.response.time:merge("dt.entity.service"):percentile(90)
		2. split by ':' => builtin:service.response.time | merge("dt.entity.service") | percentile(90) => merge("dt.entity.service") is not needed
		3. first part is the metricId and can be used for the Metric Event API => builtin:service.response.time
		4. Aggregation is limited to: AVG, COUNT, MAX, MEDIAN, MIN, OF_INTEREST, OF_INTEREST_RATIO, OTHER, OTHER_RATIO, P90, SUM, VALUE
	*/

	if project == "" || stage == "" || service == "" || metric == "" || query == "" {
		return nil, errors.New("missing input parameter values")
	}

	query = strings.TrimPrefix(query, "metricSelector=")
	// 1. split by '?' and get first part => builtin:service.response.time:merge("dt.entity.service"):percentile(90)
	split := strings.Split(query, "?")

	// 2. split by ':' => builtin:service.response.time | merge("dt.entity.service") | percentile(90) => merge("dt.entity.service") is not needed/supported by MetricEvent API
	splittedQuery := strings.Split(split[0], ":")

	if len(splittedQuery) < 2 {
		return nil, errors.New("invalid metricId")
	}
	metricId := splittedQuery[0] + ":" + splittedQuery[1]
	meAggregation := ""
	for _, transformation := range splittedQuery {
		isSupportedAggregation := false
		for _, aggregationType := range supportedAggregations {
			if strings.Contains(strings.ToLower(transformation), aggregationType) {
				isSupportedAggregation = true
			}
		}

		if isSupportedAggregation {
			meAggregation = getMetricEventAggregation(transformation)

			/*
				if meAggregation == "" {
					return nil, errors.New("unsupported aggregation type: " + transformation)
				}

			*/
		}
	}
	/*
		if meAggregation == "" {
			return nil, errors.New("no aggregation provided in query")
		}
	*/

	meAlertCondition, err := parseAlertCondition(condition)
	if err != nil {
		return nil, err
	}

	metricEvent := &dynatrace.MetricEvent{
		Metadata:          dynatrace.MEMetadata{},
		MetricID:          metricId,
		Name:              metric + " (Keptn." + project + "." + stage + "." + service + ")",
		Description:       "Keptn SLI violated: The {metricname} value of {severity} was {alert_condition} your custom threshold of {threshold}.",
		EventType:         "CUSTOM_ALERT",
		Severity:          "CUSTOM_ALERT",
		AlertCondition:    meAlertCondition,
		Samples:           5, // taken from default value of custom metric events
		ViolatingSamples:  3, // taken from default value of custom metric events
		DealertingSamples: 5, // taken from default value of custom metric events
		Threshold:         threshold,
		Enabled:           false,
		TagFilters:        nil, // not used anymore by MetricEvents API, replaced by AlertingScope
		AlertingScope: []dynatrace.MEAlertingScope{
			// LIMITATION: currently only a maximum of 3 tag filters is supported
			{
				FilterType:       "MANAGEMENT_ZONE",
				ManagementZoneID: managementZoneID,
			},
			{
				FilterType: "TAG",
				TagFilter: &dynatrace.METagFilter{
					Context: "CONTEXTLESS",
					Key:     keptnService,
					Value:   service,
				},
			},
			{
				FilterType: "TAG",
				TagFilter: &dynatrace.METagFilter{
					Context: "CONTEXTLESS",
					Key:     keptnDeployment,
					Value:   "primary",
				},
			},
		},
	}

	// LIMITATION: currently we do not have the possibility of specifying units => assume MILLI_SECONDS for response time metrics
	if strings.Contains(metric, "time") {
		metricEvent.Unit = "MILLI_SECOND"
	}

	if meAggregation != "" {
		metricEvent.AggregationType = meAggregation
	}

	return metricEvent, nil
}

func parseAlertCondition(condition string) (string, error) {
	meAlertCondition := ""
	if strings.Contains(condition, "+") || strings.Contains(condition, "-") || strings.Contains(condition, "%") {
		return "", errors.New("unsupported condition. only fixed thresholds are supported")
	}

	if strings.Contains(condition, ">") {
		meAlertCondition = "BELOW"
	} else if strings.Contains(condition, "<") {
		meAlertCondition = "ABOVE"
	} else {
		return "", errors.New("unsupported condition. only fixed thresholds are supported")
	}
	return meAlertCondition, nil
}

func getMetricEventAggregation(metricAPIAgg string) string {
	// LIMITATION: currently, only single aggregations are supported, so, e.g. not (min,max)
	metricAPIAgg = strings.ToLower(metricAPIAgg)

	if strings.Contains(metricAPIAgg, "percentile") {
		// only MEDIAN and P90 are supported for MetricEvents
		// => if the percentile in the query is >= 90, use P90, otherwise assume MEDIAN
		if strings.Contains(metricAPIAgg, "(9") {
			return "P90"
		} else {
			return "MEDIAN"
		}
	}
	// due to incompatibilities between metrics and metric event API it's safer to not pass an aggregation in the MetricEvent definition in most cases
	// the Metric Event API will default it to an appropriate aggregation
	/*else if strings.Contains(metricAPIAgg, "min") {
		return "MIN"
	} else if strings.Contains(metricAPIAgg, "max") {
		return "MAX"
	} else if strings.Contains(metricAPIAgg, "count") {
		return "COUNT"
	} else if strings.Contains(metricAPIAgg, "sum") {
		return "SUM"
	} else if strings.Contains(metricAPIAgg, "value") {
		return "VALUE"
	} else if strings.Contains(metricAPIAgg, "avg") {
		return "AVG"
	}
	*/
	return ""
}

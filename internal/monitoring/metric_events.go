package monitoring

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"

	"github.com/keptn-contrib/dynatrace-service/internal/common"

	configutils "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

const sliResourceURI = "dynatrace/sli.yaml"
const throughput = "throughput"
const errorRate = "error_rate"
const responseTimeP50 = "response_time_p50"
const responseTimeP90 = "response_time_p90"
const responseTimeP95 = "response_time_p95"

const keptnService = "keptn_service"
const keptnDeployment = "keptn_deployment"

type MetricEventCreation struct {
	client *dynatrace.DynatraceHelper
}

func NewMetricEventCreation(client *dynatrace.DynatraceHelper) MetricEventCreation {
	return MetricEventCreation{
		client: client,
	}
}

// Create creates new metric events if SLOs are specified
func (mec MetricEventCreation) Create(project string, stage string, service string) []dynatrace.ConfigResult {
	var metricEvents []dynatrace.ConfigResult
	if !lib.IsMetricEventsGenerationEnabled() {
		return metricEvents
	}

	log.Info("Creating custom metric events for project SLIs")
	slos, err := retrieveSLOs(project, stage, service)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"service": service,
				"stage":   stage}).Info("No SLOs defined for service. Skipping creation of custom metric events.")
		return metricEvents
	}
	// get custom metrics for project
	projectCustomQueries, err := mec.getCustomQueries(project, stage, service)
	if err != nil {
		log.WithError(err).WithField("project", project).Error("Failed to get custom queries for project")
		return metricEvents
	}

	managementZones, err := dynatrace.NewManagementZonesClient(mec.client).GetAll()
	var mzId int64 = -1
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"project": project, "stage": stage}).Error("Could not retrieve management zones")
		return metricEvents
	}

	// TODO 2021-08-20: check the logic below - if parsing management zone id does not work, we will continue anyway?
	if zone, wasFound := managementZones.GetBy(GetManagementZoneNameForProjectAndStage(project, stage)); wasFound {
		mzId, err = strconv.ParseInt(zone.ID, 10, 64)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"project": project, "stage": stage}).Warn("Could not parse management zone ID")
		}
	} else {
		log.WithError(err).WithFields(log.Fields{"project": project, "stage": stage}).Warn("Could not find management zone")
		return metricEvents
	}

	metricEventCreated := false
	// try to create metric events using best effort.
	for _, objective := range slos.Objectives {
		config, err := getTimeseriesConfig(objective.SLI, projectCustomQueries)
		if err != nil {
			// Error occurred but continue
			log.WithField("sli", objective.SLI).Error("Could not find query for SLI")
		}
		for _, criteria := range objective.Pass {
			for _, crit := range criteria.Criteria {
				// criteria.Criteria
				criteriaObject, err := parseCriteriaString(crit)
				if err != nil {
					// Error occurred but continue
					log.WithError(err).WithField("criteria", crit).Error("Could not parse criteria")
					continue
				}
				if criteriaObject.IsComparison {
					// comparison-based criteria cannot be mapped to alerts
					continue
				}
				newMetricEvent, err := createKeptnMetricEvent(project, stage, service, objective.SLI, config, crit, criteriaObject.Value, mzId)
				if err != nil {
					// Error occurred but continue
					log.WithError(err).WithFields(
						log.Fields{
							"sli":      objective.SLI,
							"criteria": crit,
						}).Error("Could not create metric event definition for criteria")
					continue
				}

				event, err := mec.GetMetricEvent(newMetricEvent.Name)
				if err != nil {
					// Error occurred but continue
					log.WithError(err).Error("Could not get metric event")
					continue
				}

				apiURL := "/api/config/v1/anomalyDetection/metricEvents"
				apiMethod := "POST"

				mePayload, err := json.Marshal(newMetricEvent)
				if err != nil {
					// Error occurred but continue
					log.WithError(err).Error("Could not marshal metric event")
					continue
				}

				if event != nil {
					// adapt all properties that have initially been defaulted to some value from previous (potentially modified event)
					event.Threshold = newMetricEvent.Threshold
					event.TagFilters = nil
					apiURL = apiURL + "/" + event.ID
					apiMethod = "PUT"
					mePayload, err = json.Marshal(event)
					if err != nil {
						// Error occurred but continue
						log.WithError(err).Error("Could not marshal metric event")
						continue
					}
				}

				_, err = mec.client.SendDynatraceAPIRequest(apiURL, apiMethod, mePayload)
				if err != nil {
					log.WithError(err).WithField("metricName", newMetricEvent.Name).Error("Could not create metric event")
					continue
				}
				metricEvents = append(metricEvents,
					dynatrace.ConfigResult{
						Name:    newMetricEvent.Name,
						Success: true,
					})
				log.WithFields(
					log.Fields{
						"name":     newMetricEvent.Name,
						"criteria": crit,
					}).Info("Created metric event")
				metricEventCreated = true
			}
		}
	}

	if metricEventCreated {
		// TODO: improve this?
		log.Info("To review and enable the generated custom metric events, please go to: https://" + mec.client.DynatraceCreds.Tenant + "/#settings/anomalydetection/metricevents")
	}

	return metricEvents
}

func (mec *MetricEventCreation) getCustomQueries(project string, stage string, service string) (map[string]string, error) {

	if mec.client.KeptnHandler == nil {
		return nil, errors.New("Could not retrieve SLI config: No KeptnHandler initialized")
	}
	customQueries, err := mec.client.KeptnHandler.GetSLIConfiguration(project, stage, service, sliResourceURI)
	if err != nil {
		return nil, err
	}

	return customQueries, nil
}

func (mec *MetricEventCreation) GetMetricEvent(eventKey string) (*dynatrace.MetricEvent, error) {
	res, err := mec.client.SendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents", "GET", nil)
	if err != nil {
		log.WithError(err).Error("Could not retrieve list of existing Dynatrace metric events")
		return nil, err
	}

	dtMetricEvents := &dynatrace.DTAPIListResponse{}
	err = json.Unmarshal([]byte(res), dtMetricEvents)

	if err != nil {
		log.WithError(err).Error("Could not parse list of existing Dynatrace metric events")
		return nil, err
	}

	for _, metricEvent := range dtMetricEvents.Values {
		if metricEvent.Name == eventKey {
			res, err = mec.client.SendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents/"+metricEvent.ID, "GET", nil)
			if err != nil {
				log.WithError(err).WithField("eventKey", eventKey).Error("Could not get existing metric event")
				return nil, err
			}
			retrievedMetricEvent := &dynatrace.MetricEvent{}
			err = json.Unmarshal([]byte(res), retrievedMetricEvent)
			if err != nil {
				return nil, err
			}
			return retrievedMetricEvent, nil
		}
	}
	return nil, nil
}

func (mec *MetricEventCreation) DeleteExistingMetricEvent(eventKey string) error {
	res, err := mec.client.SendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents", "GET", nil)
	if err != nil {
		log.WithError(err).Error("Could not retrieve list of existing Dynatrace metric events")
		return err
	}

	dtMetricEvents := &dynatrace.DTAPIListResponse{}
	err = json.Unmarshal([]byte(res), dtMetricEvents)

	if err != nil {
		log.WithError(err).Error("Could not parse list of existing Dynatrace metric events")
		return err
	}

	for _, metricEvent := range dtMetricEvents.Values {
		if metricEvent.Name == eventKey {
			res, err = mec.client.SendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents/"+metricEvent.ID, "DELETE", nil)
			if err != nil {
				log.WithError(err).WithField("eventKey", eventKey).Error("Could not delete existing metric event")
				return err
			}
		}
	}
	return nil
}

func retrieveSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error) {
	resourceHandler := configutils.NewResourceHandler(common.GetConfigurationServiceURL())

	resource, err := resourceHandler.GetServiceResource(project, stage, service, "slo.yaml")
	if err != nil || resource.ResourceContent == "" {
		return nil, errors.New("No SLO file available for service " + service + " in stage " + stage)
	}
	var slos keptn.ServiceLevelObjectives

	err = yaml.Unmarshal([]byte(resource.ResourceContent), &slos)

	if err != nil {
		return nil, errors.New("Invalid SLO file format")
	}

	return &slos, nil
}

// based on the requested metric a dynatrace timeseries with its aggregation type is returned
func getTimeseriesConfig(metric string, customQueries map[string]string) (string, error) {
	if val, ok := customQueries[metric]; ok {
		return val, nil
	}

	// default config
	switch metric {
	case throughput:
		return "builtin:service.requestCount.total:merge(0):count?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case errorRate:
		return "builtin:service.errors.total.rate:merge(0):avg?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP50:
		return "builtin:service.response.time:merge(0):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP90:
		return "builtin:service.response.time:merge(0):percentile(90)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP95:
		return "builtin:service.response.time:merge(0):percentile(95)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	default:
		return "", fmt.Errorf("unsupported SLI metric %s", metric)
	}
}

func parseCriteriaString(criteria string) (*dynatrace.CriteriaObject, error) {
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

	c := &dynatrace.CriteriaObject{}

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

func createKeptnMetricEvent(project string, stage string, service string, metric string, query string, condition string, threshold float64, managementZoneID int64) (*dynatrace.MetricEvent, error) {

	/*
		need to map queries used by SLI-service to metric event definition.
		example: builtin:service.response.time:merge(0):percentile(90)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)

		1. split by '?' and get first part => builtin:service.response.time:merge(0):percentile(90)
		2. split by ':' => builtin:service.response.time | merge(0) | percentile(90) => merge(0) is not needed
		3. first part is the metricId and can be used for the Metric Event API => builtin:service.response.time
		4. Aggregation is limited to: AVG, COUNT, MAX, MEDIAN, MIN, OF_INTEREST, OF_INTEREST_RATIO, OTHER, OTHER_RATIO, P90, SUM, VALUE
	*/

	if project == "" || stage == "" || service == "" || metric == "" || query == "" {
		return nil, errors.New("missing input parameter values")
	}

	query = strings.TrimPrefix(query, "metricSelector=")
	// 1. split by '?' and get first part => builtin:service.response.time:merge(0):percentile(90)
	split := strings.Split(query, "?")

	// 2. split by ':' => builtin:service.response.time | merge(0) | percentile(90) => merge(0) is not needed/supported by MetricEvent API
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

	meAlertCondition, err := getAlertCondition(condition)
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

func getAlertCondition(condition string) (string, error) {
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

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

type MetricEventCreation struct {
	client *dynatrace.DynatraceHelper
}

func NewMetricEventCreation(client *dynatrace.DynatraceHelper) MetricEventCreation {
	return MetricEventCreation{
		client: client,
	}
}

// CreateFor creates new metric events if SLOs are specified
func (mec MetricEventCreation) CreateFor(project string, stage string, service string) []dynatrace.ConfigResult {
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

	if zone, wasFound := managementZones.GetBy(GetManagementZoneNameForProjectAndStage(project, stage)); wasFound {
		mzId, err = strconv.ParseInt(zone.ID, 10, 64)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"project": project, "stage": stage}).Warn("Could not parse management zone ID")
			return metricEvents
		}
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
				newMetricEvent, err := dynatrace.CreateKeptnMetricEvent(project, stage, service, objective.SLI, config, crit, criteriaObject.Value, mzId)
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
	customQueries, err := mec.client.KeptnHandler.GetSLIConfiguration(project, stage, service, dynatrace.SliResourceURI)
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
	case dynatrace.Throughput:
		return "builtin:service.requestCount.total:merge(0):count?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case dynatrace.ErrorRate:
		return "builtin:service.errors.total.rate:merge(0):avg?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case dynatrace.ResponseTimeP50:
		return "builtin:service.response.time:merge(0):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case dynatrace.ResponseTimeP90:
		return "builtin:service.response.time:merge(0):percentile(90)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case dynatrace.ResponseTimeP95:
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

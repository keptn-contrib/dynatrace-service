package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"

	configutils "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

// CreateMetricEvents creates new metric events if SLOs are specified
func (dt *DynatraceHelper) CreateMetricEvents(project string, stage string, service string) {
	if !IsMetricEventsGenerationEnabled() {
		return
	}

	log.Info("Creating custom metric events for project SLIs")
	slos, err := retrieveSLOs(project, stage, service)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"service": service,
				"stage":   stage}).Info("No SLOs defined for service. Skipping creation of custom metric events.")
		return
	}
	// get custom metrics for project
	projectCustomQueries, err := dt.getCustomQueries(project, stage, service)
	if err != nil {
		log.WithError(err).WithField("project", project).Error("Failed to get custom queries for project")
		return
	}

	managementZones := dt.getManagementZones()
	var mzId int64 = -1
	for _, mz := range managementZones {
		if mz.Name == getManagementZoneNameForStage(project, stage) {
			mzId, err = strconv.ParseInt(mz.ID, 10, 64)
			if err != nil {
				log.WithError(err).Warn("Could not parse management zone ID")
			}
		}
	}
	if mzId < 0 {
		log.WithFields(log.Fields{
			"project": project,
			"stage":   stage,
		}).Error("No management zone found")
		return
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
				newMetricEvent, err := CreateKeptnMetricEvent(project, stage, service, objective.SLI, config, crit, criteriaObject.Value, mzId)
				if err != nil {
					// Error occurred but continue
					log.WithError(err).WithFields(
						log.Fields{
							"sli":      objective.SLI,
							"criteria": crit,
						}).Error("Could not create metric event definition for criteria")
					continue
				}

				event, err := dt.GetMetricEvent(newMetricEvent.Name)
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

				_, err = dt.sendDynatraceAPIRequest(apiURL, apiMethod, mePayload)
				if err != nil {
					log.WithError(err).WithField("metricName", newMetricEvent.Name).Error("Could not create metric event")
					continue
				}
				dt.configuredEntities.MetricEvents = append(dt.configuredEntities.MetricEvents, ConfigResult{
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
		log.Info("To review and enable the generated custom metric events, please go to: https://" + dt.DynatraceCreds.Tenant + "/#settings/anomalydetection/metricevents")
	}
}

func (dt *DynatraceHelper) getCustomQueries(project string, stage string, service string) (map[string]string, error) {

	if dt.KeptnHandler == nil {
		return nil, errors.New("Could not retrieve SLI config: No KeptnHandler initialized")
	}
	customQueries, err := dt.KeptnHandler.GetSLIConfiguration(project, stage, service, sliResourceURI)
	if err != nil {
		return nil, err
	}

	return customQueries, nil
}

func (dt *DynatraceHelper) GetMetricEvent(eventKey string) (*MetricEvent, error) {
	res, err := dt.sendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents", "GET", nil)
	if err != nil {
		log.WithError(err).Error("Could not retrieve list of existing Dynatrace metric events")
		return nil, err
	}

	dtMetricEvents := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(res), dtMetricEvents)

	if err != nil {
		log.WithError(err).Error("Could not parse list of existing Dynatrace metric events")
		return nil, err
	}

	for _, metricEvent := range dtMetricEvents.Values {
		if metricEvent.Name == eventKey {
			res, err = dt.sendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents/"+metricEvent.ID, "GET", nil)
			if err != nil {
				log.WithError(err).WithField("eventKey", eventKey).Error("Could not get existing metric event")
				return nil, err
			}
			retrievedMetricEvent := &MetricEvent{}
			err = json.Unmarshal([]byte(res), retrievedMetricEvent)
			if err != nil {
				return nil, err
			}
			return retrievedMetricEvent, nil
		}
	}
	return nil, nil
}

func (dt *DynatraceHelper) DeleteExistingMetricEvent(eventKey string) error {
	res, err := dt.sendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents", "GET", nil)
	if err != nil {
		log.WithError(err).Error("Could not retrieve list of existing Dynatrace metric events")
		return err
	}

	dtMetricEvents := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(res), dtMetricEvents)

	if err != nil {
		log.WithError(err).Error("Could not parse list of existing Dynatrace metric events")
		return err
	}

	for _, metricEvent := range dtMetricEvents.Values {
		if metricEvent.Name == eventKey {
			res, err = dt.sendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents/"+metricEvent.ID, "DELETE", nil)
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
	case Throughput:
		return "builtin:service.requestCount.total:merge(0):count?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ErrorRate:
		return "builtin:service.errors.total.rate:merge(0):avg?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ResponseTimeP50:
		return "builtin:service.response.time:merge(0):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ResponseTimeP90:
		return "builtin:service.response.time:merge(0):percentile(90)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ResponseTimeP95:
		return "builtin:service.response.time:merge(0):percentile(95)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	default:
		fmt.Sprintf("Unknown metric %s\n", metric)
		return "", fmt.Errorf("unsupported SLI metric %s", metric)
	}
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

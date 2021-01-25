package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

	dt.Logger.Info("Creating custom metric events for project SLIs")
	slos, err := retrieveSLOs(project, stage, service)
	if err != nil {
		dt.Logger.Info("No SLOs defined for service " + service + " in stage " + stage + ". Skipping creation of custom metric events.")
		return
	}
	// get custom metrics for project
	projectCustomQueries, err := dt.getCustomQueries(project, stage, service)
	if err != nil {
		dt.Logger.Error(fmt.Sprintf("failed to get custom queries for project %s: %v", project, err))
		return
	}

	managementZones := dt.getManagementZones()
	var mzId int64 = -1
	for _, mz := range managementZones {
		if mz.Name == getManagementZoneNameForStage(project, stage) {
			mzId, _ = strconv.ParseInt(mz.ID, 10, 64)
		}
	}
	if mzId < 0 {
		dt.Logger.Error(fmt.Sprintf("no management zone found for project %s and stage %s", project, stage))
		return
	}

	metricEventCreated := false
	// try to create metric events using best effort.
	for _, objective := range slos.Objectives {
		config, err := getTimeseriesConfig(objective.SLI, projectCustomQueries)
		if err != nil {
			// Error occurred but continue
			dt.Logger.Error("Could not find query for SLI " + objective.SLI)
		}
		for _, criteria := range objective.Pass {
			for _, crit := range criteria.Criteria {
				// criteria.Criteria
				criteriaObject, err := parseCriteriaString(crit)
				if err != nil {
					// Error occurred but continue
					dt.Logger.Error("Could not parse criteria " + crit + ": " + err.Error())
					continue
				}
				if criteriaObject.IsComparison {
					// comparison-based criteria cannot be mapped to alerts
					continue
				}
				newMetricEvent, err := CreateKeptnMetricEvent(project, stage, service, objective.SLI, config, crit, criteriaObject.Value, mzId)
				if err != nil {
					// Error occurred but continue
					dt.Logger.Error("Could not create metric event definition for criteria " + objective.SLI + "" + crit + ": " + err.Error())
					continue
				}

				event, err := dt.GetMetricEvent(newMetricEvent.Name)
				if err != nil {
					// Error occurred but continue
					dt.Logger.Error("Could not get metric event: " + err.Error())
					continue
				}

				apiURL := "/api/config/v1/anomalyDetection/metricEvents"
				apiMethod := "POST"

				mePayload, err := json.Marshal(newMetricEvent)
				if err != nil {
					// Error occurred but continue
					dt.Logger.Error("Could not marshal metric event: " + err.Error())
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
						dt.Logger.Error("Could not marshal metric event: " + err.Error())
						continue
					}
				}

				_, err = dt.sendDynatraceAPIRequest(apiURL, apiMethod, mePayload)
				if err != nil {
					dt.Logger.Error("Could not create metric event " + newMetricEvent.Name + ": " + err.Error())
					continue
				}
				dt.configuredEntities.MetricEvents = append(dt.configuredEntities.MetricEvents, ConfigResult{
					Name:    newMetricEvent.Name,
					Success: true,
				})
				dt.Logger.Info("Created metric event " + newMetricEvent.Name + " " + crit)
				metricEventCreated = true
			}
		}
	}

	if metricEventCreated {
		dt.Logger.Info("To review and enable the generated custom metric events, please go to: https://" + dt.DynatraceCreds.Tenant + "/#settings/anomalydetection/metricevents")
	}
	return
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
		dt.Logger.Error("Could not retrieve list of existing Dynatrace metric events: " + err.Error())
		return nil, err
	}

	dtMetricEvents := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(res), dtMetricEvents)

	if err != nil {
		dt.Logger.Error("Could not parse list of existing Dynatrace metric events: " + err.Error())
		return nil, err
	}

	for _, metricEvent := range dtMetricEvents.Values {
		if metricEvent.Name == eventKey {
			res, err = dt.sendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents/"+metricEvent.ID, "GET", nil)
			if err != nil {
				dt.Logger.Error("Could not get existing metric event " + eventKey + ": " + err.Error())
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
		dt.Logger.Error("Could not retrieve list of existing Dynatrace metric events: " + err.Error())
		return err
	}

	dtMetricEvents := &DTAPIListResponse{}
	err = json.Unmarshal([]byte(res), dtMetricEvents)

	if err != nil {
		dt.Logger.Error("Could not parse list of existing Dynatrace metric events: " + err.Error())
		return err
	}

	for _, metricEvent := range dtMetricEvents.Values {
		if metricEvent.Name == eventKey {
			res, err = dt.sendDynatraceAPIRequest("/api/config/v1/anomalyDetection/metricEvents/"+metricEvent.ID, "DELETE", nil)
			if err != nil {
				dt.Logger.Error("Could not delete existing metric event " + eventKey + ": " + err.Error())
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

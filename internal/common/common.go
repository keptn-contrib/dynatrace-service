package common

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/keptn"
)

// This is the label name for the Problem URL label
const PROBLEMURL_LABEL = "Problem URL"
const KEPTNSBRIDGE_LABEL = "Keptns Bridge"

const shipyardController = "SHIPYARD_CONTROLLER"
const configurationService = "CONFIGURATION_SERVICE"
const defaultShipyardControllerURL = "http://shipyard-controller:8080"
const defaultConfigurationServiceURL = "http://configuration-service:8080"

// GetConfigurationServiceURL Returns the endpoint to the configuration-service
func GetConfigurationServiceURL() string {
	/*
		// TODO: check previous alternate implementation:

		if os.Getenv("CONFIGURATION_SERVICE") != "" {
			return os.Getenv("CONFIGURATION_SERVICE")
		}
		return "configuration-service:8080"
	*/

	return getKeptnServiceURL(configurationService, defaultConfigurationServiceURL)
}

// GetShipyardControllerURL Returns the endpoint to the shipyard-controller
func GetShipyardControllerURL() string {
	return getKeptnServiceURL(shipyardController, defaultShipyardControllerURL)
}

func getKeptnServiceURL(servicename, defaultURL string) string {
	var baseURL string
	url, err := keptn.GetServiceEndpoint(servicename)
	if err == nil {
		baseURL = url.String()
	} else {
		baseURL = defaultURL
	}
	return baseURL
}

/**
 * Constants for supporting resource files in keptn repo
 */

/**
 * Defines the Dynatrace Configuration File structure and supporting Constants
 */
const DynatraceConfigDashboardQUERY = "query"

//
// replaces $ placeholders with actual values
// $CONTEXT, $EVENT, $SOURCE
// $PROJECT, $STAGE, $SERVICE, $DEPLOYMENT
// $TESTSTRATEGY
// $LABEL.XXXX  -> will replace that with a label called XXXX
// $ENV.XXXX    -> will replace that with an env variable called XXXX
// $SECRET.YYYY -> will replace that with the k8s secret called YYYY
//
func ReplaceKeptnPlaceholders(input string, keptnEvent adapter.EventContentAdapter) string {
	result := input

	// FIXING on 27.5.2020: URL Escaping of parameters as described in https://github.com/keptn-contrib/dynatrace-sli-service/issues/54

	// first we do the regular keptn values
	result = strings.Replace(result, "$CONTEXT", url.QueryEscape(keptnEvent.GetShKeptnContext()), -1)
	result = strings.Replace(result, "$EVENT", url.QueryEscape(keptnEvent.GetEvent()), -1)
	result = strings.Replace(result, "$SOURCE", url.QueryEscape(keptnEvent.GetSource()), -1)
	result = strings.Replace(result, "$PROJECT", url.QueryEscape(keptnEvent.GetProject()), -1)
	result = strings.Replace(result, "$STAGE", url.QueryEscape(keptnEvent.GetStage()), -1)
	result = strings.Replace(result, "$SERVICE", url.QueryEscape(keptnEvent.GetService()), -1)
	result = strings.Replace(result, "$DEPLOYMENT", url.QueryEscape(keptnEvent.GetDeployment()), -1)
	result = strings.Replace(result, "$TESTSTRATEGY", url.QueryEscape(keptnEvent.GetTestStrategy()), -1)

	// now we do the labels
	for key, value := range keptnEvent.GetLabels() {
		result = strings.Replace(result, "$LABEL."+key, url.QueryEscape(value), -1)
	}

	// now we do all environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		result = strings.Replace(result, "$ENV."+pair[0], url.QueryEscape(pair[1]), -1)
	}

	// TODO: iterate through k8s secrets!

	return result
}

// ParseUnixTimestamp parses a time stamp into Unix foramt
func ParseUnixTimestamp(timestamp string) (time.Time, error) {
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err == nil {
		return parsedTime, nil
	}

	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Now(), err
	}
	unix := time.Unix(timestampInt, 0)
	return unix, nil
}

// TimestampToString converts time stamp into string
func TimestampToString(time time.Time) string {
	return strconv.FormatInt(time.Unix()*1000, 10)
}

func ParsePassAndWarningWithoutDefaultsFrom(customName string) *keptncommon.SLO {
	return ParsePassAndWarningFromString(customName, []string{}, []string{})
}

// ParsePassAndWarningFromString takes a value such as
//   Example 1: Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true
//   Example 2: Response time (P95);sli=svc_rt_p95;pass=<+10%,<600
//   Example 3: Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false
// can also take a value like
// 	 "KQG;project=myproject;pass=90%;warning=75%;"
// This will return a SLO object
func ParsePassAndWarningFromString(customName string, defaultPass []string, defaultWarning []string) *keptncommon.SLO {
	result := &keptncommon.SLO{
		Weight:  1,
		KeySLI:  false,
		Pass:    []*keptncommon.SLOCriteria{},
		Warning: []*keptncommon.SLOCriteria{},
	}

	nameValueSplits := strings.Split(customName, ";")

	// lets iterate through all name-value pairs which are separated through ";" to extract keys such as warning, pass, weight, key, sli
	for i := 0; i < len(nameValueSplits); i++ {

		nameValueDividerIndex := strings.Index(nameValueSplits[i], "=")
		if nameValueDividerIndex < 0 {
			continue
		}

		// for each name=value pair we get the name as first part of the string until the first =
		// the value is the after that =
		nameString := strings.ToLower(nameValueSplits[i][:nameValueDividerIndex])
		valueString := nameValueSplits[i][nameValueDividerIndex+1:]
		var err error
		switch nameString /*nameValueSplit[0]*/ {
		case "sli":
			result.SLI = valueString
		case "pass":
			result.Pass = append(
				result.Pass,
				&keptncommon.SLOCriteria{Criteria: strings.Split(valueString, ",")})
		case "warning":
			result.Warning = append(
				result.Warning,
				&keptncommon.SLOCriteria{Criteria: strings.Split(valueString, ",")})
		case "key":
			result.KeySLI, err = strconv.ParseBool(valueString)
			if err != nil {
				log.WithError(err).Warn("Error parsing bool")
			}
		case "weight":
			result.Weight, err = strconv.Atoi(valueString)
			if err != nil {
				log.WithError(err).Warn("Error parsing weight")
			}
		}
	}

	// use the defaults if nothing was specified
	if (len(result.Pass) == 0) && (len(defaultPass) > 0) {
		result.Pass = append(result.Pass, &keptncommon.SLOCriteria{Criteria: defaultPass})
	}

	if (len(result.Warning) == 0) && (len(defaultWarning) > 0) {
		result.Warning = append(result.Warning, &keptncommon.SLOCriteria{Criteria: defaultWarning})
	}

	// if we have no criteria for warn or pass we just return nil
	if len(result.Pass) == 0 {
		result.Pass = nil
	}
	if len(result.Warning) == 0 {
		result.Warning = nil
	}

	return result
}

// CleanIndicatorName makes sure we have a valid indicator name by getting rid of special characters
func CleanIndicatorName(indicatorName string) string {
	indicatorName = strings.ReplaceAll(indicatorName, " ", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "/", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "%", "_")

	return indicatorName
}

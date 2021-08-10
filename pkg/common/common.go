package common

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"

	keptnmodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// This is the label name for the Problem URL label
const PROBLEMURL_LABEL = "Problem URL"
const KEPTNSBRIDGE_LABEL = "Keptns Bridge"

const shipyardController = "SHIPYARD_CONTROLLER"
const configurationService = "CONFIGURATION_SERVICE"
const defaultShipyardControllerURL = "http://shipyard-controller:8080"
const defaultConfigurationServiceURL = "http://configuration-service:8080"

// RunLocal is true if the "ENV"-environment variable is set to local
var RunLocal = os.Getenv("ENV") == "local"

// RunLocalTest is true if the "ENV"-environment variable is set to localtest
var RunLocalTest = os.Getenv("ENV") == "localtest"

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	if RunLocal || RunLocalTest {
		return nil, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

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

// KeptnCredentials contains the credentials for the Keptn API
type KeptnCredentials struct {
	ApiURL   string
	ApiToken string
}

/**
 * Finds the Problem ID that is associated with this Keptn Workflow
 * It first parses it from Problem URL label - if it cant be found there it will look for the Initial Problem Open Event and gets the ID from there!
 */
func FindProblemIDForEvent(keptnHandler *keptnv2.Keptn, labels map[string]string) (string, error) {

	// Step 1 - see if we have a Problem Url in the labels
	// iterate through the labels and find Problem URL
	for labelName, labelValue := range labels {
		if labelName == PROBLEMURL_LABEL {
			// the value should be of form https://dynatracetenant/#problems/problemdetails;pid=8485558334848276629_1604413609638V2
			// so - lets get the last part after pid=

			ix := strings.LastIndex(labelValue, ";pid=")
			if ix > 0 {
				return labelValue[ix+5:], nil
			}
		}
	}

	// Step 2 - lets see if we have a ProblemOpenEvent for this KeptnContext - if so - we try to extract the Problem ID
	eventHandler := keptnapi.NewEventHandler(os.Getenv("DATASTORE"))

	events, errObj := eventHandler.GetEvents(&keptnapi.EventFilter{
		Project:      keptnHandler.KeptnBase.Event.GetProject(),
		EventType:    keptncommon.ProblemOpenEventType,
		KeptnContext: keptnHandler.KeptnContext,
	})

	if errObj != nil {
		msg := "cannot send DT problem comment: Could not retrieve problem.open event for incoming event: " + *errObj.Message
		return "", errors.New(msg)
	}

	if len(events) == 0 {
		msg := "cannot send DT problem comment: Could not retrieve problem.open event for incoming event: no events returned"
		return "", errors.New(msg)
	}

	problemOpenEvent := &keptncommon.ProblemEventData{}
	err := keptnv2.Decode(events[0].Data, problemOpenEvent)

	if err != nil {
		msg := "could not decode problem.open event: " + err.Error()
		return "", errors.New(msg)
	}

	if problemOpenEvent.PID == "" {
		return "", errors.New("cannot send DT problem comment: No problem ID is included in the event")
	}

	return problemOpenEvent.PID, nil
}

/**
 * Constants for supporting resource files in keptn repo
 */
const DynatraceDashboardFilename = "dynatrace/dashboard.json"
const DynatraceSLIFilename = "dynatrace/sli.yaml"
const KeptnSLOFilename = "slo.yaml"
const KeptnSLIResultFilename = "sliresult.json"

const ConfigLevelProject = "Project"
const ConfigLevelStage = "Stage"
const ConfigLevelService = "Service"

/**
 * Defines the Dynatrace Configuration File structure and supporting Constants
 */
const DynatraceConfigFilename = "dynatrace/dynatrace.conf.yaml"
const DynatraceConfigFilenameLOCAL = "dynatrace/_dynatrace.conf.yaml"
const DynatraceConfigDashboardQUERY = "query"

type DynatraceConfigFile struct {
	SpecVersion string `json:"spec_version" yaml:"spec_version"`
	DtCreds     string `json:"dtCreds,omitempty" yaml:"dtCreds,omitempty"`
	Dashboard   string `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`
}

type DTCredentials struct {
	Tenant    string `json:"DT_TENANT" yaml:"DT_TENANT"`
	ApiToken  string `json:"DT_API_TOKEN" yaml:"DT_API_TOKEN"`
	PaaSToken string `json:"DT_PAAS_TOKEN" yaml:"DT_PAAS_TOKEN"`
}

type BaseKeptnEvent struct {
	Context string
	Source  string
	Event   string

	Project            string
	Stage              string
	Service            string
	Deployment         string
	TestStrategy       string
	DeploymentStrategy string

	Image string
	Tag   string

	Labels map[string]string
}

var namespace = getPodNamespace()

func getPodNamespace() string {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return "keptn"
	}

	return ns
}

//
// replaces $ placeholders with actual values
// $CONTEXT, $EVENT, $SOURCE
// $PROJECT, $STAGE, $SERVICE, $DEPLOYMENT
// $TESTSTRATEGY
// $LABEL.XXXX  -> will replace that with a label called XXXX
// $ENV.XXXX    -> will replace that with an env variable called XXXX
// $SECRET.YYYY -> will replace that with the k8s secret called YYYY
//
func ReplaceKeptnPlaceholders(input string, keptnEvent *BaseKeptnEvent) string {
	result := input

	// FIXING on 27.5.2020: URL Escaping of parameters as described in https://github.com/keptn-contrib/dynatrace-sli-service/issues/54

	// first we do the regular keptn values
	result = strings.Replace(result, "$CONTEXT", url.QueryEscape(keptnEvent.Context), -1)
	result = strings.Replace(result, "$EVENT", url.QueryEscape(keptnEvent.Event), -1)
	result = strings.Replace(result, "$SOURCE", url.QueryEscape(keptnEvent.Source), -1)
	result = strings.Replace(result, "$PROJECT", url.QueryEscape(keptnEvent.Project), -1)
	result = strings.Replace(result, "$STAGE", url.QueryEscape(keptnEvent.Stage), -1)
	result = strings.Replace(result, "$SERVICE", url.QueryEscape(keptnEvent.Service), -1)
	result = strings.Replace(result, "$DEPLOYMENT", url.QueryEscape(keptnEvent.Deployment), -1)
	result = strings.Replace(result, "$TESTSTRATEGY", url.QueryEscape(keptnEvent.TestStrategy), -1)

	// now we do the labels
	for key, value := range keptnEvent.Labels {
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

//
// Downloads a resource from the Keptn Configuration Repo based on the level (Project, Stage, Service)
// In RunLocal mode it gets it from the local disk
//
func GetKeptnResourceOnConfigLevel(keptnEvent *BaseKeptnEvent, resourceURI string, level string) (string, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	var fileContent string
	if RunLocal {
		resourceURI = strings.ToLower(strings.ReplaceAll(resourceURI, "dynatrace/", "../../../dynatrace/"+level+"_"))
		localFileContent, err := ioutil.ReadFile(resourceURI)
		if err != nil {
			log.WithFields(
				log.Fields{
					"resourceURI": resourceURI,
					"service":     keptnEvent.Service,
					"stage":       keptnEvent.Stage,
					"project":     keptnEvent.Project,
				}).Info("File not found locally")
			return "", nil
		}
		log.WithField("resourceURI", resourceURI).Info("Loaded LOCAL file")
		fileContent = string(localFileContent)
	} else {
		resourceHandler := keptnapi.NewResourceHandler(GetConfigurationServiceURL())

		var keptnResourceContent *keptnmodels.Resource
		var err error
		if strings.Compare(level, ConfigLevelProject) == 0 {
			keptnResourceContent, err = resourceHandler.GetProjectResource(keptnEvent.Project, resourceURI)
		} else if strings.Compare(level, ConfigLevelStage) == 0 {
			keptnResourceContent, err = resourceHandler.GetStageResource(keptnEvent.Project, keptnEvent.Stage, resourceURI)
		} else if strings.Compare(level, ConfigLevelService) == 0 {
			keptnResourceContent, err = resourceHandler.GetServiceResource(keptnEvent.Project, keptnEvent.Stage, keptnEvent.Service, resourceURI)
		} else {
			return "", errors.New("Config level not valid: " + level)
		}

		if err != nil {
			return "", err
		}

		if keptnResourceContent == nil {
			return "", errors.New("Found resource " + resourceURI + " on level " + level + " but didnt contain content")
		}

		fileContent = keptnResourceContent.ResourceContent
	}

	return fileContent, nil
}

//
// Downloads a resource from the Keptn Configuration Repo
// In RunLocal mode it gets it from the local disk
// In normal mode it first tries to find it on service level, then stage and then project level
//
func GetKeptnResource(keptnEvent *BaseKeptnEvent, resourceURI string) (string, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	var fileContent string
	if RunLocal {
		localFileContent, err := ioutil.ReadFile(resourceURI)
		if err != nil {
			log.WithFields(
				log.Fields{
					"resourceURI": resourceURI,
					"service":     keptnEvent.Service,
					"stage":       keptnEvent.Stage,
					"project":     keptnEvent.Project,
				}).Info("File not found locally")
			return "", nil
		}
		log.WithField("resourceURI", resourceURI).Info("Loaded LOCAL file")
		fileContent = string(localFileContent)
	} else {
		resourceHandler := keptnapi.NewResourceHandler(GetConfigurationServiceURL())

		// Lets search on SERVICE-LEVEL
		keptnResourceContent, err := resourceHandler.GetServiceResource(keptnEvent.Project, keptnEvent.Stage, keptnEvent.Service, resourceURI)
		if err != nil || keptnResourceContent == nil || keptnResourceContent.ResourceContent == "" {
			// Lets search on STAGE-LEVEL
			keptnResourceContent, err = resourceHandler.GetStageResource(keptnEvent.Project, keptnEvent.Stage, resourceURI)
			if err != nil || keptnResourceContent == nil || keptnResourceContent.ResourceContent == "" {
				// Lets search on PROJECT-LEVEL
				keptnResourceContent, err = resourceHandler.GetProjectResource(keptnEvent.Project, resourceURI)
				if err != nil || keptnResourceContent == nil || keptnResourceContent.ResourceContent == "" {
					// log.Debugf("No Keptn Resource found: %s/%s/%s/%s - %s", keptnEvent.Project, keptnEvent.Stage, keptnEvent.Service, resourceURI, err)
					return "", err
				}

				log.WithFields(
					log.Fields{
						"resourceURI": resourceURI,
						"project":     keptnEvent.Project,
					}).Debug("Found resource on project level")
			} else {
				log.WithFields(
					log.Fields{
						"resourceURI": resourceURI,
						"project":     keptnEvent.Project,
						"stage":       keptnEvent.Stage,
					}).Debug("Found resource on stage level")
			}
		} else {
			log.WithFields(
				log.Fields{
					"resourceURI": resourceURI,
					"project":     keptnEvent.Project,
					"stage":       keptnEvent.Stage,
					"service":     keptnEvent.Service,
				}).Debug("Found resource on service level")
		}
		fileContent = keptnResourceContent.ResourceContent
	}

	return fileContent, nil
}

/**
 * Loads SLIs from a local file and adds it to the SLI map
 */
func addResourceContentToSLIMap(SLIs map[string]string, sliFileContent string) (map[string]string, error) {

	if sliFileContent != "" {
		sliConfig := keptn.SLIConfig{}
		err := yaml.Unmarshal([]byte(sliFileContent), &sliConfig)
		if err != nil {
			return nil, err
		}

		for key, value := range sliConfig.Indicators {

			if _, keyPresent := SLIs[key]; keyPresent {
				log.WithFields(
					log.Fields{"key": key,
						"value": value,
					}).Warn("Overwriting SLI in SLIMap")
			}
			SLIs[key] = value
		}
	}
	return SLIs, nil
}

/**
 * getCustomQueries loads custom SLIs from dynatrace/sli.yaml
 * if there is no sli.yaml it will just return an empty map
 */
func GetCustomQueries(keptnEvent *BaseKeptnEvent) map[string]string {
	var sliMap = map[string]string{}
	/*if common.RunLocal || common.RunLocalTest {
		sliMap, _ = AddResourceContentToSLIMap(sliMap, "dynatrace/sli.yaml", "")
		return sliMap, nil
	}*/

	// We need to load sli.yaml in the sequence of project, stage then service level where service level will overwrite stage & project and stage will overwrite project level sli defintions
	// details can be found here: https://github.com/keptn-contrib/dynatrace-sli-service/issues/112

	// Step 1: Load Project Level
	foundLocation := ""
	sliContent, err := GetKeptnResourceOnConfigLevel(keptnEvent, DynatraceSLIFilename, ConfigLevelProject)
	if err != nil {
		log.WithError(err).Debug("Could not load SLIs on project level")
	} else {
		sliMap, err = addResourceContentToSLIMap(sliMap, sliContent)
		if err != nil {
			log.WithError(err).Debug("Could not add SLIs to SLIMap on project level")
		} else {
			foundLocation = "project,"
		}
	}

	// Step 2: Load Stage Level
	sliContent, err = GetKeptnResourceOnConfigLevel(keptnEvent, DynatraceSLIFilename, ConfigLevelStage)
	if err != nil {
		log.WithError(err).Debug("Could not load SLIs on stage level")
	} else {
		sliMap, err = addResourceContentToSLIMap(sliMap, sliContent)
		if err != nil {
			log.WithError(err).Debug("Could not add SLIs to SLIMap on stage level")
		} else {
			foundLocation = foundLocation + "stage,"
		}
	}

	// Step 3: Load Service Level
	sliContent, err = GetKeptnResourceOnConfigLevel(keptnEvent, DynatraceSLIFilename, ConfigLevelService)
	if err != nil {
		log.WithError(err).Debug("Could not load SLIs on service level")
	} else {
		sliMap, err = addResourceContentToSLIMap(sliMap, sliContent)
		if err != nil {
			log.WithError(err).Debug("Could not add SLIs to SLIMap on service level")
		} else {
			foundLocation = foundLocation + "service"
		}
	}

	// couldnt load any SLIs
	if len(sliMap) == 0 {
		log.WithFields(
			log.Fields{
				"project": keptnEvent.Project,
				"stage":   keptnEvent.Stage,
				"service": keptnEvent.Service,
			}).Info("No custom SLI queries found as no dynatrace/sli.yaml in repo, using defaults")
	} else {
		log.WithFields(
			log.Fields{
				"project":   keptnEvent.Project,
				"stage":     keptnEvent.Stage,
				"service":   keptnEvent.Service,
				"count":     len(sliMap),
				"locations": foundLocation,
			}).Info("Found SLI queries in dynatrace/sli.yaml")
	}

	return sliMap
}

// GetDynatraceConfig loads dynatrace.conf for the current service.
// If none is found, it returns a default configuration.
func GetDynatraceConfig(keptnEvent *BaseKeptnEvent) DynatraceConfigFile {
	dynatraceConfFile := getBaseDynatraceConfig(keptnEvent)
	if dynatraceConfFile.DtCreds == "" {
		dynatraceConfFile.DtCreds = "dynatrace"
	}
	// implementing https://github.com/keptn-contrib/dynatrace-sli-service/issues/90
	dynatraceConfFile.DtCreds = ReplaceKeptnPlaceholders(dynatraceConfFile.DtCreds, keptnEvent)
	return dynatraceConfFile
}

func getBaseDynatraceConfig(keptnEvent *BaseKeptnEvent) DynatraceConfigFile {

	var defaultDynatraceConfigFile = DynatraceConfigFile{
		SpecVersion: "0.1.0",
		DtCreds:     "dynatrace",
		Dashboard:   "",
	}

	yamlString, err := GetKeptnResource(keptnEvent, DynatraceConfigFilename)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"service": keptnEvent.Service,
				"stage":   keptnEvent.Stage,
				"project": keptnEvent.Project,
			}).Debug("Error getting keptn resource")
		return defaultDynatraceConfigFile
	}
	dynatraceConfFile, err := parseDynatraceConfigFile(yamlString)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"yaml":    yamlString,
				"service": keptnEvent.Service,
				"stage":   keptnEvent.Stage,
				"project": keptnEvent.Project,
			}).Error("Error parsing DynatraceConfigFile, using default configuration")
		return defaultDynatraceConfigFile
	}
	return dynatraceConfFile
}

// UploadKeptnResource uploads a file to the Keptn Configuration Service
func UploadKeptnResource(contentToUpload []byte, remoteResourceURI string, keptnEvent *BaseKeptnEvent) error {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if RunLocal || RunLocalTest {
		err := ioutil.WriteFile(remoteResourceURI, contentToUpload, 0644)
		if err != nil {
			return fmt.Errorf("Couldnt write local file %s: %v", remoteResourceURI, err)
		}
		log.WithField("remoteResourceURI", remoteResourceURI).Info("Local file written")
	} else {
		resourceHandler := keptnapi.NewResourceHandler(GetConfigurationServiceURL())

		// lets upload it
		resources := []*keptnmodels.Resource{{ResourceContent: string(contentToUpload), ResourceURI: &remoteResourceURI}}
		_, err := resourceHandler.CreateResources(keptnEvent.Project, keptnEvent.Stage, keptnEvent.Service, resources)
		if err != nil {
			return fmt.Errorf("Couldnt upload remote resource %s: %s", remoteResourceURI, *err.Message)
		}

		log.WithField("remoteResourceURI", remoteResourceURI).Info("Uploaded file")
	}

	return nil
}

/**
 * parses the dynatrace.conf.yaml file that is passed as parameter
 */
func parseDynatraceConfigFile(yamlString string) (DynatraceConfigFile, error) {
	dynatraceConfFile := DynatraceConfigFile{}
	err := yaml.Unmarshal([]byte(yamlString), &dynatraceConfFile)
	return dynatraceConfFile, err
}

/**
 * Pulls the Dynatrace Credentials from the passed secret
 */
func GetDTCredentials(dynatraceSecretName string) (*DTCredentials, error) {
	if dynatraceSecretName == "" {
		return nil, nil
	}

	dtCreds := &DTCredentials{}
	if RunLocal || RunLocalTest {
		// if we RunLocal we take it from the env-variables
		dtCreds.Tenant = os.Getenv("DT_TENANT")
		dtCreds.ApiToken = os.Getenv("DT_API_TOKEN")
	} else {
		kubeAPI, err := GetKubernetesClient()
		if err != nil {
			return nil, fmt.Errorf("error retrieving Dynatrace credentials: could not initialize Kubernetes client: %v", err)
		}
		secret, err := kubeAPI.CoreV1().Secrets(namespace).Get(context.TODO(), dynatraceSecretName, metav1.GetOptions{})

		if err != nil {
			return nil, fmt.Errorf("error retrieving Dynatrace credentials: could not retrieve secret %s.%s: %v", namespace, dynatraceSecretName, err)
		}

		// grabnerandi: remove check on DT_PAAS_TOKEN as it is not relevant for quality-gate-only use case
		if string(secret.Data["DT_TENANT"]) == "" || string(secret.Data["DT_API_TOKEN"]) == "" { //|| string(secret.Data["DT_PAAS_TOKEN"]) == "" {
			return nil, errors.New("invalid or no Dynatrace credentials found. Need DT_TENANT & DT_API_TOKEN stored in secret!")
		}

		dtCreds.Tenant = string(secret.Data["DT_TENANT"])
		dtCreds.ApiToken = string(secret.Data["DT_API_TOKEN"])
	}

	// ensure URL always has http or https in front
	if !(strings.HasPrefix(dtCreds.Tenant, "https://") || strings.HasPrefix(dtCreds.Tenant, "http://")) {
		dtCreds.Tenant = "https://" + dtCreds.Tenant
	}

	return dtCreds, nil
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

// ParseMarkdownConfiguration parses a text that can be used in a Markdown tile to specify global SLO properties
func ParseMarkdownConfiguration(markdown string, slo *keptncommon.ServiceLevelObjectives) {
	markdownSplits := strings.Split(markdown, ";")

	for _, markdownSplitValue := range markdownSplits {
		configValueSplits := strings.Split(markdownSplitValue, "=")
		if len(configValueSplits) != 2 {
			continue
		}

		// lets get configname and value
		configName := strings.ToLower(configValueSplits[0])
		configValue := configValueSplits[1]

		switch configName {
		case "kqg.total.pass":
			slo.TotalScore.Pass = configValue
		case "kqg.total.warning":
			slo.TotalScore.Warning = configValue
		case "kqg.compare.withscore":
			slo.Comparison.IncludeResultWithScore = configValue
			if (configValue == "pass") || (configValue == "pass_or_warn") || (configValue == "all") {
				slo.Comparison.IncludeResultWithScore = configValue
			} else {
				slo.Comparison.IncludeResultWithScore = "pass"
			}
		case "kqg.compare.results":
			noresults, err := strconv.Atoi(configValue)
			if err != nil {
				slo.Comparison.NumberOfComparisonResults = 1
			} else {
				slo.Comparison.NumberOfComparisonResults = noresults
			}
			if slo.Comparison.NumberOfComparisonResults > 1 {
				slo.Comparison.CompareWith = "several_results"
			} else {
				slo.Comparison.CompareWith = "single_result"
			}
		case "kqg.compare.function":
			if (configValue == "avg") || (configValue == "p50") || (configValue == "p90") || (configValue == "p95") {
				slo.Comparison.AggregateFunction = configValue
			} else {
				slo.Comparison.AggregateFunction = "avg"
			}
		}
	}
}

// cleanIndicatorName makes sure we have a valid indicator name by getting rid of special characters
func CleanIndicatorName(indicatorName string) string {
	indicatorName = strings.ReplaceAll(indicatorName, " ", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "/", "_")
	indicatorName = strings.ReplaceAll(indicatorName, "%", "_")

	return indicatorName
}

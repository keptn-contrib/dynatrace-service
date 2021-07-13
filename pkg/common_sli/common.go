package common_sli

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

	log "github.com/sirupsen/logrus"

	keptnmodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var RunLocal = (os.Getenv("ENV") == "local")
var RunLocalTest = (os.Getenv("ENV") == "localtest")

/**
 * Constants for supporting resource files in keptn repo
 */
const DynatraceDashboardFilename = "dynatrace/dashboard.json"
const DynatraceSLIFilename = "dynatrace/sli.yaml"
const KeptnSLOFilename = "slo.yaml"

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

func GetConfigurationServiceURL() string {
	if os.Getenv("CONFIGURATION_SERVICE") != "" {
		return os.Getenv("CONFIGURATION_SERVICE")
	}
	return "configuration-service:8080"
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
func AddResourceContentToSLIMap(SLIs map[string]string, sliFilePath string, sliFileContent string) (map[string]string, error) {

	if sliFilePath != "" {
		localFileContent, err := ioutil.ReadFile(sliFilePath)
		if err != nil {
			log.WithField("sliFilePath", sliFilePath).Info("Could not load file")
			return nil, nil
		}
		log.WithField("sliFilePath", sliFilePath).Info("Loaded LOCAL file")
		sliFileContent = string(localFileContent)
	}

	if sliFileContent != "" {
		sliConfig := keptn.SLIConfig{}
		err := yaml.Unmarshal([]byte(sliFileContent), &sliConfig)
		if err != nil {
			return nil, err
		}

		for key, value := range sliConfig.Indicators {
			SLIs[key] = value
		}
	}
	return SLIs, nil
}

/**
 * getCustomQueries loads custom SLIs from dynatrace/sli.yaml
 * if there is no sli.yaml it will just return an empty map
 */
func GetCustomQueries(keptnEvent *BaseKeptnEvent) (map[string]string, error) {
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
	if err == nil && sliContent != "" {
		sliMap, _ = AddResourceContentToSLIMap(sliMap, "", sliContent)
		foundLocation = "project,"
	}

	// Step 2: Load Stage Level
	sliContent, err = GetKeptnResourceOnConfigLevel(keptnEvent, DynatraceSLIFilename, ConfigLevelStage)
	if err == nil && sliContent != "" {
		sliMap, _ = AddResourceContentToSLIMap(sliMap, "", sliContent)
		foundLocation = foundLocation + "stage,"
	}

	// Step 3: Load Service Level
	sliContent, err = GetKeptnResourceOnConfigLevel(keptnEvent, DynatraceSLIFilename, ConfigLevelService)
	if err == nil && sliContent != "" {
		sliMap, _ = AddResourceContentToSLIMap(sliMap, "", sliContent)
		foundLocation = foundLocation + "service"
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

	return sliMap, nil
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
	if strings.HasPrefix(dtCreds.Tenant, "https://") || strings.HasPrefix(dtCreds.Tenant, "http://") {
		dtCreds.Tenant = dtCreds.Tenant
	} else {
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

// ParsePassAndWarningFromString takes a value such as
// Example 1: Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true
// Example 2: Response time (P95);sli=svc_rt_p95;pass=<+10%,<600
// Example 3: Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false
// can also take a value like "KQG;project=myproject;pass=90%;warning=75%;"
// This will return
// #1: teststep_rt
// #2: []SLOCriteria { Criteria{"<500ms","<+10%"}}
// #3: []SLOCriteria { ["<1000ms","<+20%" }}
// #4: 1
// #5: true
func ParsePassAndWarningFromString(customName string, defaultPass []string, defaultWarning []string) (string, []*keptncommon.SLOCriteria, []*keptncommon.SLOCriteria, int, bool) {
	nameValueSplits := strings.Split(customName, ";")

	// lets initialize it
	sliName := ""
	weight := 1
	keySli := false
	passCriteria := []*keptncommon.SLOCriteria{}
	warnCriteria := []*keptncommon.SLOCriteria{}

	// lets iterate through all name-value pairs which are seprated through ";" to extract keys such as warning, pass, weight, key, sli
	for i := 0; i < len(nameValueSplits); i++ {

		nameValueDividerIndex := strings.Index(nameValueSplits[i], "=")
		if nameValueDividerIndex < 0 {
			continue
		}

		// for each name=value pair we get the name as first part of the string until the first =
		// the value is the after that =
		nameString := strings.ToLower(nameValueSplits[i][:nameValueDividerIndex])
		valueString := nameValueSplits[i][nameValueDividerIndex+1:]
		switch nameString /*nameValueSplit[0]*/ {
		case "sli":
			sliName = valueString
		case "pass":
			passCriteria = append(passCriteria, &keptncommon.SLOCriteria{
				Criteria: strings.Split(valueString, ","),
			})
		case "warning":
			warnCriteria = append(warnCriteria, &keptncommon.SLOCriteria{
				Criteria: strings.Split(valueString, ","),
			})
		case "key":
			keySli, _ = strconv.ParseBool(valueString)
		case "weight":
			weight, _ = strconv.Atoi(valueString)
		}
	}

	// use the defaults if nothing was specified
	if (len(passCriteria) == 0) && (len(defaultPass) > 0) {
		passCriteria = append(passCriteria, &keptncommon.SLOCriteria{
			Criteria: defaultPass,
		})
	}

	if (len(warnCriteria) == 0) && (len(defaultWarning) > 0) {
		warnCriteria = append(warnCriteria, &keptncommon.SLOCriteria{
			Criteria: defaultWarning,
		})
	}

	// if we have no criteria for warn or pass we just return nil
	if len(passCriteria) == 0 {
		passCriteria = nil
	}
	if len(warnCriteria) == 0 {
		warnCriteria = nil
	}

	return sliName, passCriteria, warnCriteria, weight, keySli
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

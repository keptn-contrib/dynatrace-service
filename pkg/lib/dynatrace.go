package lib

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	keptnutils "github.com/keptn/go-utils/pkg/api/utils"
)

const DefaultOperatorVersion = "v0.8.0"
const sliResourceURI = "dynatrace/sli.yaml"
const Throughput = "throughput"
const ErrorRate = "error_rate"
const ResponseTimeP50 = "response_time_p50"
const ResponseTimeP90 = "response_time_p90"
const ResponseTimeP95 = "response_time_p95"

type criteriaObject struct {
	Operator        string
	Value           float64
	CheckPercentage bool
	IsComparison    bool
	CheckIncrease   bool
}

type DTAPIListResponse struct {
	Values []Values `json:"values"`
}
type Values struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DynatraceHelper struct {
	DynatraceCreds     *credentials.DTCredentials
	Logger             keptncommon.LoggerInterface
	OperatorTag        string
	KeptnHandler       *keptnv2.Keptn
	KeptnBridge        string
	configuredEntities *ConfiguredEntities
}

// ConfigResult godoc
type ConfigResult struct {
	Name    string
	Success bool
	Message string
}

// ConfiguredEntities contains information about the entities configures in Dynatrace
type ConfiguredEntities struct {
	TaggingRulesEnabled         bool
	TaggingRules                []ConfigResult
	ProblemNotificationsEnabled bool
	ProblemNotifications        ConfigResult
	ManagementZonesEnabled      bool
	ManagementZones             []ConfigResult
	DashboardEnabled            bool
	Dashboard                   ConfigResult
	MetricEventsEnabled         bool
	MetricEvents                []ConfigResult
}

// NewDynatraceHelper creates a new DynatraceHelper
func NewDynatraceHelper(keptnHandler *keptnv2.Keptn, dynatraceCreds *credentials.DTCredentials, logger keptncommon.LoggerInterface) *DynatraceHelper {
	return &DynatraceHelper{
		DynatraceCreds: dynatraceCreds,
		KeptnHandler:   keptnHandler,
		Logger:         logger,
	}
}

// ConfigureMonitoring configures Dynatrace for a Keptn project
func (dt *DynatraceHelper) ConfigureMonitoring(project string, shipyard *keptnv2.Shipyard) (*ConfiguredEntities, error) {

	dt.configuredEntities = &ConfiguredEntities{
		TaggingRulesEnabled:         IsTaggingRulesGenerationEnabled(),
		TaggingRules:                []ConfigResult{},
		ProblemNotificationsEnabled: IsProblemNotificationsGenerationEnabled(),
		ProblemNotifications:        ConfigResult{},
		ManagementZonesEnabled:      IsManagementZonesGenerationEnabled(),
		ManagementZones:             []ConfigResult{},
		DashboardEnabled:            IsDashboardsGenerationEnabled(),
		Dashboard:                   ConfigResult{},
		MetricEventsEnabled:         IsMetricEventsGenerationEnabled(),
		MetricEvents:                []ConfigResult{},
	}
	dt.EnsureDTTaggingRulesAreSetUp()

	dt.EnsureProblemNotificationsAreSetUp()

	if project != "" && shipyard != nil {
		dt.CreateManagementZones(project, *shipyard)

		configHandler := keptnutils.NewServiceHandler("shipyard-controller:8080")
		dt.CreateDashboard(project, *shipyard)

		// try to create metric events - if one fails, don't fail the whole setup
		for _, stage := range shipyard.Spec.Stages {
			if shouldCreateMetricEvents(stage) {
				services, err := configHandler.GetAllServices(project, stage.Name)
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve services of project %s: %v", project, err.Error())
				}
				for _, service := range services {
					dt.CreateMetricEvents(project, stage.Name, service.ServiceName)
				}
			}
		}
	}
	return dt.configuredEntities, nil
}

// shouldCreateMetricEvents checks if a task sequence with the name 'remediation' is available - this would be the equivalent of remediation_strategy: automated of Keptn < 0.8.x
func shouldCreateMetricEvents(stage keptnv2.Stage) bool {
	for _, taskSequence := range stage.Sequences {
		if taskSequence.Name == "remediation" {
			return true
		}
	}
	return false
}

/**
 * if dtCredsSecretName is passed and it is not dynatrace (=default) then we try to pull the secret based on that name and is it for this API Call
 */
func (dt *DynatraceHelper) sendDynatraceAPIRequest(apiPath string, method string, body []byte) (string, error) {

	if common.RunLocal || common.RunLocalTest {
		dt.Logger.Info("Dynatrace.sendDynatraceAPIRequest(RUNLOCAL) - not sending event to " +
			dt.DynatraceCreds.Tenant + "). Here is the payload: " + string(body))
		return "", nil
	}

	req, err := dt.createRequest(apiPath, method, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	client, err := dt.createClient(req)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}

	response, err := dt.doRequest(client, req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %v", err)
	}

	return response, nil
}

// creates http request for api call with appropriate headers including authorization
func (dt *DynatraceHelper) createRequest(apiPath string, method string, body []byte) (*http.Request, error) {
	var url string
	if !strings.HasPrefix(dt.DynatraceCreds.Tenant, "http://") && !strings.HasPrefix(dt.DynatraceCreds.Tenant, "https://") {
		url = "https://" + dt.DynatraceCreds.Tenant + apiPath
	} else {
		url = dt.DynatraceCreds.Tenant + apiPath
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Token "+dt.DynatraceCreds.ApiToken)
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+os.Getenv("version"))

	return req, nil
}

// creates http client with proxy and TLS configuration
func (dt *DynatraceHelper) createClient(req *http.Request) (*http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !IsHttpSSLVerificationEnabled()},
		Proxy:           http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: tr}

	return client, nil
}

// performs the request and reads the response
func (dt *DynatraceHelper) doRequest(client *http.Client, req *http.Request) (string, error) {
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send Dynatrace API request: %v", err)
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return string(responseBody), fmt.Errorf("api request failed with status %s and response %s", resp.Status, string(responseBody))
	}

	return string(responseBody), nil
}

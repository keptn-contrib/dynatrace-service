package lib

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

const DefaultOperatorVersion = "v0.8.0"
const SliResourceURI = "dynatrace/sli.yaml"
const Throughput = "throughput"
const ErrorRate = "error_rate"
const ResponseTimeP50 = "response_time_p50"
const ResponseTimeP90 = "response_time_p90"
const ResponseTimeP95 = "response_time_p95"

type CriteriaObject struct {
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
func NewDynatraceHelper(keptnHandler *keptnv2.Keptn, dynatraceCreds *credentials.DTCredentials) *DynatraceHelper {
	return &DynatraceHelper{
		DynatraceCreds: dynatraceCreds,
		KeptnHandler:   keptnHandler,
	}
}

// SendDynatraceAPIRequest makes an Dynatrace API request and returns the response
func (dt *DynatraceHelper) SendDynatraceAPIRequest(apiPath string, method string, body []byte) (string, error) {

	if common.RunLocal || common.RunLocalTest {
		log.WithFields(
			log.Fields{
				"tenant": dt.DynatraceCreds.Tenant,
				"body":   string(body),
			}).Info("Dynatrace.sendDynatraceAPIRequest(RUNLOCAL) - not sending event to tenant")
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

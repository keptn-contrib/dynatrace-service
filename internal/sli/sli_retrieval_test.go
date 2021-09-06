package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"net/http"
	"net/http/httptest"

	_ "github.com/keptn/go-utils/pkg/lib"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"golang.org/x/net/context"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const QUALITYGATE_DASHBOARD_ID = "12345678-1111-4444-8888-123456789012"
const QUALITYGATE_PROJECT = "qualitygate"
const QUALTIYGATE_SERVICE = "evalservice"
const QUALITYGATE_STAGE = "qualitystage"

// Mocking Http Responses
// testingDynatraceHTTPClient builds a test client with a httptest server that responds to specific Dynatrace REST API Calls
func testingDynatraceHTTPClient() (*http.Client, string, func()) {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// we handle these if the URLs are a full match
		completeUrlMatchToResponseFileMap := map[string]string{
			"/api/config/v1/dashboards":                                      "./testfiles/test_get_dashboards.json",
			"/api/config/v1/dashboards/12345678-1111-4444-8888-123456789012": "./testfiles/test_get_dashboards_id.json",
			"/api/v2/metrics/builtin:tech.generic.processCount":              "./testfiles/test_get_metrics_processcount.json",
			"/api/v2/metrics/builtin:service.response.time":                  "./testfiles/test_get_metrics_svcresponsetime.json",
			"/api/v2/metrics/builtin:tech.generic.mem.workingSetSize":        "./testfiles/test_get_metrics_workingsetsize.json",
			"/api/v2/metrics/builtin:tech.generic.cpu.usage":                 "./testfiles/test_get_metrics_cpuusage.json",
			"/api/v2/metrics/builtin:service.errors.server.rate":             "./testfiles/test_get_metrics_errorrate.json",
			"/api/v2/metrics/builtin:service.requestCount.total":             "./testfiles/test_get_metrics_requestcount.json",
			"/api/v2/metrics/builtin:host.cpu.usage":                         "./testfiles/test_get_metrics_hostcpuusage.json",
			"/api/v2/metrics/builtin:host.mem.usage":                         "./testfiles/test_get_metrics_hostmemusage.json",
			"/api/v2/metrics/builtin:host.disk.queueLength":                  "./testfiles/test_get_metrics_hostdiskqueue.json",
			"/api/v2/metrics/builtin:service.nonDbChildCallCount":            "./testfiles/test_get_metrics_nondbcallcount.json",
			"/api/v2/metrics/jmeter.usermetrics.transaction.meantime":        "./testfiles/test_get_metrics_jmeter_usermetrics_transaction_meantime.json",
		}

		log.Println("Mock for: " + r.URL.Path)

		// we handle these if the URL "starts with"
		startsWithUrlToResponseFileMap := map[string]string{
			"/api/v2/metrics/query":    "./testfiles/test_get_metrics_query.json",
			"/api/v2/slo":              "./testfiles/test_get_slo_id.json",
			"/api/v2/problems":         "./testfiles/test_get_problems.json",
			"/api/v2/securityProblems": "./testfiles/test_get_securityproblems.json",
		}

		for url, file := range completeUrlMatchToResponseFileMap {
			if strings.Compare(url, r.URL.Path) == 0 {
				log.Println("Found Mock: " + url + " --> " + file)
				localFileContent, err := ioutil.ReadFile(file)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, "couldnt load local test file "+file)
					log.Println("couldnt load local test file " + file)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write(localFileContent)
				return
			}
		}

		for url, file := range startsWithUrlToResponseFileMap {
			if strings.Index(r.URL.Path, url) == 0 {
				log.Println("Found Mock: " + url + " --> " + file)

				localFileContent, err := ioutil.ReadFile(file)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, "couldnt load local test file "+file)
					log.Println("couldnt load local test file " + file)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write(localFileContent)
				return
			}
		}

		// if nothing matches we have a bad URL
		w.WriteHeader(http.StatusBadRequest)
	})

	server := httptest.NewTLSServer(handler)

	cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
	if err != nil {
		log.Fatal(err)
	}

	certpool := x509.NewCertPool()
	certpool.AddCert(cert)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, server.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				RootCAs: certpool,
			},
		},
	}

	return client, server.URL, server.Close
}

/**
 * Creates a new Keptn Event
 */
func testingGetKeptnEvent(project string, stage string, service string, deployment string, test string) GetSLITriggeredAdapterInterface {
	keptnEvent := &BaseKeptnEvent{}
	keptnEvent.Project = project
	keptnEvent.Stage = stage
	keptnEvent.Service = service
	keptnEvent.DeploymentStrategy = deployment
	keptnEvent.TestStrategy = test

	return keptnEvent
}

/**
 * This function will create a new HTTP Server for handling Dynatrace REST Calls.
 * It returns the Dynatrace Retrieval as well as the httpClient, mocked server url and the teardown method
 * ATTENTION: When using this method you have to call the "teardown" method that is returned in the last parameter
 */
func testingGetDynatraceHandler(keptnEvent GetSLITriggeredAdapterInterface) (*Retrieval, *http.Client, string, func()) {
	httpClient, url, teardown := testingDynatraceHTTPClient()

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	dh := NewRetrieval(
		keptnEvent,
		dynatrace.NewClientWithHTTP(dtCredentials, httpClient),
		KeptnClientMock{})

	return dh, httpClient, url, teardown
}

func TestExecuteDynatraceREST(t *testing.T) {
	keptnEvent := testingGetKeptnEvent("sockshop", "dev", "carts", "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	body, err := dh.dtClient.Get("/api/config/v1/dashboards")

	if body == nil {
		t.Errorf("No body returned by Dynatrace REST")
	}

	if err != nil {
		t.Errorf("%+v\n", err)
	}
}

func TestExecuteDynatraceRESTBadRequest(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	_, err := dh.dtClient.Get("/BADAPI")

	// TODO 2021-08-31: check for Dynatrace API status
	if err == nil {
		t.Errorf("Dynatrace REST not returning http 400")
	}
}

func TestFindDynatraceDashboardSuccess(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	dashboardID, err := dh.findDynatraceDashboard(keptnEvent)

	if err != nil {
		t.Error(err)
	}

	if dashboardID != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("findDynatraceDashboard not finding quality gate dashboard")
	}
}

func TestFindDynatraceDashboardNoneExistingDashboard(t *testing.T) {
	keptnEvent := testingGetKeptnEvent("BAD_PROJECT", QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	dashboardID, err := dh.findDynatraceDashboard(keptnEvent)

	if err != nil {
		t.Error(err)
	}

	if dashboardID != "" {
		t.Errorf("findDynatraceDashboard found a dashboard that should not have been found: " + dashboardID)
	}
}

func TestLoadDynatraceDashboardWithQUERY(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.loadDynatraceDashboard(keptnEvent, common.DynatraceConfigDashboardQUERY)

	if dashboardJSON == nil {
		t.Errorf("Didnt query dashboard for quality gate project even though it shoudl exist: " + dashboard)
	}

	if dashboard != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("Didnt query the dashboard that matches the project/stage/service names: " + dashboard)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoadDynatraceDashboardWithID(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.loadDynatraceDashboard(keptnEvent, QUALITYGATE_DASHBOARD_ID)

	if dashboardJSON == nil {
		t.Errorf("Didnt query dashboard for quality gate project even though it should exist by ID")
	}

	if dashboard != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("loadDynatraceDashboard should return the passed in dashboard id")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoadDynatraceDashboardWithEmptyDashboard(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.loadDynatraceDashboard(keptnEvent, "")

	if dashboardJSON != nil {
		t.Errorf("No dashboard should be loaded if no dashboard is passed")
	}

	if dashboard != "" {
		t.Errorf("dashboard should be empty as by default we dont load a dashboard")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestGetEntitySelectorFromEntityFilter(t *testing.T) {
	var filtersPerEntityType = map[string]map[string][]string{
		"SERVICE": {
			"SPECIFIC_ENTITIES": {"SERVICE-086C46F600BA1DC6"},
			"AUTO_TAGS":         {"keptn_deployment:primary"},
		},
	}
	entityTileFilter := getEntitySelectorFromEntityFilter(filtersPerEntityType, "SERVICE")

	if strings.Compare(entityTileFilter, ",entityId(\"SERVICE-086C46F600BA1DC6\"),tag(\"keptn_deployment:primary\")") != 0 {
		t.Errorf("getEntitySelectorFromEntityFilter wrong. Returned: " + entityTileFilter)
	}
}

func TestQueryDynatraceDashboardForSLIs(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	result, err := dh.QueryDynatraceDashboardForSLIs(keptnEvent, common.DynatraceConfigDashboardQUERY, startTime, endTime)

	if result == nil {
		t.Fatalf("No result returned")
	}

	if result.dashboardLink == nil {
		t.Errorf("No dashboard link label generated")
	}

	if result.dashboard == nil {
		t.Errorf("No Dashboard JSON returned")
	}

	const expectedSLOs = 14

	// validate the SLIs - there should be 9 SLIs coming back
	if result.sli != nil {
		if len(result.sli.Indicators) != expectedSLOs {
			t.Errorf("Excepted %d SLIs to come back but got %d", expectedSLOs, len(result.sli.Indicators))
		}
	} else {
		t.Errorf("No SLI returned")
	}

	// validate the SLOs
	if result.slo != nil {
		if len(result.slo.Objectives) != expectedSLOs {
			t.Errorf("Excepted %d SLOs to come back but got %d", expectedSLOs, len(result.slo.Objectives))
		}
		if result.slo.TotalScore.Pass != "90%" || result.slo.TotalScore.Warning != "70%" {
			t.Errorf("Total Warning and Pass Scores not as expected. Got %s (pass) and %s (warning)", result.slo.TotalScore.Pass, result.slo.TotalScore.Warning)
		}
		if result.slo.Comparison.CompareWith != "single_result" ||
			result.slo.Comparison.IncludeResultWithScore != "pass" ||
			result.slo.Comparison.NumberOfComparisonResults != 1 ||
			result.slo.Comparison.AggregateFunction != "avg" {
			t.Errorf(
				"Incorrect Comparisons: %s, %s, %d, %s",
				result.slo.Comparison.CompareWith,
				result.slo.Comparison.IncludeResultWithScore,
				result.slo.Comparison.NumberOfComparisonResults,
				result.slo.Comparison.AggregateFunction)
		}
	} else {
		t.Errorf("No SLO return")
	}

	// validate the SLI Results
	if result.sliResults != nil {
		if len(result.sliResults) != expectedSLOs {
			t.Errorf("Excepted %d SLI Results to come back but got %d", expectedSLOs, len(result.sliResults))
		}
	} else {
		t.Errorf("No SLI Results returned")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestExecuteGetDynatraceSLO(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	sloID := "524ca177-849b-3e8c-8175-42b93fbc33c5"
	sloResult, err := dynatrace.NewSLOClient(dh.dtClient).Get(sloID, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if sloResult == nil {
		t.Errorf("No SLO Result returned for " + sloID)
	}

	if sloResult.EvaluatedPercentage != 95.66405076939219 {
		t.Error("Not returning expected value for SLO")
	}
}

func TestGetSLIValueWithSLOPrefix(t *testing.T) {

	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	customQueries := make(map[string]string)
	customQueries["RT_faster_500ms"] = "SLO;524ca177-849b-3e8c-8175-42b93fbc33c5"

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()

	_, err := dh.GetSLIValue("RT_faster_500ms", startTime, endTime, keptn.NewCustomQueries(customQueries))

	if err != nil {
		t.Error(err)
	}
}

func TestExecuteGetDynatraceProblems(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(open)"
	problemResult, err := dynatrace.NewProblemsV2Client(dh.dtClient).GetByQuery(problemQuery, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if problemResult == nil {
		t.Fatal("No Problem Result returned for " + problemQuery)
	}

	if problemResult.TotalCount != 1 {
		t.Error("Not returning expected value for Problem Query")
	}
}

func TestExecuteGetDynatraceSecurityProblems(t *testing.T) {
	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(OPEN)"

	// TODO 2021-09-02: fix dependency on sli/Retrieval below!
	problemResult, err := dynatrace.NewSecurityProblemsClient(dh.dtClient).GetByQuery(problemQuery, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if problemResult == nil {
		t.Fatal("No Problem Result returned for " + problemQuery)
	}

	if problemResult.TotalCount != 0 {
		t.Error("Not returning expected value for Problem Query")
	}
}

func TestGetSLIValueWithPV2Prefix(t *testing.T) {

	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	customQueries := make(map[string]string)
	customQueries["problems"] = "PV2;problemEntity=status(open)"

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()

	_, err := dh.GetSLIValue("problems", startTime, endTime, keptn.NewCustomQueries(customQueries))

	if err != nil {
		t.Error(err)
	}
}

func TestGetSLIValueWithSECPV2Prefix(t *testing.T) {

	keptnEvent := testingGetKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE, "", "")
	dh, _, _, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	customQueries := make(map[string]string)
	customQueries["security_problems"] = "SECPV2;problemEntity=status(open)"

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()

	_, err := dh.GetSLIValue("security_problems", startTime, endTime, keptn.NewCustomQueries(customQueries))

	if err != nil {
		t.Error(err)
	}
}

func TestCreateNewDynatraceHandler(t *testing.T) {
	keptnEvent := testingGetKeptnEvent("sockshop", "dev", "carts", "direct", "")
	dh, _, url, teardown := testingGetDynatraceHandler(keptnEvent)
	defer teardown()

	if dh.dtClient.Credentials().Tenant != url {
		t.Errorf("dh.client.DynatraceCreds.Tenant=%s; want %s", dh.dtClient.Credentials().Tenant, url)
	}

	if dh.KeptnEvent.GetProject() != "sockshop" {
		t.Errorf("dh.Project=%s; want sockshop", dh.KeptnEvent.GetProject())
	}

	if dh.KeptnEvent.GetStage() != "dev" {
		t.Errorf("dh.Stage=%s; want dev", dh.KeptnEvent.GetStage())
	}

	if dh.KeptnEvent.GetService() != "carts" {
		t.Errorf("dh.Service=%s; want carts", dh.KeptnEvent.GetService())
	}
	if dh.KeptnEvent.GetDeploymentStrategy() != "direct" {
		t.Errorf("dh.Deployment=%s; want direct", dh.KeptnEvent.GetDeploymentStrategy())
	}
}

func TestTimestampToString(t *testing.T) {
	dt := time.Now()

	got := common.TimestampToString(dt)

	expected := strconv.FormatInt(dt.Unix()*1000, 10)

	if got != expected {
		t.Errorf("timestampToString() returned %s, expected %s", got, expected)
	}
}

// tests the parseUnixTimestamp with invalid params
func TestParseInvalidUnixTimestamp(t *testing.T) {
	_, err := common.ParseUnixTimestamp("")

	if err == nil {
		t.Errorf("parseUnixTimestamp(\"\") did not return an error")
	}
}

// tests the parseUnixTimestamp with valid params
func TestParseValidUnixTimestamp(t *testing.T) {
	got, err := common.ParseUnixTimestamp("2019-10-24T15:44:27.152330783Z")

	if err != nil {
		t.Fatalf("parseUnixTimestamp(\"2019-10-24T15:44:27.152330783Z\") returned error %s", err.Error())
	}

	if got.Year() != 2019 {
		t.Errorf("parseUnixTimestamp() returned year %d, expected 2019", got.Year())
	}

	if got.Month() != 10 {
		t.Errorf("parseUnixTimestamp() returned month %d, expected 10", got.Month())
	}

	if got.Day() != 24 {
		t.Errorf("parseUnixTimestamp() returned day %d, expected 24", got.Day())
	}

	if got.Hour() != 15 {
		t.Errorf("parseUnixTimestamp() returned hour %d, expected 15", got.Hour())
	}

	if got.Minute() != 44 {
		t.Errorf("parseUnixTimestamp() returned minute %d, expected 44", got.Minute())
	}
}

func TestScaleData(t *testing.T) {
	if scaleData("", "MicroSecond", 1000000.0) != 1000.0 {
		t.Errorf("scaleData incorrectly scales MicroSecond")
	}
	if scaleData("", "Byte", 1024.0) != 1.0 {
		t.Errorf("scaleData incorrectly scales Bytes")
	}
	if scaleData("builtin:service.response.time", "", 1000000.0) != 1000.0 {
		t.Errorf("scaleData incorrectly scales builtin:service.response.time")
	}
}

func TestParsePassAndWarningFromString(t *testing.T) {
	type args struct {
		customName string
	}
	tests := []struct {
		name string
		args args
		want keptnapi.SLO
	}{
		{
			name: "simple test",
			args: args{
				customName: "Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true",
			},
			want: keptnapi.SLO{
				SLI:     "teststep_rt",
				Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{"<500ms", "<+10%"}}},
				Warning: []*keptnapi.SLOCriteria{{Criteria: []string{"<1000ms", "<+20%"}}},
				Weight:  1,
				KeySLI:  true,
			},
		},
		{
			name: "test with = in pass/warn expression",
			args: args{
				customName: "Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false",
			},
			want: keptnapi.SLO{
				SLI:     "host_disk_queue",
				Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
				Warning: []*keptnapi.SLOCriteria{{Criteria: []string{"<1"}}},
				Weight:  1,
				KeySLI:  false,
			},
		},
		{
			name: "test weight",
			args: args{
				customName: "Host CPU %;sli=host_cpu;pass=<20;warning=<50;key=false;weight=2",
			},
			want: keptnapi.SLO{
				SLI:     "host_cpu",
				Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{"<20"}}},
				Warning: []*keptnapi.SLOCriteria{{Criteria: []string{"<50"}}},
				Weight:  2,
				KeySLI:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.ParsePassAndWarningWithoutDefaultsFrom(tt.args.customName)
			if got.SLI != tt.want.SLI {
				t.Errorf("ParsePassAndWarningFromString() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got.Pass, tt.want.Pass) {
				t.Errorf("ParsePassAndWarningFromString() Pass = %v, want %v", got.Pass, tt.want.Pass)
			}
			if !reflect.DeepEqual(got.Warning, tt.want.Warning) {
				t.Errorf("ParsePassAndWarningFromString() Warning = %v, want %v", got.Warning, tt.want.Warning)
			}
			if got.Weight != tt.want.Weight {
				t.Errorf("ParsePassAndWarningFromString() Weight = %v, want %v", got.Weight, tt.want.Weight)
			}
			if got.KeySLI != tt.want.KeySLI {
				t.Errorf("ParsePassAndWarningFromString() KeySLI = %v, want %v", got.KeySLI, tt.want.KeySLI)
			}
		})
	}
}

func TestParseMarkdownConfiguration(t *testing.T) {

	dashboardSLO1 := &keptnapi.ServiceLevelObjectives{
		Objectives: []*keptnapi.SLO{},
		TotalScore: &keptnapi.SLOScore{Pass: "", Warning: ""},
		Comparison: &keptnapi.SLOComparison{CompareWith: "", IncludeResultWithScore: "", NumberOfComparisonResults: 0, AggregateFunction: ""},
	}

	// first run - single result
	common.ParseMarkdownConfiguration("KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg", dashboardSLO1)

	if dashboardSLO1.TotalScore.Pass != "90%" {
		t.Errorf("Total Pass not 90% - is " + dashboardSLO1.TotalScore.Pass)
	}
	if dashboardSLO1.TotalScore.Warning != "70%" {
		t.Errorf("Total Pass not 70% - is " + dashboardSLO1.TotalScore.Warning)
	}
	if dashboardSLO1.Comparison.CompareWith != "single_result" {
		t.Errorf("CompareWith not single_result - is " + dashboardSLO1.Comparison.CompareWith)
	}
	if dashboardSLO1.Comparison.IncludeResultWithScore != "pass" {
		t.Errorf("IncludeResultWithScore not pass - is " + dashboardSLO1.Comparison.IncludeResultWithScore)
	}
	if dashboardSLO1.Comparison.NumberOfComparisonResults != 1 {
		t.Errorf("NumberOfComparisonResults not 1 - but its %d ", dashboardSLO1.Comparison.NumberOfComparisonResults)
	}
	if dashboardSLO1.Comparison.AggregateFunction != "avg" {
		t.Errorf("AggregateFunction not avg - is " + dashboardSLO1.Comparison.AggregateFunction)
	}

	// second run - multiple results
	dashboardSLO2 := &keptnapi.ServiceLevelObjectives{
		Objectives: []*keptnapi.SLO{},
		TotalScore: &keptnapi.SLOScore{Pass: "", Warning: ""},
		Comparison: &keptnapi.SLOComparison{CompareWith: "", IncludeResultWithScore: "", NumberOfComparisonResults: 0, AggregateFunction: ""},
	}
	common.ParseMarkdownConfiguration("KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p50", dashboardSLO2)

	if dashboardSLO2.TotalScore.Pass != "50%" {
		t.Errorf("Total Pass not 50% - is " + dashboardSLO2.TotalScore.Pass)
	}
	if dashboardSLO2.TotalScore.Warning != "40%" {
		t.Errorf("Total Pass not 40% - is " + dashboardSLO2.TotalScore.Warning)
	}
	if dashboardSLO2.Comparison.CompareWith != "several_results" {
		t.Errorf("CompareWith not several_results - is " + dashboardSLO2.Comparison.CompareWith)
	}
	if dashboardSLO2.Comparison.IncludeResultWithScore != "pass" {
		t.Errorf("IncludeResultWithScore not pass - is " + dashboardSLO2.Comparison.IncludeResultWithScore)
	}
	if dashboardSLO2.Comparison.NumberOfComparisonResults != 3 {
		t.Errorf("NumberOfComparisonResults not 3 - but its %d ", dashboardSLO2.Comparison.NumberOfComparisonResults)
	}
	if dashboardSLO2.Comparison.AggregateFunction != "p50" {
		t.Errorf("AggregateFunction not p50 - is " + dashboardSLO2.Comparison.AggregateFunction)
	}

	// third run - invalid functionresults
	dashboardSLO3 := &keptnapi.ServiceLevelObjectives{
		Objectives: []*keptnapi.SLO{},
		TotalScore: &keptnapi.SLOScore{Pass: "", Warning: ""},
		Comparison: &keptnapi.SLOComparison{CompareWith: "", IncludeResultWithScore: "", NumberOfComparisonResults: 0, AggregateFunction: ""},
	}
	common.ParseMarkdownConfiguration("KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=INVALID", dashboardSLO3)

	if dashboardSLO3.TotalScore.Pass != "50%" {
		t.Errorf("Total Pass not 50% - is " + dashboardSLO3.TotalScore.Pass)
	}
	if dashboardSLO3.TotalScore.Warning != "40%" {
		t.Errorf("Total Pass not 40% - is " + dashboardSLO3.TotalScore.Warning)
	}
	if dashboardSLO3.Comparison.CompareWith != "several_results" {
		t.Errorf("CompareWith not several_results - is " + dashboardSLO3.Comparison.CompareWith)
	}
	if dashboardSLO3.Comparison.IncludeResultWithScore != "pass" {
		t.Errorf("IncludeResultWithScore not pass - is " + dashboardSLO3.Comparison.IncludeResultWithScore)
	}
	if dashboardSLO3.Comparison.NumberOfComparisonResults != 3 {
		t.Errorf("NumberOfComparisonResults not 3 - but its %d ", dashboardSLO3.Comparison.NumberOfComparisonResults)
	}
	if dashboardSLO3.Comparison.AggregateFunction != "avg" {
		t.Errorf("AggregateFunction not avg - is " + dashboardSLO3.Comparison.AggregateFunction)
	}

}

func TestIsValidUUID(t *testing.T) {
	testConfigs := []struct {
		uuid string
		want bool
	}{
		// reproduce issue with "|"
		{
			"311f4aa7-5257-41d7-|bd1-70420500e1c8",
			false,
		},
		// valid UUID v4, variant 1
		{
			"311f4aa7-5257-41d7-abd1-70420500e1c8",
			true,
		},
		// NIL UUID is not valid
		{
			"00000000-0000-0000-0000-000000000000",
			false,
		},
	}
	for _, config := range testConfigs {
		got := isValidUUID(config.uuid)
		if got != config.want {
			t.Errorf("uuid: %s, result should have been: %v, but got: %v", config.uuid, config.want, got)
		}
	}
}
